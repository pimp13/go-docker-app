package rbac

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"test/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var roleSlugPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// Service manages roles and permissions for the admin panel.
type Service interface {
	ListPermissions(ctx context.Context) ([]PermissionResponse, error)
	ListRoles(ctx context.Context) ([]RoleResponse, error)
	GetRoleByID(ctx context.Context, id string) (RoleResponse, error)
	CreateRole(ctx context.Context, req CreateRoleRequest) (RoleResponse, error)
	UpdateRole(ctx context.Context, id string, req UpdateRoleRequest) (RoleResponse, error)
	SetRolePermissions(ctx context.Context, id string, req SetRolePermissionsRequest) (RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error

	SetRoleToUser(ctx context.Context, id string, req SetRoleToUserRequest) (UserResponse, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) ListPermissions(ctx context.Context) ([]PermissionResponse, error) {
	var permissions []model.Permission
	if err := s.db.WithContext(ctx).Order("resource ASC, action ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return toPermissionResponses(permissions), nil
}

func (s *service) ListRoles(ctx context.Context) ([]RoleResponse, error) {
	var roles []model.Role
	if err := s.db.WithContext(ctx).
		Preload("Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("resource ASC, action ASC")
		}).
		Order("slug ASC").
		Find(&roles).Error; err != nil {
		return nil, err
	}
	return toRoleResponses(roles), nil
}

func (s *service) GetRoleByID(ctx context.Context, id string) (RoleResponse, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return RoleResponse{}, ErrRoleNotFound
	}

	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return RoleResponse{}, err
	}
	return toRoleResponse(role), nil
}

func (s *service) CreateRole(ctx context.Context, req CreateRoleRequest) (RoleResponse, error) {
	slug, err := normalizeRoleSlug(req.Slug)
	if err != nil {
		return RoleResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	if name == "" {
		return RoleResponse{}, ErrInvalidInput
	}

	permissions, err := s.findPermissionsByCodes(ctx, req.PermissionCodes)
	if err != nil {
		return RoleResponse{}, err
	}

	role := model.Role{
		ID:          uuid.New(),
		Slug:        slug,
		Name:        name,
		Description: description,
		IsSystem:    false,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return mapRoleUniqueViolation(err)
		}
		if len(permissions) > 0 {
			if err := tx.Model(&role).Association("Permissions").Replace(permissions); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return RoleResponse{}, err
	}

	stored, err := s.findRoleByID(ctx, role.ID)
	if err != nil {
		return RoleResponse{}, err
	}
	return toRoleResponse(stored), nil
}

func (s *service) UpdateRole(ctx context.Context, id string, req UpdateRoleRequest) (RoleResponse, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return RoleResponse{}, ErrRoleNotFound
	}

	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	if name == "" {
		return RoleResponse{}, ErrInvalidInput
	}

	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return RoleResponse{}, err
	}

	if err := s.db.WithContext(ctx).Model(&role).Updates(map[string]any{
		"name":        name,
		"description": description,
	}).Error; err != nil {
		return RoleResponse{}, err
	}

	stored, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return RoleResponse{}, err
	}
	return toRoleResponse(stored), nil
}

func (s *service) SetRolePermissions(ctx context.Context, id string, req SetRolePermissionsRequest) (RoleResponse, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return RoleResponse{}, ErrRoleNotFound
	}

	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return RoleResponse{}, err
	}

	permissions, err := s.findPermissionsByCodes(ctx, req.PermissionCodes)
	if err != nil {
		return RoleResponse{}, err
	}

	assoc := s.db.WithContext(ctx).Model(&role).Association("Permissions")
	switch strings.ToLower(strings.TrimSpace(req.Action)) {
	case PermissionActionAdd:
		if err := assoc.Append(permissions); err != nil {
			return RoleResponse{}, err
		}
	case PermissionActionRemove:
		if err := assoc.Delete(permissions); err != nil {
			return RoleResponse{}, err
		}
	case PermissionActionReplace:
		if err := assoc.Replace(permissions); err != nil {
			return RoleResponse{}, err
		}
	default:
		return RoleResponse{}, ErrInvalidAction
	}

	stored, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return RoleResponse{}, err
	}
	return toRoleResponse(stored), nil
}

func (s *service) DeleteRole(ctx context.Context, id string) error {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return ErrRoleNotFound
	}

	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return ErrSystemRoleProtected
	}

	var users int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("role_id = ?", roleID).Count(&users).Error; err != nil {
		return err
	}
	if users > 0 {
		return ErrRoleInUse
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&role).Association("Permissions").Clear(); err != nil {
			return err
		}
		return tx.Delete(&role).Error
	})
}

func (s *service) SetRoleToUser(
	ctx context.Context,
	id string,
	req SetRoleToUserRequest,
) (UserResponse, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return UserResponse{}, ErrRoleNotFound
	}
	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return UserResponse{}, err
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return UserResponse{}, ErrUserNotFound
	}

	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		return UserResponse{}, err
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	switch action {
	case UserRoleActionAssign:
		if err := s.db.WithContext(ctx).
			Model(&model.User{}).
			Where("id = ?", user.ID).
			Update("role_id", role.ID).Error; err != nil {
			return UserResponse{}, err
		}
	case UserRoleActionRemove:
		if user.RoleID != role.ID {
			return UserResponse{}, ErrUserRoleMismatch
		}
		defaultRole, err := s.findRoleBySlug(ctx, model.RoleSlugUser)
		if err != nil {
			return UserResponse{}, err
		}
		if defaultRole.ID == role.ID {
			return UserResponse{}, ErrInvalidInput
		}
		if err := s.db.WithContext(ctx).
			Model(&model.User{}).
			Where("id = ?", user.ID).
			Update("role_id", defaultRole.ID).Error; err != nil {
			return UserResponse{}, err
		}
	default:
		return UserResponse{}, ErrInvalidAction
	}

	stored, err := s.findUserByID(ctx, user.ID)
	if err != nil {
		return UserResponse{}, err
	}
	return toUserResponse(stored), nil
}

func (s *service) findRoleByID(ctx context.Context, id uuid.UUID) (model.Role, error) {
	var role model.Role
	err := s.db.WithContext(ctx).
		Preload("Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("resource ASC, action ASC")
		}).
		Where("id = ?", id).
		Take(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Role{}, ErrRoleNotFound
	}
	return role, err
}

func (s *service) findRoleBySlug(ctx context.Context, slug string) (model.Role, error) {
	var role model.Role
	err := s.db.WithContext(ctx).
		Preload("Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("resource ASC, action ASC")
		}).
		Where("slug = ?", slug).
		Take(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Role{}, ErrRoleNotFound
	}
	return role, err
}

func (s *service) findUserByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).
		Preload("Role.Permissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("resource ASC, action ASC")
		}).
		Where("id = ?", id).
		Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrUserNotFound
	}
	return user, err
}

func (s *service) findPermissionsByCodes(ctx context.Context, codes []string) ([]model.Permission, error) {
	normalized := uniqueNonEmpty(codes)
	if len(normalized) == 0 {
		return []model.Permission{}, nil
	}

	var permissions []model.Permission
	if err := s.db.WithContext(ctx).Where("code IN ?", normalized).Find(&permissions).Error; err != nil {
		return nil, err
	}
	if len(permissions) != len(normalized) {
		found := make(map[string]struct{}, len(permissions))
		for _, p := range permissions {
			found[p.Code] = struct{}{}
		}
		for _, code := range normalized {
			if _, ok := found[code]; !ok {
				return nil, fmt.Errorf("%w: %s", ErrPermissionNotFound, code)
			}
		}
		return nil, ErrPermissionNotFound
	}
	return permissions, nil
}

func normalizeRoleSlug(raw string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	if !roleSlugPattern.MatchString(slug) || len(slug) < 2 || len(slug) > 50 {
		return "", ErrInvalidRoleSlug
	}
	return slug, nil
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func mapRoleUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrRoleSlugTaken
	}
	return err
}

func toRoleResponses(roles []model.Role) []RoleResponse {
	out := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		out = append(out, toRoleResponse(role))
	}
	return out
}

func toRoleResponse(role model.Role) RoleResponse {
	return RoleResponse{
		ID:          role.ID.String(),
		Slug:        role.Slug,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: toPermissionResponses(role.Permissions),
	}
}

func toPermissionResponses(permissions []model.Permission) []PermissionResponse {
	out := make([]PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		out = append(out, PermissionResponse{
			ID:          permission.ID.String(),
			Code:        permission.Code,
			Name:        permission.Name,
			Resource:    permission.Resource,
			Action:      permission.Action,
			Description: permission.Description,
		})
	}
	return out
}

func toUserResponse(user model.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      toRoleResponse(user.Role),
	}
}
