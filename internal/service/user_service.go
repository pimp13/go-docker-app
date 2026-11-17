package service

import (
	"errors"
	"fmt"
	"strings"
	"test/internal/dto"
	"test/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(req dto.UserCreateRequest, meta map[string]any) (dto.UserResponse, error)
	UpdateUser(id string, req dto.UserUpdateRequest, actorID string) (dto.UserResponse, error)
	DeleteUser(id string, actorID string) error
	ListUsers(query dto.UserListQuery) ([]dto.UserResponse, error)
	GetUserByID(id string) (dto.UserResponse, error)
	SecretKey() string
}

type userService struct {
	db        *gorm.DB
	secretKey string
}

func NewUserService(db *gorm.DB, secretKey string) UserService {
	return &userService{db: db, secretKey: secretKey}
}

func (s *userService) SecretKey() string {
	return s.secretKey
}

func (s *userService) CreateUser(
	req dto.UserCreateRequest,
	meta map[string]any,
) (dto.UserResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	req.RoleSlug = strings.ToLower(strings.TrimSpace(req.RoleSlug))

	if meta == nil {
		meta = map[string]any{}
	}

	roleSlug := req.RoleSlug
	if roleSlug == "" {
		roleSlug = model.RoleSlugUser
	}

	role, err := s.findRoleBySlug(roleSlug)
	if err != nil {
		return dto.UserResponse{}, err
	}

	if err := s.ensureEmailAvailable(req.Email, uuid.Nil); err != nil {
		return dto.UserResponse{}, err
	}
	if req.Username != "" {
		if err := s.ensureUsernameAvailable(req.Username, uuid.Nil); err != nil {
			return dto.UserResponse{}, err
		}
	}

	user := model.User{
		ID:        uuid.New(),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		Phone:     strings.TrimSpace(req.Phone),
		RoleID:    role.ID,
		Status:    model.UserStatusActive,
		Meta:      meta,
	}

	create := s.db
	if req.Username == "" {
		create = create.Omit("Username")
	}
	if err := create.Create(&user).Error; err != nil {
		return dto.UserResponse{}, err
	}

	return s.GetUserByID(user.ID.String())
}

func (s *userService) UpdateUser(
	id string,
	req dto.UserUpdateRequest,
	actorID string,
) (dto.UserResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return dto.UserResponse{}, ErrUserNotFound
	}

	var user model.User
	if err := s.db.Preload("Role").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserResponse{}, ErrUserNotFound
		}
		return dto.UserResponse{}, err
	}

	updates := map[string]any{}

	if req.FirstName != nil {
		updates["first_name"] = strings.TrimSpace(*req.FirstName)
	}
	if req.LastName != nil {
		updates["last_name"] = strings.TrimSpace(*req.LastName)
	}
	if req.Phone != nil {
		updates["phone"] = strings.TrimSpace(*req.Phone)
	}
	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		if err := s.ensureEmailAvailable(email, user.ID); err != nil {
			return dto.UserResponse{}, err
		}
		updates["email"] = email
	}
	if req.Username != nil {
		username := strings.ToLower(strings.TrimSpace(*req.Username))
		if username == "" {
			updates["username"] = nil
		} else {
			if err := s.ensureUsernameAvailable(username, user.ID); err != nil {
				return dto.UserResponse{}, err
			}
			updates["username"] = username
		}
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return dto.UserResponse{}, err
		}
		updates["password"] = string(hashed)
	}
	if req.Status != nil {
		updates["status"] = model.UserStatus(strings.ToLower(strings.TrimSpace(*req.Status)))
	}

	if req.RoleSlug != nil {
		roleSlug := strings.ToLower(strings.TrimSpace(*req.RoleSlug))
		role, err := s.findRoleBySlug(roleSlug)
		if err != nil {
			return dto.UserResponse{}, err
		}

		if user.Role.Slug == model.RoleSlugSuperAdmin && role.Slug != model.RoleSlugSuperAdmin {
			if actorID != "" && actorID == user.ID.String() {
				return dto.UserResponse{}, ErrCannotDemoteSelf
			}
			if err := s.ensureNotLastSuperAdmin(user.ID); err != nil {
				return dto.UserResponse{}, err
			}
		}
		updates["role_id"] = role.ID
	}

	if len(updates) == 0 {
		return toUserResponse(user), nil
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return dto.UserResponse{}, err
	}

	return s.GetUserByID(user.ID.String())
}

func (s *userService) DeleteUser(id string, actorID string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return ErrUserNotFound
	}
	if actorID != "" && actorID == userID.String() {
		return ErrCannotDeleteSelf
	}

	var user model.User
	if err := s.db.Preload("Role").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if user.Role.Slug == model.RoleSlugSuperAdmin {
		if err := s.ensureNotLastSuperAdmin(user.ID); err != nil {
			return err
		}
	}

	return s.db.Delete(&user).Error
}

func (s *userService) ListUsers(query dto.UserListQuery) ([]dto.UserResponse, error) {
	db := s.db.Model(&model.User{}).Preload("Role")

	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		db = db.Where(
			`LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(COALESCE(username, '')) LIKE ?`,
			like, like, like, like,
		)
	}

	if role := strings.ToLower(strings.TrimSpace(query.Role)); role != "" {
		db = db.Joins("JOIN roles ON roles.id = users.role_id").Where("roles.slug = ?", role)
	}

	var users []model.User
	if err := db.Order("users.created_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	out := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		out = append(out, toUserResponse(user))
	}
	return out, nil
}

func (s *userService) GetUserByID(id string) (dto.UserResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return dto.UserResponse{}, ErrUserNotFound
	}

	var user model.User
	if err := s.db.Preload("Role").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserResponse{}, ErrUserNotFound
		}
		return dto.UserResponse{}, fmt.Errorf("failed to get user by id: %w", err)
	}
	return toUserResponse(user), nil
}

func (s *userService) findRoleBySlug(slug string) (model.Role, error) {
	var role model.Role
	if err := s.db.Where("slug = ?", slug).Take(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if slug == model.RoleSlugUser {
				return model.Role{}, ErrDefaultRoleMissing
			}
			return model.Role{}, ErrUserRoleNotFound
		}
		return model.Role{}, err
	}
	return role, nil
}

func (s *userService) ensureEmailAvailable(email string, exceptID uuid.UUID) error {
	q := s.db.Model(&model.User{}).Where("email = ?", email)
	if exceptID != uuid.Nil {
		q = q.Where("id <> ?", exceptID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrUserEmailTaken
	}
	return nil
}

func (s *userService) ensureUsernameAvailable(username string, exceptID uuid.UUID) error {
	q := s.db.Model(&model.User{}).Where("username = ?", username)
	if exceptID != uuid.Nil {
		q = q.Where("id <> ?", exceptID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrUserUsernameTaken
	}
	return nil
}

func (s *userService) ensureNotLastSuperAdmin(exceptID uuid.UUID) error {
	var count int64
	err := s.db.Model(&model.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.slug = ? AND users.id <> ?", model.RoleSlugSuperAdmin, exceptID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrLastSuperAdminProtected
	}
	return nil
}

func toUserResponse(user model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Username:  user.Username,
		Phone:     user.Phone,
		Status:    string(user.Status),
		Role: dto.UserRoleSummary{
			ID:   user.Role.ID.String(),
			Slug: user.Role.Slug,
			Name: user.Role.Name,
		},
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
