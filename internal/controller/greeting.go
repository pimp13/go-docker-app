package controller

import (
	"net/http"

	"test/internal/service"

	"github.com/labstack/echo/v5"
)

// GreetingController handles greeting-related HTTP endpoints.
type GreetingController struct {
	greeting service.GreetingService
}

// NewGreetingController injects GreetingService into the controller.
func NewGreetingController(greeting service.GreetingService) *GreetingController {
	return &GreetingController{greeting: greeting}
}

// Register mounts greeting routes onto the given API group.
func (c *GreetingController) Register(api *echo.Group) {
	api.GET("", c.Hello)
	api.GET("/", c.Hello)
	api.GET("/my-token", c.MyToken)
	api.GET("/hello/:name", c.HelloName)
}

// Hello returns a default greeting.
func (c *GreetingController) Hello(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, c.greeting.Hello())
}

// HelloName returns a personalized greeting.
func (c *GreetingController) HelloName(ctx *echo.Context) error {
	name := ctx.Param("name")
	return ctx.JSON(http.StatusOK, c.greeting.HelloName(name))
}

func (c *GreetingController) MyToken(ctx *echo.Context) error {
	token := c.greeting.MyToken()
	if token == "" {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"token": token})
}
