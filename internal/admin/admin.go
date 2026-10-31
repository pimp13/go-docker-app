package admin

import (
	adminrbac "test/internal/admin/rbac"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Services is the DI container for all admin-domain services.
type Services struct {
	RBAC adminrbac.Service
}

// API wires admin modules under /admin.
type API struct {
	services *Services
}

// New builds admin services.
// Add future admin services here (users, articles, settings, …).
func New(db *gorm.DB) *API {
	return &API{
		services: &Services{
			RBAC: adminrbac.NewService(db),
		},
	}
}

// Services exposes admin services for tests or cross-wiring.
func (a *API) Services() *Services {
	return a.services
}

// Register mounts every admin module on /admin.
// Base authentication is applied once; each module adds its own permissions.
func (a *API) Register(api *echo.Group, guard Guard) {
	adminGroup := api.Group("/admin", guard.RequireAuth)

	for _, module := range a.modules(guard) {
		module.Register(adminGroup)
	}
}

// modules lists every admin feature area.
// To add a new admin section: create internal/admin/<name>, then append it here.
func (a *API) modules(guard Guard) []Module {
	return []Module{
		adminrbac.NewModule(a.services.RBAC, guard.RequirePermission),
	}
}
