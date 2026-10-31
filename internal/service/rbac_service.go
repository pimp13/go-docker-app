package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"test/internal/dto"
	"test/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound        = errors.New("role not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrRoleSlugTaken       = errors.New("role slug already exists")
	ErrSystemRoleProtected = errors.New("system roles cannot be deleted")
	ErrRoleInUse           = errors.New("role is assigned to one or more users")
	ErrInvalidRoleSlug     = errors.New("role slug must be lowercase letters, numbers, hyphens, or underscores")
)

var roleSlugPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

type RBACService interface {
	ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error)
	ListRoles(ctx context.Context) ([]dto.RoleResponse, error)
	GetRoleByID(ctx context.Context, id string) (dto.RoleResponse, error)
	GetRoleBySlug(ctx context.Context, slug string) (model.Role, error)
	CreateRole(ctx context.Context, req dto.CreateRoleRequest) (dto.RoleResponse, error)
	UpdateRole(ctx context.Context, id string, req dto.UpdateRoleRequest) (dto.RoleResponse, error)
	SetRolePermissions(ctx context.Context, id string, req dto.SetRolePermissionsRequest) (dto.RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error
	UserHasPermission(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	UserHasAllPermissions(ctx context.Context, userID uuid.UUID, codes ...string) (bool, error)
}

type rbacService struct {
	db *gorm.DB
}

func NewRBACService(db *gorm.DB) RBACService {
	return &rbacService{db: db}
}

func (s *rbacService) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	var permissions []model.Permission
	if err := s.db.WithContext(ctx).Order("resource ASC, action ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return toPermissionResponses(permissions), nil
}

func (s *rbacService) ListRoles(ctx context.Context) ([]dto.RoleResponse, error) {
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

func (s *rbacService) GetRoleByID(ctx context.Context, id string) (dto.RoleResponse, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return dto.RoleResponse{}, ErrRoleNotFound
	}

	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return toRoleResponse(role), nil
}

func (s *rbacService) GetRoleBySlug(ctx context.Context, slug string) (model.Role, error) {
	var role model.Role
	err := s.db.WithContext(ctx).
		Preload("Permissions").
		Where("slug = ?", slug).
		Take(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Role{}, ErrRoleNotFound
	}
	return role, err
}

func (s *rbacService) CreateRole(ctx context.Context, req dto.CreateRoleRequest) (dto.RoleResponse, error) {
	slug, err := normalizeRoleSlug(req.Slug)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	if name == "" {
		return dto.RoleResponse{}, ErrInvalidInput
	}

	permissions, err := s.findPermissionsByCodes(ctx, req.PermissionCodes)
	if err != nil {
		return dto.RoleResponse{}, err
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
		return dto.RoleResponse{}, err
	}

	stored, err := s.findRoleByID(ctx, role.ID)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return toRoleResponse(stored), nil
}

func (s *rbacService) UpdateRole(ctx context.Context, id string, req dto.UpdateRoleRequest) (dto.RoleResponse, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return dto.RoleResponse{}, ErrRoleNotFound
	}

	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	if name == "" {
		return dto.RoleResponse{}, ErrInvalidInput
	}

	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return dto.RoleResponse{}, err
	}

	if err := s.db.WithContext(ctx).Model(&role).Updates(map[string]any{
		"name":        name,
		"description": description,
	}).Error; err != nil {
		return dto.RoleResponse{}, err
	}

	stored, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return toRoleResponse(stored), nil
}

func (s *rbacService) SetRolePermissions(ctx context.Context, id string, req dto.SetRolePermissionsRequest) (dto.RoleResponse, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return dto.RoleResponse{}, ErrRoleNotFound
	}

	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return dto.RoleResponse{}, err
	}

	permissions, err := s.findPermissionsByCodes(ctx, req.PermissionCodes)
	if err != nil {
		return dto.RoleResponse{}, err
	}

	if err := s.db.WithContext(ctx).Model(&role).Association("Permissions").Replace(permissions); err != nil {
		return dto.RoleResponse{}, err
	}

	stored, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return toRoleResponse(stored), nil
}

func (s *rbacService) DeleteRole(ctx context.Context, id string) error {
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

func (s *rbacService) UserHasPermission(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	return s.UserHasAllPermissions(ctx, userID, code)
}

func (s *rbacService) UserHasAllPermissions(ctx context.Context, userID uuid.UUID, codes ...string) (bool, error) {
	if len(codes) == 0 {
		return true, nil
	}

	var user model.User
	err := s.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("id = ?", userID).
		Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, ErrRoleNotFound
	}
	if err != nil {
		return false, err
	}

	for _, code := range codes {
		if !user.HasPermission(code) {
			return false, nil
		}
	}
	return true, nil
}

func (s *rbacService) findRoleByID(ctx context.Context, id uuid.UUID) (model.Role, error) {
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

func (s *rbacService) findPermissionsByCodes(ctx context.Context, codes []string) ([]model.Permission, error) {
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

func toRoleResponses(roles []model.Role) []dto.RoleResponse {
	out := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		out = append(out, toRoleResponse(role))
	}
	return out
}

func toRoleResponse(role model.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:          role.ID.String(),
		Slug:        role.Slug,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: toPermissionResponses(role.Permissions),
	}
}

func toPermissionResponses(permissions []model.Permission) []dto.PermissionResponse {
	out := make([]dto.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		out = append(out, dto.PermissionResponse{
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
