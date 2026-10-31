package rbac

import (
	"errors"
	"test/internal/model"
	"test/pkg/response"

	"github.com/labstack/echo/v5"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// Register mounts RBAC admin routes under /admin/rbac.
func (c *Controller) Register(admin *echo.Group, requirePerm RequirePermissionFunc) {
	group := admin.Group("/rbac", requirePerm(model.PermRolesManage))

	group.GET("/permissions", c.ListPermissions)
	group.GET("/roles", c.ListRoles)
	group.GET("/roles/:id", c.GetRole)
	group.POST("/roles", c.CreateRole)
	group.PUT("/roles/:id", c.UpdateRole)
	group.PUT("/roles/:id/permissions", c.SetRolePermissions)
	group.DELETE("/roles/:id", c.DeleteRole)
}

// ListPermissions returns all permissions.
//
//	@Summary		List permissions
//	@Tags			admin-rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Success		200	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Router			/api/admin/rbac/permissions [get]
func (c *Controller) ListPermissions(ctx *echo.Context) error {
	permissions, err := c.service.ListPermissions(ctx.Request().Context())
	if err != nil {
		return writeError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"permissions": permissions})
}

// ListRoles returns all roles with permissions.
//
//	@Summary		List roles
//	@Tags			admin-rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Success		200	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Router			/api/admin/rbac/roles [get]
func (c *Controller) ListRoles(ctx *echo.Context) error {
	roles, err := c.service.ListRoles(ctx.Request().Context())
	if err != nil {
		return writeError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"roles": roles})
}

// GetRole returns one role.
//
//	@Summary		Get role by ID
//	@Tags			admin-rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id	path		string	true	"Role ID"
//	@Success		200	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Failure		404	{object}	response.APIResponse
//	@Router			/api/admin/rbac/roles/{id} [get]
func (c *Controller) GetRole(ctx *echo.Context) error {
	role, err := c.service.GetRoleByID(ctx.Request().Context(), ctx.Param("id"))
	if err != nil {
		return writeError(ctx, err)
	}
	return response.OK(ctx, role)
}

// CreateRole creates a custom role.
//
//	@Summary		Create role
//	@Tags			admin-rbac
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			body	body		CreateRoleRequest	true	"Role payload"
//	@Success		201		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Failure		409		{object}	response.APIResponse
//	@Router			/api/admin/rbac/roles [post]
func (c *Controller) CreateRole(ctx *echo.Context) error {
	var req CreateRoleRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	role, err := c.service.CreateRole(ctx.Request().Context(), req)
	if err != nil {
		return writeError(ctx, err)
	}
	return response.Created(ctx, role)
}

// UpdateRole updates role metadata.
//
//	@Summary		Update role
//	@Tags			admin-rbac
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id		path		string				true	"Role ID"
//	@Param			body	body		UpdateRoleRequest	true	"Role update payload"
//	@Success		200		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Failure		404		{object}	response.APIResponse
//	@Router			/api/admin/rbac/roles/{id} [put]
func (c *Controller) UpdateRole(ctx *echo.Context) error {
	var req UpdateRoleRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	role, err := c.service.UpdateRole(ctx.Request().Context(), ctx.Param("id"), req)
	if err != nil {
		return writeError(ctx, err)
	}
	return response.OK(ctx, role)
}

// SetRolePermissions replaces permissions for a role.
//
//	@Summary		Set role permissions
//	@Tags			admin-rbac
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id		path		string						true	"Role ID"
//	@Param			body	body		SetRolePermissionsRequest	true	"Permission codes"
//	@Success		200		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Failure		404		{object}	response.APIResponse
//	@Router			/api/admin/rbac/roles/{id}/permissions [put]
func (c *Controller) SetRolePermissions(ctx *echo.Context) error {
	var req SetRolePermissionsRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	role, err := c.service.SetRolePermissions(ctx.Request().Context(), ctx.Param("id"), req)
	if err != nil {
		return writeError(ctx, err)
	}
	return response.OK(ctx, role)
}

// DeleteRole deletes a non-system role.
//
//	@Summary		Delete role
//	@Tags			admin-rbac
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id	path		string	true	"Role ID"
//	@Success		200	{object}	response.APIResponse
//	@Failure		400	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Failure		404	{object}	response.APIResponse
//	@Router			/api/admin/rbac/roles/{id} [delete]
func (c *Controller) DeleteRole(ctx *echo.Context) error {
	if err := c.service.DeleteRole(ctx.Request().Context(), ctx.Param("id")); err != nil {
		return writeError(ctx, err)
	}
	return response.OK(ctx, map[string]string{"message": "role deleted"})
}

func writeError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrRoleNotFound),
		errors.Is(err, ErrPermissionNotFound):
		return response.NotFound(ctx, err)
	case errors.Is(err, ErrRoleSlugTaken):
		return response.Conflict(ctx, err)
	case errors.Is(err, ErrSystemRoleProtected),
		errors.Is(err, ErrRoleInUse),
		errors.Is(err, ErrInvalidRoleSlug),
		errors.Is(err, ErrInvalidInput):
		return response.BadRequest(ctx, err)
	default:
		return response.InternalServerError(ctx, err)
	}
}
