package sso

import (
	"context"
	"test/internal/dto"

	"github.com/labstack/echo/v5"
)

// Guard is the auth surface SSO needs from the general auth controller.
type Guard interface {
	RequireAuth(next echo.HandlerFunc) echo.HandlerFunc
	TokenFromRequest(ctx *echo.Context) (string, error)
	CurrentUser(ctx *echo.Context) (dto.AuthUser, error)
}

// Authenticator validates session JWTs for the authorize flow.
type Authenticator interface {
	Authenticate(ctx context.Context, rawToken string) (dto.AuthUser, error)
}
