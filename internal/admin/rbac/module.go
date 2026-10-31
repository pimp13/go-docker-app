package rbac

import "github.com/labstack/echo/v5"

// RequirePermissionFunc is provided by the admin auth guard.
type RequirePermissionFunc func(codes ...string) echo.MiddlewareFunc

// Module exposes the admin RBAC feature area.
type Module struct {
	controller  *Controller
	requirePerm RequirePermissionFunc
}

func NewModule(service Service, requirePerm RequirePermissionFunc) *Module {
	return &Module{
		controller:  NewController(service),
		requirePerm: requirePerm,
	}
}

func (m *Module) Name() string {
	return "rbac"
}

func (m *Module) Register(admin *echo.Group) {
	m.controller.Register(admin, m.requirePerm)
}
