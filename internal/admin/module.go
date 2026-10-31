package admin

import "github.com/labstack/echo/v5"

// Guard is the auth surface shared by every admin module.
type Guard interface {
	RequireAuth(next echo.HandlerFunc) echo.HandlerFunc
	RequirePermission(codes ...string) echo.MiddlewareFunc
}

// Module is one admin feature area with its own routes.
type Module interface {
	Name() string
	Register(admin *echo.Group)
}
