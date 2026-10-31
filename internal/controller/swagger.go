package controller

import (
	"github.com/labstack/echo/v5"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

// RegisterSwagger mounts the Swagger UI and generated OpenAPI docs.
func RegisterSwagger(e *echo.Echo) {
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}
