package rbac

const (
	PermissionActionAdd     = "add"
	PermissionActionRemove  = "remove"
	PermissionActionReplace = "replace"

	UserRoleActionAssign = "assign"
	UserRoleActionRemove = "remove"
)

type PermissionResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
}

type RoleResponse struct {
	ID          string               `json:"id"`
	Slug        string               `json:"slug"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions"`
}

type CreateRoleRequest struct {
	Slug            string   `json:"slug" validate:"required,min=2,max=50"`
	Name            string   `json:"name" validate:"required,min=2,max=100"`
	Description     string   `json:"description" validate:"omitempty,max=255"`
	PermissionCodes []string `json:"permission_codes" validate:"omitempty,dive,required"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

// SetRolePermissionsRequest toggles or replaces permissions on a role.
// action: add | remove | replace
type SetRolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes" validate:"required,min=1,dive,required"`
	Action          string   `json:"action" validate:"required,oneof=add remove replace"`
}

// SetRoleToUserRequest assigns a role to a user or removes that role from them.
// action: assign | remove
// On remove, the user falls back to the default "user" role.
type SetRoleToUserRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Action string `json:"action" validate:"required,oneof=assign remove"`
}

type UserResponse struct {
	ID        string       `json:"id"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Email     string       `json:"email"`
	Role      RoleResponse `json:"role"`
}
