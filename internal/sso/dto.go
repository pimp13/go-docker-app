package sso

import "time"

type AuthorizeQuery struct {
	ResponseType        string `query:"response_type" validate:"required,eq=code"`
	ClientID            string `query:"client_id" validate:"required"`
	RedirectURI         string `query:"redirect_uri" validate:"required,url"`
	State               string `query:"state"`
	Scope               string `query:"scope"`
	CodeChallenge       string `query:"code_challenge"`
	CodeChallengeMethod string `query:"code_challenge_method" validate:"omitempty,oneof=S256 plain"`
}

type TokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type" validate:"required,eq=authorization_code"`
	Code         string `json:"code" form:"code" validate:"required"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri" validate:"required,url"`
	ClientID     string `json:"client_id" form:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	CodeVerifier string `json:"code_verifier" form:"code_verifier"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

type UserInfoResponse struct {
	Sub         string   `json:"sub"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	GivenName   string   `json:"given_name,omitempty"`
	FamilyName  string   `json:"family_name,omitempty"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status"`
}

type IntrospectRequest struct {
	Token        string `json:"token" form:"token" validate:"required"`
	ClientID     string `json:"client_id" form:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" form:"client_secret" validate:"required"`
}

type IntrospectResponse struct {
	Active      bool     `json:"active"`
	Sub         string   `json:"sub,omitempty"`
	ClientID    string   `json:"client_id,omitempty"`
	Email       string   `json:"email,omitempty"`
	Role        string   `json:"role,omitempty"`
	Scope       string   `json:"scope,omitempty"`
	TokenType   string   `json:"token_type,omitempty"`
	Exp         int64    `json:"exp,omitempty"`
	Iat         int64    `json:"iat,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type SessionExchangeRequest struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

type SessionExchangeResponse struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	ExpiresIn   int64            `json:"expires_in"`
	ExpiresAt   time.Time        `json:"expires_at"`
	User        UserInfoResponse `json:"user"`
}

type DiscoveryDocument struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
}
