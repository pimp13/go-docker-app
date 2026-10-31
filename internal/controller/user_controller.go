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
	api.POST("/users/create", c.CreateUser, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersCreate))
	api.GET("/users/get-all", c.GetAllUsers, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersRead))
	api.GET("/users/get-by-id/:id", c.GetUserByID, c.auth.RequireAuth, c.auth.RequirePermission(model.PermUsersRead))
}

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
	result, err := c.userService.CreateUser(req, meta)
	if err != nil {
		return response.InternalServerError(ctx, err)
	}
	return response.OK(ctx, result)
}

func (c *UserController) GetAllUsers(ctx *echo.Context) error {
	users, err := c.userService.GetAllUsers()
	if err != nil {
		return response.InternalServerError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"users": users})
}

func (c *UserController) GetUserByID(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}
	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return response.InternalServerError(ctx, err)
	}
	return response.OK(ctx, map[string]any{"user": user})
}
