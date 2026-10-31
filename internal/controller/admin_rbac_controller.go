package controller

import (
	"errors"
	"test/internal/dto"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/response"

	"github.com/labstack/echo/v5"
)

type AdminRBACController struct {
	rbac service.RBACService
	auth *AuthController
}

func NewAdminRBACController(rbac service.RBACService, auth *AuthController) *AdminRBACController {
	return &AdminRBACController{rbac: rbac, auth: auth}
}

func (c *AdminRBACController) Register(api *echo.Group) {
	admin := api.Group("/admin", c.auth.RequireAuth, c.auth.RequirePermission(model.PermRolesManage))

	admin.GET("/permissions", c.ListPermissions)
	admin.GET("/roles", c.ListRoles)
	admin.GET("/roles/:id", c.GetRole)
	admin.POST("/roles", c.CreateRole)
	admin.PUT("/roles/:id", c.UpdateRole)
	admin.PUT("/roles/:id/permissions", c.SetRolePermissions)
	admin.DELETE("/roles/:id", c.DeleteRole)
}

func (c *AdminRBACController) ListPermissions(ctx *echo.Context) error {
	permissions, err := c.rbac.ListPermissions(ctx.Request().Context())
	if err != nil {
		return writeRBACError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"permissions": permissions})
}

func (c *AdminRBACController) ListRoles(ctx *echo.Context) error {
	roles, err := c.rbac.ListRoles(ctx.Request().Context())
	if err != nil {
		return writeRBACError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"roles": roles})
}

func (c *AdminRBACController) GetRole(ctx *echo.Context) error {
	role, err := c.rbac.GetRoleByID(ctx.Request().Context(), ctx.Param("id"))
	if err != nil {
		return writeRBACError(ctx, err)
	}
	return response.OK(ctx, role)
}

func (c *AdminRBACController) CreateRole(ctx *echo.Context) error {
	var req dto.CreateRoleRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	role, err := c.rbac.CreateRole(ctx.Request().Context(), req)
	if err != nil {
		return writeRBACError(ctx, err)
	}
	return response.Created(ctx, role)
}

func (c *AdminRBACController) UpdateRole(ctx *echo.Context) error {
	var req dto.UpdateRoleRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	role, err := c.rbac.UpdateRole(ctx.Request().Context(), ctx.Param("id"), req)
	if err != nil {
		return writeRBACError(ctx, err)
	}
	return response.OK(ctx, role)
}

func (c *AdminRBACController) SetRolePermissions(ctx *echo.Context) error {
	var req dto.SetRolePermissionsRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	role, err := c.rbac.SetRolePermissions(ctx.Request().Context(), ctx.Param("id"), req)
	if err != nil {
		return writeRBACError(ctx, err)
	}
	return response.OK(ctx, role)
}

func (c *AdminRBACController) DeleteRole(ctx *echo.Context) error {
	if err := c.rbac.DeleteRole(ctx.Request().Context(), ctx.Param("id")); err != nil {
		return writeRBACError(ctx, err)
	}
	return response.OK(ctx, map[string]string{"message": "role deleted"})
}

func writeRBACError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrRoleNotFound),
		errors.Is(err, service.ErrPermissionNotFound):
		return response.NotFound(ctx, err)
	case errors.Is(err, service.ErrRoleSlugTaken):
		return response.Conflict(ctx, err)
	case errors.Is(err, service.ErrSystemRoleProtected),
		errors.Is(err, service.ErrRoleInUse),
		errors.Is(err, service.ErrInvalidRoleSlug),
		errors.Is(err, service.ErrInvalidInput):
		return response.BadRequest(ctx, err)
	case errors.Is(err, service.ErrPermissionDenied):
		return response.Forbidden(ctx, err)
	default:
		return response.InternalServerError(ctx, err)
	}
}
