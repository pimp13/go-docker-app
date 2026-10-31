package model

import (
	"time"

	"github.com/google/uuid"
)

// Built-in role slugs.
const (
	RoleSlugUser   = "user"
	RoleSlugEditor = "editor"
	RoleSlugAdmin  = "admin"
)

// Permission codes follow resource:action.
const (
	PermArticlesRead   = "articles:read"
	PermArticlesCreate = "articles:create"
	PermArticlesUpdate = "articles:update"
	PermArticlesDelete = "articles:delete"

	PermUsersRead   = "users:read"
	PermUsersCreate = "users:create"
	PermUsersUpdate = "users:update"
	PermUsersDelete = "users:delete"

	PermRolesManage = "roles:manage"
)

// Role is an assignable access profile.
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Slug        string       `gorm:"size:50;not null;uniqueIndex" json:"slug"`
	Name        string       `gorm:"size:100;not null" json:"name"`
	Description string       `gorm:"size:255" json:"description,omitempty"`
	IsSystem    bool         `gorm:"not null;default:false" json:"is_system"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

// HasPermission reports whether the role includes the given permission code.
func (r Role) HasPermission(code string) bool {
	for _, p := range r.Permissions {
		if p.Code == code {
			return true
		}
	}
	return false
}

// PermissionCodes returns the role's permission codes.
func (r Role) PermissionCodes() []string {
	codes := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		codes = append(codes, p.Code)
	}
	return codes
}

// Permission is a fine-grained capability.
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Code        string    `gorm:"size:100;not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Resource    string    `gorm:"size:50;not null;index" json:"resource"`
	Action      string    `gorm:"size:50;not null" json:"action"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
}

func (Permission) TableName() string {
	return "permissions"
}

// RolePermission is the explicit join table for role ↔ permission.
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
