package dto

import "time"

type UserCreateRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name"  validate:"required,min=2,max=100"`
	Email     string `json:"email"      validate:"required,email"`
	Username  string `json:"username"   validate:"omitempty,min=3,max=50,alphanum"`
	Password  string `json:"password"   validate:"required,min=8,max=72"`
	Phone     string `json:"phone"      validate:"omitempty,e164"`
	RoleSlug  string `json:"role_slug"  validate:"omitempty,min=2,max=50"`
}

type UserUpdateRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName  *string `json:"last_name"  validate:"omitempty,min=2,max=100"`
	Email     *string `json:"email"      validate:"omitempty,email"`
	Username  *string `json:"username"   validate:"omitempty,min=3,max=50,alphanum"`
	Password  *string `json:"password"   validate:"omitempty,min=8,max=72"`
	Phone     *string `json:"phone"      validate:"omitempty,e164"`
	RoleSlug  *string `json:"role_slug"  validate:"omitempty,min=2,max=50"`
	Status    *string `json:"status"     validate:"omitempty,oneof=active inactive banned"`
}

type UserListQuery struct {
	Search string `query:"search"`
	Role   string `query:"role"` // role slug, e.g. super_admin
}

type UserRoleSummary struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type UserResponse struct {
	ID        string          `json:"id"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Email     string          `json:"email"`
	Username  string          `json:"username,omitempty"`
	Phone     string          `json:"phone,omitempty"`
	Status    string          `json:"status"`
	Role      UserRoleSummary `json:"role"`
	LastLogin *time.Time      `json:"last_login,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
