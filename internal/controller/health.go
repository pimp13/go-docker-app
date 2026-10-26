package controller

import (
	"net/http"

	"test/internal/service"

	"github.com/labstack/echo/v5"
)

// HealthController handles health-check HTTP endpoints.
type HealthController struct {
	health service.HealthService
}

// NewHealthController injects HealthService into the controller.
func NewHealthController(health service.HealthService) *HealthController {
	return &HealthController{health: health}
}

// Check returns application health status.
func (c *HealthController) Check(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, c.health.Check())
}
