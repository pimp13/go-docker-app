package controller

import (
	"net/http"
	"test/internal/dto"
	"test/internal/service"
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
}

func (c *UserController) CreateUser(ctx *echo.Context) error {
	var req dto.UserCreateRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	result, err := c.userService.CreateUser(req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, result)
}

func (c *UserController) GetAllUsers(ctx *echo.Context) error {
	tokenStr := ctx.Request().Header.Get("token")
	if tokenStr == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "token is required"})
	}
	if err := token.ValidateToken(tokenStr, c.userService.SecretKey()); err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	users, err := c.userService.GetAllUsers()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, users)
}
