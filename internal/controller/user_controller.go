package controller

import (
	"errors"
	"test/internal/dto"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/response"

	"github.com/labstack/echo/v5"
)

type UserController struct {
	userService service.UserService
	auth        *AuthController
}

func NewUserController(userService service.UserService, auth *AuthController) *UserController {
	return &UserController{userService: userService, auth: auth}
}

func (c *UserController) Register(api *echo.Group) {
	api.GET("/public/users/:handle", c.GetPublicProfile)
	api.POST("/users/create", c.CreateUser, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersCreate))
	api.GET("/users/get-all", c.GetAllUsers, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersRead))
	api.GET("/users/get-by-id/:id", c.GetUserByID, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersRead))
	api.PUT("/users/:id", c.UpdateUser, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersUpdate))
	api.DELETE("/users/:id", c.DeleteUser, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersDelete))
}

// GetPublicProfile returns a shareable profile by username or id.
//
//	@Summary		Public user profile
//	@Tags			users
//	@Produce		json
//	@Param			handle	path		string	true	"Username or user id"
//	@Success		200		{object}	response.APIResponse
//	@Failure		404		{object}	response.APIResponse
//	@Router			/api/public/users/{handle} [get]
func (c *UserController) GetPublicProfile(ctx *echo.Context) error {
	handle := ctx.Param("handle")
	if handle == "" {
		return response.BadRequest(ctx, errors.New("handle is required"))
	}
	profile, err := c.auth.AuthService().GetPublicProfile(ctx.Request().Context(), handle)
	if err != nil {
		return writeUserError(ctx, err)
	}
	return response.OK(ctx, profile)
}

// CreateUser creates a user (admin/staff capability).
//
//	@Summary		Create user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			body	body		dto.UserCreateRequest	true	"User payload"
//	@Success		201		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Failure		409		{object}	response.APIResponse
//	@Router			/api/users/create [post]
func (c *UserController) CreateUser(ctx *echo.Context) error {
	var req dto.UserCreateRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	meta := map[string]any{
		"ip":         ctx.RealIP(),
		"user-agent": ctx.Request().UserAgent(),
	}
	user, err := c.userService.CreateUser(req, meta)
	if err != nil {
		return writeUserError(ctx, err)
	}
	return response.Created(ctx, user)
}

// GetAllUsers lists users.
//
//	@Summary		List users
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			search	query		string	false	"Search name/email/username"
//	@Param			role	query		string	false	"Filter by role slug"
//	@Success		200		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Router			/api/users/get-all [get]
func (c *UserController) GetAllUsers(ctx *echo.Context) error {
	var query dto.UserListQuery
	_ = ctx.Bind(&query)

	users, err := c.userService.ListUsers(query)
	if err != nil {
		return writeUserError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"users": users})
}

// GetUserByID returns one user.
//
//	@Summary		Get user by ID
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	response.APIResponse
//	@Failure		400	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Failure		404	{object}	response.APIResponse
//	@Router			/api/users/get-by-id/{id} [get]
func (c *UserController) GetUserByID(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}
	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return writeUserError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"user": user})
}

// UpdateUser updates a user profile/role/status.
//
//	@Summary		Update user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id		path		string					true	"User ID"
//	@Param			body	body		dto.UserUpdateRequest	true	"Update payload"
//	@Success		200		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Failure		404		{object}	response.APIResponse
//	@Failure		409		{object}	response.APIResponse
//	@Router			/api/users/{id} [put]
func (c *UserController) UpdateUser(ctx *echo.Context) error {
	var req dto.UserUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	actorID := ""
	if actor, err := c.auth.CurrentUser(ctx); err == nil {
		actorID = actor.ID
	}

	user, err := c.userService.UpdateUser(ctx.Param("id"), req, actorID)
	if err != nil {
		return writeUserError(ctx, err)
	}
	return response.OK(ctx, user)
}

// DeleteUser removes a user.
//
//	@Summary		Delete user
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	response.APIResponse
//	@Failure		400	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Failure		404	{object}	response.APIResponse
//	@Router			/api/users/{id} [delete]
func (c *UserController) DeleteUser(ctx *echo.Context) error {
	actorID := ""
	if actor, err := c.auth.CurrentUser(ctx); err == nil {
		actorID = actor.ID
	}

	if err := c.userService.DeleteUser(ctx.Param("id"), actorID); err != nil {
		return writeUserError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"message": "user deleted"})
}

func writeUserError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrUserRoleNotFound),
		errors.Is(err, service.ErrDefaultRoleMissing):
		return response.NotFound(ctx, err)
	case errors.Is(err, service.ErrUserEmailTaken),
		errors.Is(err, service.ErrUserUsernameTaken):
		return response.Conflict(ctx, err)
	case errors.Is(err, service.ErrCannotDeleteSelf),
		errors.Is(err, service.ErrCannotDemoteSelf),
		errors.Is(err, service.ErrLastSuperAdminProtected),
		errors.Is(err, service.ErrInvalidInput):
		return response.BadRequest(ctx, err)
	default:
		return response.InternalServerError(ctx, err)
	}
}
