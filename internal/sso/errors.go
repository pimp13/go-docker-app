package sso

import "errors"

var (
	ErrInvalidClient      = errors.New("invalid client credentials")
	ErrInactiveClient     = errors.New("client is inactive")
	ErrInvalidRedirectURI = errors.New("invalid redirect_uri")
	ErrInvalidGrant       = errors.New("invalid authorization grant")
	ErrCodeExpired        = errors.New("authorization code expired")
	ErrCodeUsed           = errors.New("authorization code already used")
	ErrPKCERequired       = errors.New("pkce code_challenge is required")
	ErrPKCEInvalid        = errors.New("pkce verification failed")
	ErrUnsupportedType    = errors.New("unsupported response or grant type")
	ErrLoginRequired      = errors.New("authentication required")
	ErrInvalidScope       = errors.New("invalid scope")
	ErrInvalidInput       = errors.New("invalid input")
)
