package controller

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"test/internal/config"
	"test/internal/service"
	"test/internal/sso"
	"test/pkg/response"

	"github.com/labstack/echo/v5"
)

type SSOController struct {
	sso     sso.Service
	authAPI service.AuthService
	guard   *AuthController
	cfg     *config.Config
}

func NewSSOController(svc sso.Service, authAPI service.AuthService, guard *AuthController, cfg *config.Config) *SSOController {
	return &SSOController{sso: svc, authAPI: authAPI, guard: guard, cfg: cfg}
}

func (c *SSOController) Register(api *echo.Group) {
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
//	@Success		200	{object}	sso.DiscoveryDocument
//	@Router			/api/sso/.well-known/openid-configuration [get]
func (c *SSOController) Discovery(ctx *echo.Context) error {
	base := c.publicBaseURL(ctx)
	return ctx.JSON(http.StatusOK, c.sso.Discovery(base))
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
func (c *SSOController) Authorize(ctx *echo.Context) error {
	var q sso.AuthorizeQuery
	if err := ctx.Bind(&q); err != nil {
		return response.BadRequest(ctx, err)
	}
	q.ClientID = strings.TrimSpace(q.ClientID)
	q.RedirectURI = strings.TrimSpace(q.RedirectURI)
	q.ResponseType = strings.TrimSpace(q.ResponseType)
	if err := ctx.Validate(&q); err != nil {
		return response.BadRequest(ctx, err)
	}

	raw, err := c.guard.tokenFromRequest(ctx)
	if err != nil {
		return c.loginRequired(ctx, q)
	}
	user, err := c.authAPI.Authenticate(ctx.Request().Context(), raw)
	if err != nil {
		return c.loginRequired(ctx, q)
	}

	redirectTo, err := c.sso.Authorize(ctx.Request().Context(), user, q)
	if err != nil {
		return writeSSOError(ctx, err)
	}
	return ctx.Redirect(http.StatusFound, redirectTo)
}

// Token exchanges an authorization code for an SSO access token.
//
//	@Summary		SSO token exchange
//	@Tags			sso
//	@Accept			json
//	@Produce		json
//	@Param			body	body		sso.TokenRequest	true	"authorization_code grant"
//	@Success		200		{object}	sso.TokenResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/sso/token [post]
func (c *SSOController) Token(ctx *echo.Context) error {
	var req sso.TokenRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	token, err := c.sso.ExchangeToken(ctx.Request().Context(), req)
	if err != nil {
		return writeSSOError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, token)
}

// UserInfo returns the authenticated SSO user profile.
//
//	@Summary		SSO userinfo
//	@Tags			sso
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	sso.UserInfoResponse
//	@Failure		401	{object}	response.APIResponse
//	@Router			/api/sso/userinfo [get]
func (c *SSOController) UserInfo(ctx *echo.Context) error {
	raw := bearerToken(ctx.Request().Header.Get("Authorization"))
	if raw == "" {
		return response.Unauthorized(ctx, sso.ErrLoginRequired)
	}
	info, err := c.sso.UserInfo(ctx.Request().Context(), raw)
	if err != nil {
		return writeSSOError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, info)
}

// Introspect validates an SSO access token for a confidential client.
//
//	@Summary		SSO token introspection
//	@Tags			sso
//	@Accept			json
//	@Produce		json
//	@Param			body	body		sso.IntrospectRequest	true	"token + client credentials"
//	@Success		200		{object}	sso.IntrospectResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/sso/introspect [post]
func (c *SSOController) Introspect(ctx *echo.Context) error {
	var req sso.IntrospectRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.sso.Introspect(ctx.Request().Context(), req)
	if err != nil {
		return writeSSOError(ctx, err)
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
//	@Param			body	body		sso.SessionExchangeRequest	true	"client credentials"
//	@Success		200		{object}	sso.SessionExchangeResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/sso/session [post]
func (c *SSOController) ExchangeSession(ctx *echo.Context) error {
	user, err := currentAuthUser(ctx)
	if err != nil {
		return response.Unauthorized(ctx, err)
	}

	var req sso.SessionExchangeRequest
	if err := ctx.Bind(&req); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&req); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.sso.ExchangeSession(ctx.Request().Context(), user, req)
	if err != nil {
		return writeSSOError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, result)
}

func (c *SSOController) loginRequired(ctx *echo.Context, q sso.AuthorizeQuery) error {
	loginURL := strings.TrimSpace(c.cfg.SSO.LoginURL)
	if loginURL == "" {
		return response.Unauthorized(ctx, sso.ErrLoginRequired)
	}

	returnTo := c.publicBaseURL(ctx) + strings.TrimRight(c.cfg.App.Prefix, "/") + "/sso/authorize?" + ctx.Request().URL.RawQuery
	u, err := url.Parse(loginURL)
	if err != nil {
		return response.Unauthorized(ctx, sso.ErrLoginRequired)
	}
	query := u.Query()
	query.Set("return_to", returnTo)
	if q.State != "" {
		query.Set("state", q.State)
	}
	u.RawQuery = query.Encode()
	return ctx.Redirect(http.StatusFound, u.String())
}

func (c *SSOController) publicBaseURL(ctx *echo.Context) string {
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

func writeSSOError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, sso.ErrLoginRequired),
		errors.Is(err, sso.ErrInvalidClient),
		errors.Is(err, sso.ErrInvalidGrant),
		errors.Is(err, sso.ErrCodeExpired),
		errors.Is(err, sso.ErrCodeUsed),
		errors.Is(err, sso.ErrPKCEInvalid):
		return response.Unauthorized(ctx, err)
	case errors.Is(err, sso.ErrInactiveClient):
		return response.Forbidden(ctx, err)
	case errors.Is(err, sso.ErrInvalidRedirectURI),
		errors.Is(err, sso.ErrPKCERequired),
		errors.Is(err, sso.ErrUnsupportedType),
		errors.Is(err, sso.ErrInvalidScope),
		errors.Is(err, sso.ErrInvalidInput):
		return response.BadRequest(ctx, err)
	default:
		return response.InternalServerError(ctx, err)
	}
}
