package rbac

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

type SetRolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes" validate:"required,dive,required"`
}
