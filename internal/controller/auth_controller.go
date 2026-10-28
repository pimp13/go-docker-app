package controller

import (
	"errors"
	"net/http"
	"strings"
	"test/internal/config"
	"test/internal/dto"
	"test/internal/service"
	"test/pkg/response"
	"time"

	"github.com/labstack/echo/v5"
)

const authUserKey = "auth_user"

type AuthController struct {
	auth service.AuthService
	cfg  config.AuthConfig
}

func NewAuthController(auth service.AuthService, cfg config.AuthConfig) *AuthController {
	return &AuthController{auth: auth, cfg: cfg}
}

func (c *AuthController) Register(api *echo.Group) {
	auth := api.Group("/auth")
	auth.POST("/register", c.SignUp)
	auth.POST("/login", c.Login)
	auth.POST("/logout", c.Logout)
	auth.GET("/me", c.Me, c.RequireAuth)
	auth.POST("/change-password", c.ChangePassword, c.RequireAuth)
}

func (c *AuthController) SignUp(ctx *echo.Context) error {
	var req dto.RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	req.Phone = strings.TrimSpace(req.Phone)
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	user, err := c.auth.Register(ctx.Request().Context(), req)
	if err != nil {
		return writeAuthError(ctx, err)
	}
	return response.Created(ctx, user)
}

func (c *AuthController) Login(ctx *echo.Context) error {
	var req dto.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	req.Email = strings.TrimSpace(req.Email)
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	session, err := c.auth.Login(ctx.Request().Context(), req)
	if err != nil {
		return writeAuthError(ctx, err)
	}

	c.setSessionCookie(ctx, session)
	return response.OK(ctx, session)
}

func (c *AuthController) Logout(ctx *echo.Context) error {
	c.clearSessionCookie(ctx)
	return response.OK(ctx, map[string]string{"message": "logged out"})
}

func (c *AuthController) Me(ctx *echo.Context) error {
	user, err := currentAuthUser(ctx)
	if err != nil {
		return response.Unauthorized(ctx, err)
	}
	return response.OK(ctx, user)
}

func (c *AuthController) ChangePassword(ctx *echo.Context) error {
	user, err := currentAuthUser(ctx)
	if err != nil {
		return response.Unauthorized(ctx, err)
	}

	var req dto.ChangePasswordRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	session, err := c.auth.ChangePassword(ctx.Request().Context(), user.ID, req)
	if err != nil {
		return writeAuthError(ctx, err)
	}

	c.setSessionCookie(ctx, session)
	return response.OK(ctx, session)
}

// RequireAuth reads the JWT from the configured auth cookie and loads the user.
func (c *AuthController) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		raw, err := c.tokenFromCookie(ctx)
		if err != nil {
			return response.Unauthorized(ctx, err)
		}

		user, err := c.auth.Authenticate(ctx.Request().Context(), raw)
		if err != nil {
			return writeAuthError(ctx, err)
		}

		ctx.Set(authUserKey, user)
		return next(ctx)
	}
}

func (c *AuthController) tokenFromCookie(ctx *echo.Context) (string, error) {
	cookie, err := ctx.Cookie(c.cfg.CookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", service.ErrUnauthenticated
	}
	return cookie.Value, nil
}

func (c *AuthController) setSessionCookie(ctx *echo.Context, session dto.Session) {
	maxAge := int(time.Until(session.ExpiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	c.writeCookie(ctx, session.Token, session.ExpiresAt.UTC(), maxAge)
}

func (c *AuthController) clearSessionCookie(ctx *echo.Context) {
	c.writeCookie(ctx, "", time.Unix(0, 0).UTC(), -1)
}

func (c *AuthController) writeCookie(ctx *echo.Context, value string, expires time.Time, maxAge int) {
	cookie := &http.Cookie{
		Name:     c.cfg.CookieName,
		Value:    value,
		Path:     c.cfg.CookiePath,
		Expires:  expires,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   c.cfg.CookieSecure,
		SameSite: sameSiteMode(c.cfg.CookieSameSite),
	}
	if c.cfg.CookieDomain != "" {
		cookie.Domain = c.cfg.CookieDomain
	}
	ctx.SetCookie(cookie)
}

func currentAuthUser(ctx *echo.Context) (dto.AuthUser, error) {
	user, ok := ctx.Get(authUserKey).(dto.AuthUser)
	if !ok || user.ID == "" {
		return dto.AuthUser{}, service.ErrUnauthenticated
	}
	return user, nil
}

func writeAuthError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrInvalidToken),
		errors.Is(err, service.ErrTokenExpired),
		errors.Is(err, service.ErrUnauthenticated):
		return response.Unauthorized(ctx, err)
	case errors.Is(err, service.ErrAccountDisabled):
		return response.Forbidden(ctx, err)
	case errors.Is(err, service.ErrEmailTaken),
		errors.Is(err, service.ErrUsernameTaken):
		return response.Conflict(ctx, err)
	case errors.Is(err, service.ErrWrongPassword),
		errors.Is(err, service.ErrPasswordUnchanged),
		errors.Is(err, service.ErrInvalidPassword),
		errors.Is(err, service.ErrInvalidInput):
		return response.BadRequest(ctx, err)
	default:
		return response.InternalServerError(ctx, err)
	}
}

func sameSiteMode(value string) http.SameSite {
	switch value {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
