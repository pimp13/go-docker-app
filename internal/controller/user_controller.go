package controller

import (
	"errors"
	"test/internal/dto"
	"test/internal/service"
	"test/pkg/response"
	"test/pkg/token"

	"github.com/labstack/echo/v5"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService}
}

func (c *UserController) Register(api *echo.Group) {
	api.POST("/users/create", c.CreateUser)
	api.GET("/users/get-all", c.GetAllUsers)
	api.GET("/users/get-by-id/:id", c.GetUserByID)
}

func (c *UserController) CreateUser(ctx *echo.Context) error {
	tokenStr := token.ExtractToken(ctx)
	if tokenStr == "" {
		return response.Unauthorized(ctx, errors.New("token is required"))
	}
	if err := token.ValidateToken(tokenStr, c.userService.SecretKey()); err != nil {
		return response.Unauthorized(ctx, err)
	}
	var req dto.UserCreateRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	meta := map[string]any{
		"ip":         ctx.RealIP(),
		"user-agent": ctx.Request().UserAgent(),
		"token":      tokenStr,
	}
	result, err := c.userService.CreateUser(req, meta)
	if err != nil {
		return response.InternalServerError(ctx, err)
	}
	return response.OK(ctx, result)
}

func (c *UserController) GetAllUsers(ctx *echo.Context) error {
	tokenStr := token.ExtractToken(ctx)
	if tokenStr == "" {
		return response.Unauthorized(ctx, errors.New("token is required"))
	}
	if err := token.ValidateToken(tokenStr, c.userService.SecretKey()); err != nil {
		return response.Unauthorized(ctx, err)
	}

	users, err := c.userService.GetAllUsers()
	if err != nil {
		return response.InternalServerError(ctx, err)
	}
	finalRes := map[string]any{
		"users": users,
		"token": tokenStr,
	}
	return response.OK(ctx, finalRes)
}

func (c *UserController) GetUserByID(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}
	tokenStr := token.ExtractToken(ctx)
	if tokenStr == "" {
		return response.Unauthorized(ctx, errors.New("token is required"))
	}
	if err := token.ValidateToken(tokenStr, c.userService.SecretKey()); err != nil {
		return response.Unauthorized(ctx, err)
	}
	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return response.InternalServerError(ctx, err)
	}
	finalRes := map[string]any{
		"user":  user,
		"token": tokenStr,
	}
	return response.OK(ctx, finalRes)
}
