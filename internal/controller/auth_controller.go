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
	auth.POST("/sign-up", c.SignUp)
	auth.POST("/login", c.Login)
	auth.POST("/logout", c.Logout)
	auth.GET("/me", c.Me, c.RequireAuth)
	auth.POST("/change-password", c.ChangePassword, c.RequireAuth)
}

func (c *AuthController) trimInputs(req *dto.RegisterRequest) *dto.RegisterRequest {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)
	req.Phone = strings.TrimSpace(req.Phone)

	return req
}

// SignUp registers a new user.
//
//	@Summary		Register a new user
//	@Description	Creates a user with the default role
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterRequest	true	"Registration payload"
//	@Success		201		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		409		{object}	response.APIResponse
//	@Router			/api/auth/sign-up [post]
func (c *AuthController) SignUp(ctx *echo.Context) error {
	var req dto.RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	c.trimInputs(&req)

	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	user, err := c.auth.Register(ctx.Request().Context(), req)
	if err != nil {
		return writeAuthError(ctx, err)
	}
	return response.Created(ctx, user)
}

// Login authenticates a user and sets the JWT cookie.
//
//	@Summary		Login
//	@Description	Authenticates the user and sets the JWT auth cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginRequest	true	"Login payload"
//	@Success		200		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Header			200		{string}	Set-Cookie	"auth_token cookie containing JWT"
//	@Router			/api/auth/login [post]
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

// Logout clears the auth cookie.
//
//	@Summary		Logout
//	@Description	Clears the auth cookie
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	response.APIResponse
//	@Router			/api/auth/logout [post]
func (c *AuthController) Logout(ctx *echo.Context) error {
	c.clearSessionCookie(ctx)
	return response.OK(ctx, map[string]string{"message": "logged out"})
}

// Me returns the authenticated user.
//
//	@Summary		Current user
//	@Description	Returns the authenticated user profile and permissions
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Success		200	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Router			/api/auth/me [get]
func (c *AuthController) Me(ctx *echo.Context) error {
	user, err := currentAuthUser(ctx)
	if err != nil {
		return response.Unauthorized(ctx, err)
	}
	return response.OK(ctx, user)
}

// ChangePassword updates the password and refreshes the session cookie.
//
//	@Summary		Change password
//	@Description	Changes password and refreshes the auth cookie/token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			body	body		dto.ChangePasswordRequest	true	"Password change payload"
//	@Success		200		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/auth/change-password [post]
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

// RequireAuth reads the JWT from the auth cookie, then from Authorization: Bearer.
func (c *AuthController) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		raw, err := c.tokenFromRequest(ctx)
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

// RequirePermission ensures the authenticated user has every listed permission.
// Always register after RequireAuth.
func (c *AuthController) RequirePermission(codes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			user, err := currentAuthUser(ctx)
			if err != nil {
				return response.Unauthorized(ctx, err)
			}
			if !user.HasAllPermissions(codes...) {
				return response.Forbidden(ctx, service.ErrPermissionDenied)
			}
			return next(ctx)
		}
	}
}

func (c *AuthController) tokenFromRequest(ctx *echo.Context) (string, error) {
	if cookie, err := ctx.Cookie(c.cfg.CookieName); err == nil {
		if token := strings.TrimSpace(cookie.Value); token != "" {
			return token, nil
		}
	}

	if token := bearerToken(ctx.Request().Header.Get("Authorization")); token != "" {
		return token, nil
	}

	return "", service.ErrUnauthenticated
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
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
	case errors.Is(err, service.ErrAccountDisabled),
		errors.Is(err, service.ErrPermissionDenied):
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
