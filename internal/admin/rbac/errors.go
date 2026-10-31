package rbac

import "errors"

var (
	ErrRoleNotFound        = errors.New("role not found")
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrRoleSlugTaken       = errors.New("role slug already exists")
	ErrSystemRoleProtected = errors.New("system roles cannot be deleted")
	ErrRoleInUse           = errors.New("role is assigned to one or more users")
	ErrInvalidRoleSlug     = errors.New("role slug must be lowercase letters, numbers, hyphens, or underscores")
	ErrInvalidInput        = errors.New("invalid input")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserRoleMismatch    = errors.New("user does not currently have this role")
	ErrInvalidAction       = errors.New("invalid action")
)
