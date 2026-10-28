package token

import (
	"strings"

	"github.com/labstack/echo/v5"
)

const headerTokenKey = "_token"

func ExtractToken(ctx *echo.Context) string {
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[len("Bearer "):]
	}
	return ctx.Request().Header.Get(headerTokenKey)
}
