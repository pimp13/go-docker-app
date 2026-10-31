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
//
//	@Summary		Health check
//	@Description	Returns service readiness status and version
//	@Tags			general
//	@Produce		json
//	@Success		200	{object}	service.HealthStatus
//	@Router			/health [get]
func (c *HealthController) Check(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, c.health.Check())
}
