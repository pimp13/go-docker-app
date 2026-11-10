package sso

import (
	"test/internal/config"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// API owns the SSO domain: service + HTTP controller.
type API struct {
	service Service
	cfg     *config.Config
}

// New builds the SSO module.
func New(db *gorm.DB, cfg *config.Config) *API {
	return &API{
		service: NewService(db, cfg),
		cfg:     cfg,
	}
}

// Service exposes the SSO service for tests or cross-wiring.
func (a *API) Service() Service {
	return a.service
}

// Register mounts SSO routes under the API prefix.
func (a *API) Register(api *echo.Group, authAPI Authenticator, guard Guard) {
	NewController(a.service, authAPI, guard, a.cfg).Register(api)
}
