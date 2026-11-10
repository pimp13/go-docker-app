package sso

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"test/internal/config"
	"test/pkg/response"

	"github.com/labstack/echo/v5"
)

type Controller struct {
	service Service
	authAPI Authenticator
	guard   Guard
	cfg     *config.Config
}

func NewController(service Service, authAPI Authenticator, guard Guard, cfg *config.Config) *Controller {
	return &Controller{service: service, authAPI: authAPI, guard: guard, cfg: cfg}
}

func (c *Controller) Register(api *echo.Group) {
	group := api.Group("/sso")
	group.GET("/.well-known/openid-configuration", c.Discovery)
	group.GET("/authorize", c.Authorize)
	group.POST("/token", c.Token)
	group.GET("/userinfo", c.UserInfo)
	group.POST("/introspect", c.Introspect)
	group.POST("/session", c.ExchangeSession, c.guard.RequireAuth)
}

// Discovery returns OpenID-style metadata for this SSO issuer.
//
//	@Summary		SSO discovery document
//	@Tags			sso
//	@Produce		json
//	@Success		200	{object}	DiscoveryDocument
//	@Router			/api/sso/.well-known/openid-configuration [get]
func (c *Controller) Discovery(ctx *echo.Context) error {
	base := c.publicBaseURL(ctx)
	return ctx.JSON(http.StatusOK, c.service.Discovery(base))
}

// Authorize starts the OAuth2 authorization-code flow using the current session cookie.
//
//	@Summary		SSO authorize
//	@Description	Requires an authenticated browser session cookie. Redirects back with ?code=&state=
//	@Tags			sso
//	@Produce		json
//	@Param			response_type			query		string	true	"Must be code"
//	@Param			client_id				query		string	true	"Registered client id"
//	@Param			redirect_uri			query		string	true	"Allowed redirect URI"
//	@Param			state					query		string	false	"Opaque CSRF state"
//	@Param			scope					query		string	false	"openid profile email"
//	@Param			code_challenge			query		string	false	"PKCE challenge"
//	@Param			code_challenge_method	query		string	false	"S256 or plain"
//	@Success		302	{string}	string
//	@Failure		400	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Router			/api/sso/authorize [get]
func (c *Controller) Authorize(ctx *echo.Context) error {
	var q AuthorizeQuery
	if err := ctx.Bind(&q); err != nil {
		return response.BadRequest(ctx, err)
	}
	q.ClientID = strings.TrimSpace(q.ClientID)
	q.RedirectURI = strings.TrimSpace(q.RedirectURI)
	q.ResponseType = strings.TrimSpace(q.ResponseType)
	if err := ctx.Validate(&q); err != nil {
		return response.BadRequest(ctx, err)
	}

	raw, err := c.guard.TokenFromRequest(ctx)
	if err != nil {
		return c.loginRequired(ctx, q)
	}
	user, err := c.authAPI.Authenticate(ctx.Request().Context(), raw)
	if err != nil {
		return c.loginRequired(ctx, q)
	}

	redirectTo, err := c.service.Authorize(ctx.Request().Context(), user, q)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.Redirect(http.StatusFound, redirectTo)
}

// Token exchanges an authorization code for an SSO access token.
//
//	@Summary		SSO token exchange
//	@Tags			sso
//	@Accept			json
//	@Produce		json
//	@Param			body	body		TokenRequest	true	"authorization_code grant"
//	@Success		200		{object}	TokenResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/sso/token [post]
func (c *Controller) Token(ctx *echo.Context) error {
	var req TokenRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	token, err := c.service.ExchangeToken(ctx.Request().Context(), req)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, token)
}

// UserInfo returns the authenticated SSO user profile.
//
//	@Summary		SSO userinfo
//	@Tags			sso
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UserInfoResponse
//	@Failure		401	{object}	response.APIResponse
//	@Router			/api/sso/userinfo [get]
func (c *Controller) UserInfo(ctx *echo.Context) error {
	raw := bearerToken(ctx.Request().Header.Get("Authorization"))
	if raw == "" {
		return response.Unauthorized(ctx, ErrLoginRequired)
	}
	info, err := c.service.UserInfo(ctx.Request().Context(), raw)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, info)
}

// Introspect validates an SSO access token for a confidential client.
//
//	@Summary		SSO token introspection
//	@Tags			sso
//	@Accept			json
//	@Produce		json
//	@Param			body	body		IntrospectRequest	true	"token + client credentials"
//	@Success		200		{object}	IntrospectResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/sso/introspect [post]
func (c *Controller) Introspect(ctx *echo.Context) error {
	var req IntrospectRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.service.Introspect(ctx.Request().Context(), req)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, result)
}

// ExchangeSession converts the current session cookie into an SSO access token for a client app.
//
//	@Summary		Exchange session for SSO token
//	@Description	Use when the user is already logged into this IdP and a trusted client wants an access token without browser redirect.
//	@Tags			sso
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			body	body		SessionExchangeRequest	true	"client credentials"
//	@Success		200		{object}	SessionExchangeResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/sso/session [post]
func (c *Controller) ExchangeSession(ctx *echo.Context) error {
	user, err := c.guard.CurrentUser(ctx)
	if err != nil {
		return response.Unauthorized(ctx, err)
	}

	var req SessionExchangeRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.service.ExchangeSession(ctx.Request().Context(), user, req)
	if err != nil {
		return writeError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, result)
}

func (c *Controller) loginRequired(ctx *echo.Context, q AuthorizeQuery) error {
	loginURL := strings.TrimSpace(c.cfg.SSO.LoginURL)
	if loginURL == "" {
		return response.Unauthorized(ctx, ErrLoginRequired)
	}

	returnTo := c.publicBaseURL(ctx) + strings.TrimRight(c.cfg.App.Prefix, "/") + "/sso/authorize?" + ctx.Request().URL.RawQuery
	u, err := url.Parse(loginURL)
	if err != nil {
		return response.Unauthorized(ctx, ErrLoginRequired)
	}
	query := u.Query()
	query.Set("return_to", returnTo)
	if q.State != "" {
		query.Set("state", q.State)
	}
	u.RawQuery = query.Encode()
	return ctx.Redirect(http.StatusFound, u.String())
}

func (c *Controller) publicBaseURL(ctx *echo.Context) string {
	if base := strings.TrimSpace(c.cfg.SSO.PublicBaseURL); base != "" {
		return strings.TrimRight(base, "/")
	}
	scheme := "http"
	if ctx.Request().TLS != nil {
		scheme = "https"
	}
	if proto := ctx.Request().Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + ctx.Request().Host
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

func writeError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrLoginRequired),
		errors.Is(err, ErrInvalidClient),
		errors.Is(err, ErrInvalidGrant),
		errors.Is(err, ErrCodeExpired),
		errors.Is(err, ErrCodeUsed),
		errors.Is(err, ErrPKCEInvalid):
		return response.Unauthorized(ctx, err)
	case errors.Is(err, ErrInactiveClient):
		return response.Forbidden(ctx, err)
	case errors.Is(err, ErrInvalidRedirectURI),
		errors.Is(err, ErrPKCERequired),
		errors.Is(err, ErrUnsupportedType),
		errors.Is(err, ErrInvalidScope),
		errors.Is(err, ErrInvalidInput):
		return response.BadRequest(ctx, err)
	default:
		return response.InternalServerError(ctx, err)
	}
}
