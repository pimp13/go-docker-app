package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"test/internal/config"
	"test/internal/dto"
	"test/internal/model"
	"test/pkg/jwtutil"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already exists")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrAccountDisabled    = errors.New("account is not active")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUnauthenticated    = errors.New("authentication required")
	ErrWrongPassword      = errors.New("current password is incorrect")
	ErrPasswordUnchanged  = errors.New("new password must be different from the current password")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidPassword    = errors.New("password must be between 8 and 72 characters")
	ErrDefaultRoleMissing = errors.New("default user role is not configured")
)

// dummyPasswordHash is a valid bcrypt hash used only so a missing account
// spends about the same time as a real password check.
var dummyPasswordHash = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (dto.AuthUser, error)
	Login(ctx context.Context, req dto.LoginRequest) (dto.Session, error)
	Authenticate(ctx context.Context, rawToken string) (dto.AuthUser, error)
	ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) (dto.Session, error)
}

type authService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) AuthService {
	return &authService{db: db, cfg: cfg}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (dto.AuthUser, error) {
	if err := normalizeRegister(&req); err != nil {
		return dto.AuthUser{}, err
	}

	defaultRole, err := s.findRoleBySlug(ctx, model.RoleSlugUser)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AuthUser{}, ErrDefaultRoleMissing
		}
		return dto.AuthUser{}, err
	}

	user := model.User{
		ID:        uuid.New(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		Phone:     req.Phone,
		RoleID:    defaultRole.ID,
		Role:      defaultRole,
		Status:    model.UserStatusActive,
		Meta:      datatypes.JSONMap{},
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrEmailTaken
		}

		if req.Username != "" {
			count = 0
			if err := tx.Model(&model.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return ErrUsernameTaken
			}
		}

		create := tx
		if req.Username == "" {
			create = tx.Omit("Username")
		}
		if err := create.Omit("Role").Create(&user).Error; err != nil {
			return mapUniqueViolation(err)
		}
		return nil
	})
	if err != nil {
		return dto.AuthUser{}, err
	}

	return toAuthUser(user), nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (dto.Session, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return dto.Session{}, ErrInvalidCredentials
	}

	user, err := s.findByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(req.Password))
			return dto.Session{}, ErrInvalidCredentials
		}
		return dto.Session{}, err
	}

	if !user.CheckPassword(req.Password) {
		return dto.Session{}, ErrInvalidCredentials
	}
	if user.Status != model.UserStatusActive {
		return dto.Session{}, ErrAccountDisabled
	}

	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).Update("last_login", now).Error; err != nil {
		return dto.Session{}, err
	}
	user.LastLogin = &now

	return s.issue(user)
}

func (s *authService) Authenticate(ctx context.Context, rawToken string) (dto.AuthUser, error) {
	claims, err := jwtutil.Parse(s.cfg.Auth.JWTSecret, s.cfg.Auth.JWTIssuer, rawToken)
	if err != nil {
		if errors.Is(err, jwtutil.ErrExpired) {
			return dto.AuthUser{}, ErrTokenExpired
		}
		return dto.AuthUser{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return dto.AuthUser{}, ErrInvalidToken
	}

	user, err := s.findByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AuthUser{}, ErrInvalidToken
		}
		return dto.AuthUser{}, err
	}
	if !fingerprintMatches(user.Password, claims.PasswordFingerprint) {
		return dto.AuthUser{}, ErrInvalidToken
	}
	if user.Status != model.UserStatusActive {
		return dto.AuthUser{}, ErrAccountDisabled
	}

	return toAuthUser(user), nil
}

func (s *authService) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) (dto.Session, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return dto.Session{}, ErrInvalidToken
	}
	if !validPassword(req.NewPassword) {
		return dto.Session{}, ErrInvalidPassword
	}
	if req.CurrentPassword == req.NewPassword {
		return dto.Session{}, ErrPasswordUnchanged
	}

	user, err := s.findByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.Session{}, ErrInvalidToken
		}
		return dto.Session{}, err
	}
	if user.Status != model.UserStatusActive {
		return dto.Session{}, ErrAccountDisabled
	}
	if !user.CheckPassword(req.CurrentPassword) {
		return dto.Session{}, ErrWrongPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return dto.Session{}, err
	}
	user.Password = string(hashed)

	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).Update("password", user.Password).Error; err != nil {
		return dto.Session{}, err
	}

	return s.issue(user)
}

func (s *authService) issue(user model.User) (dto.Session, error) {
	signed, expiresAt, err := jwtutil.Issue(jwtutil.IssueParams{
		Secret:              s.cfg.Auth.JWTSecret,
		Issuer:              s.cfg.Auth.JWTIssuer,
		Subject:             user.ID.String(),
		Role:                user.Role.Slug,
		PasswordFingerprint: passwordFingerprint(user.Password),
		TTL:                 s.cfg.Auth.TokenTTL,
	})
	if err != nil {
		return dto.Session{}, err
	}

	return dto.Session{
		Token:     signed,
		ExpiresAt: expiresAt,
		User:      toAuthUser(user),
	}, nil
}

func (s *authService) findByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("email = ?", email).
		Take(&user).Error
	return user, err
}

func (s *authService) findByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("id = ?", id).
		Take(&user).Error
	return user, err
}

func (s *authService) findRoleBySlug(ctx context.Context, slug string) (model.Role, error) {
	var role model.Role
	err := s.db.WithContext(ctx).
		Preload("Permissions").
		Where("slug = ?", slug).
		Take(&role).Error
	return role, err
}

func normalizeRegister(req *dto.RegisterRequest) error {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	req.Phone = strings.TrimSpace(req.Phone)

	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return ErrInvalidInput
	}
	if !validPassword(req.Password) {
		return ErrInvalidPassword
	}
	return nil
}

func validPassword(password string) bool {
	n := len(password)
	return n >= 8 && n <= 72
}

func toAuthUser(user model.User) dto.AuthUser {
	return dto.AuthUser{
		ID:          user.ID.String(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Username:    user.Username,
		Phone:       user.Phone,
		Role:        user.Role.Slug,
		Permissions: user.Role.PermissionCodes(),
		Status:      string(user.Status),
	}
}

func passwordFingerprint(hashedPassword string) string {
	sum := sha256.Sum256([]byte(hashedPassword))
	return hex.EncodeToString(sum[:16])
}

func fingerprintMatches(hashedPassword, claim string) bool {
	expected := passwordFingerprint(hashedPassword)
	if len(expected) != len(claim) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(claim)) == 1
}

func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return err
	}

	detail := strings.ToLower(pgErr.ConstraintName + " " + pgErr.Detail)
	switch {
	case strings.Contains(detail, "username"):
		return ErrUsernameTaken
	case strings.Contains(detail, "email"):
		return ErrEmailTaken
	default:
		return ErrEmailTaken
	}
}
