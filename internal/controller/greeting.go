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
//
//	@Summary		API root status
//	@Description	Simple readiness payload for the API prefix
//	@Tags			general
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/api [get]
//	@Router			/api/ [get]
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
//
//	@Summary		Personalized greeting
//	@Tags			general
//	@Produce		json
//	@Param			name	path		string	true	"Name"
//	@Success		200		{object}	response.APIResponse
//	@Router			/api/hello/{name} [get]
func (c *GreetingController) HelloName(ctx *echo.Context) error {
	name := ctx.Param("name")
	return response.OK(ctx, c.greeting.HelloName(name))
}

// MyToken generates a development token.
//
//	@Summary		Generate development token
//	@Tags			general
//	@Produce		json
//	@Success		200	{object}	response.APIResponse
//	@Failure		500	{object}	response.APIResponse
//	@Router			/api/my-token [post]
func (c *GreetingController) MyToken(ctx *echo.Context) error {
	token := c.greeting.MyToken()
	if token == "" {
		return response.InternalServerError(ctx, errors.New("failed to generate token"))
	}
	return response.OK(ctx, map[string]string{"token": token})
}
