package service

import (
	"context"
	"errors"
	"test/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrPermissionNotFound = errors.New("permission not found")
)

type RBACService interface {
	GetRoleBySlug(ctx context.Context, slug string) (model.Role, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
	UserHasPermission(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	UserHasAllPermissions(ctx context.Context, userID uuid.UUID, codes ...string) (bool, error)
}

type rbacService struct {
	db *gorm.DB
}

func NewRBACService(db *gorm.DB) RBACService {
	return &rbacService{db: db}
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

func (s *rbacService) ListRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := s.db.WithContext(ctx).
		Preload("Permissions").
		Order("slug ASC").
		Find(&roles).Error
	return roles, err
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
