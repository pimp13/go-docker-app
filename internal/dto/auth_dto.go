package dto

import "time"

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Username  string `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
	Phone     string `json:"phone" validate:"omitempty,e164"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=72"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,max=72"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Username  string `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
	Phone     string `json:"phone" validate:"omitempty,e164"`
	Avatar    string `json:"avatar" validate:"omitempty,max=500"`
}

// AuthUser is the public profile returned by auth endpoints.
type AuthUser struct {
	ID          string   `json:"id"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	Username    string   `json:"username,omitempty"`
	Phone       string   `json:"phone,omitempty"`
	Avatar      string   `json:"avatar,omitempty"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status"`
}

// PublicProfile is a shareable profile by username or id.
type PublicProfile struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Username  string    `json:"username,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
	Role      string    `json:"role"`
	RoleName  string    `json:"role_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// HasPermission reports whether the authenticated user includes the code.
func (u AuthUser) HasPermission(code string) bool {
	for _, p := range u.Permissions {
		if p == code {
			return true
		}
	}
	return false
}

// HasAllPermissions reports whether every required code is present.
func (u AuthUser) HasAllPermissions(codes ...string) bool {
	for _, code := range codes {
		if !u.HasPermission(code) {
			return false
		}
	}
	return true
}

// Session is the result of a successful login or password change.
// Token is written to the auth cookie and is never serialized in JSON.
type Session struct {
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	User      AuthUser  `json:"user"`
}
