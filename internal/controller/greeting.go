package controller

import (
	"errors"
	"net/http"

	"test/internal/service"
	"test/pkg/response"

	"github.com/labstack/echo/v5"
)

// GreetingController handles greeting-related HTTP endpoints.
type GreetingController struct {
	greeting service.GreetingService
	version  string
}

// NewGreetingController injects GreetingService into the controller.
func NewGreetingController(greeting service.GreetingService, version string) *GreetingController {
	return &GreetingController{greeting: greeting, version: version}
}

// Register mounts greeting routes onto the given API group.
func (c *GreetingController) Register(api *echo.Group) {
	api.GET("", c.Hello)
	api.GET("/", c.Hello)
	api.POST("/my-token", c.MyToken)
	api.GET("/hello/:name", c.HelloName)
}

// Hello returns a default greeting.
func (c *GreetingController) Hello(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"success":   true,
		"message":   "Server is up!",
		"version":   c.version,
		"IP":        ctx.RealIP(),
		"userAgent": ctx.Request().UserAgent(),
	})
}

// HelloName returns a personalized greeting.
func (c *GreetingController) HelloName(ctx *echo.Context) error {
	name := ctx.Param("name")
	return response.OK(ctx, c.greeting.HelloName(name))
}

func (c *GreetingController) MyToken(ctx *echo.Context) error {
	token := c.greeting.MyToken()
	if token == "" {
		return response.InternalServerError(ctx, errors.New("failed to generate token"))
	}
	return response.OK(ctx, map[string]string{"token": token})
}
