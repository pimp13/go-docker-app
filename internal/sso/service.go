package sso

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"test/internal/config"
	"test/internal/dto"
	"test/internal/model"
	"test/pkg/jwtutil"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	defaultScope = "openid profile email"
)

type Service interface {
	Discovery(baseURL string) DiscoveryDocument
	Authorize(ctx context.Context, user dto.AuthUser, q AuthorizeQuery) (string, error)
	ExchangeToken(ctx context.Context, req TokenRequest) (TokenResponse, error)
	UserInfo(ctx context.Context, accessToken string) (UserInfoResponse, error)
	Introspect(ctx context.Context, req IntrospectRequest) (IntrospectResponse, error)
	ExchangeSession(ctx context.Context, user dto.AuthUser, req SessionExchangeRequest) (SessionExchangeResponse, error)
}

type service struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewService(db *gorm.DB, cfg *config.Config) Service {
	return &service{db: db, cfg: cfg}
}

func (s *service) Discovery(baseURL string) DiscoveryDocument {
	baseURL = strings.TrimRight(baseURL, "/")
	prefix := strings.TrimRight(s.cfg.App.Prefix, "/")
	return DiscoveryDocument{
		Issuer:                            s.cfg.SSO.Issuer,
		AuthorizationEndpoint:             baseURL + prefix + "/sso/authorize",
		TokenEndpoint:                     baseURL + prefix + "/sso/token",
		UserInfoEndpoint:                  baseURL + prefix + "/sso/userinfo",
		IntrospectionEndpoint:             baseURL + prefix + "/sso/introspect",
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post"},
		ScopesSupported:                   []string{"openid", "profile", "email"},
		CodeChallengeMethodsSupported:     []string{"S256", "plain"},
		SubjectTypesSupported:             []string{"public"},
	}
}

func (s *service) Authorize(ctx context.Context, user dto.AuthUser, q AuthorizeQuery) (string, error) {
	if strings.TrimSpace(q.ResponseType) != "code" {
		return "", ErrUnsupportedType
	}
	if user.ID == "" || user.Status != string(model.UserStatusActive) {
		return "", ErrLoginRequired
	}

	client, err := s.findActiveClient(ctx, q.ClientID)
	if err != nil {
		return "", err
	}
	if !clientAllowsRedirect(client, q.RedirectURI) {
		return "", ErrInvalidRedirectURI
	}

	challenge := strings.TrimSpace(q.CodeChallenge)
	method := strings.TrimSpace(q.CodeChallengeMethod)
	if client.RequirePKCE && challenge == "" {
		return "", ErrPKCERequired
	}
	if challenge != "" && method == "" {
		method = "plain"
	}
	if method != "" && method != "S256" && method != "plain" {
		return "", ErrPKCEInvalid
	}

	scopes := normalizeScopes(q.Scope)
	rawCode, err := randomToken(32)
	if err != nil {
		return "", err
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return "", ErrLoginRequired
	}

	row := model.SSOAuthorizationCode{
		ID:                  uuid.New(),
		CodeHash:            hashSecret(rawCode),
		ClientID:            client.ClientID,
		UserID:              userID,
		RedirectURI:         q.RedirectURI,
		Scopes:              scopes,
		CodeChallenge:       challenge,
		CodeChallengeMethod: method,
		ExpiresAt:           time.Now().UTC().Add(s.cfg.SSO.CodeTTL),
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return "", err
	}

	u, err := url.Parse(q.RedirectURI)
	if err != nil {
		return "", ErrInvalidRedirectURI
	}
	query := u.Query()
	query.Set("code", rawCode)
	if strings.TrimSpace(q.State) != "" {
		query.Set("state", q.State)
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (s *service) ExchangeToken(ctx context.Context, req TokenRequest) (TokenResponse, error) {
	if req.GrantType != "authorization_code" {
		return TokenResponse{}, ErrUnsupportedType
	}

	client, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return TokenResponse{}, err
	}
	if !clientAllowsRedirect(client, req.RedirectURI) {
		return TokenResponse{}, ErrInvalidRedirectURI
	}

	var code model.SSOAuthorizationCode
	err = s.db.WithContext(ctx).Where("code_hash = ?", hashSecret(req.Code)).Take(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return TokenResponse{}, ErrInvalidGrant
	}
	if err != nil {
		return TokenResponse{}, err
	}
	if code.ClientID != client.ClientID || code.RedirectURI != req.RedirectURI {
		return TokenResponse{}, ErrInvalidGrant
	}
	if code.UsedAt != nil {
		return TokenResponse{}, ErrCodeUsed
	}
	if time.Now().UTC().After(code.ExpiresAt) {
		return TokenResponse{}, ErrCodeExpired
	}
	if err := verifyPKCE(code, req.CodeVerifier); err != nil {
		return TokenResponse{}, err
	}

	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).Model(&code).Update("used_at", now).Error; err != nil {
		return TokenResponse{}, err
	}

	user, err := s.findUser(ctx, code.UserID)
	if err != nil {
		return TokenResponse{}, err
	}
	if user.Status != model.UserStatusActive {
		return TokenResponse{}, ErrInvalidGrant
	}

	authUser := toAuthUser(user)
	accessToken, expiresAt, err := s.issueAccessToken(client.ClientID, authUser)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		Scope:       code.Scopes,
	}, nil
}

func (s *service) UserInfo(ctx context.Context, accessToken string) (UserInfoResponse, error) {
	claims, err := jwtutil.ParseSSO(s.cfg.Auth.JWTSecret, s.cfg.SSO.Issuer, "", accessToken)
	if err != nil {
		if errors.Is(err, jwtutil.ErrExpired) {
			return UserInfoResponse{}, ErrInvalidGrant
		}
		return UserInfoResponse{}, ErrInvalidGrant
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return UserInfoResponse{}, ErrInvalidGrant
	}
	user, err := s.findUser(ctx, userID)
	if err != nil {
		return UserInfoResponse{}, ErrInvalidGrant
	}
	if user.Status != model.UserStatusActive {
		return UserInfoResponse{}, ErrInvalidGrant
	}
	return toUserInfo(toAuthUser(user)), nil
}

func (s *service) Introspect(ctx context.Context, req IntrospectRequest) (IntrospectResponse, error) {
	if _, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret); err != nil {
		return IntrospectResponse{}, err
	}

	claims, err := jwtutil.ParseSSO(s.cfg.Auth.JWTSecret, s.cfg.SSO.Issuer, req.ClientID, req.Token)
	if err != nil {
		return IntrospectResponse{Active: false}, nil
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return IntrospectResponse{Active: false}, nil
	}
	user, err := s.findUser(ctx, userID)
	if err != nil || user.Status != model.UserStatusActive {
		return IntrospectResponse{Active: false}, nil
	}

	var exp, iat int64
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Unix()
	}
	if claims.IssuedAt != nil {
		iat = claims.IssuedAt.Unix()
	}

	return IntrospectResponse{
		Active:      true,
		Sub:         claims.Subject,
		ClientID:    req.ClientID,
		Email:       claims.Email,
		Role:        claims.Role,
		Scope:       defaultScope,
		TokenType:   "Bearer",
		Exp:         exp,
		Iat:         iat,
		Permissions: claims.Permissions,
	}, nil
}

func (s *service) ExchangeSession(ctx context.Context, user dto.AuthUser, req SessionExchangeRequest) (SessionExchangeResponse, error) {
	if user.ID == "" || user.Status != string(model.UserStatusActive) {
		return SessionExchangeResponse{}, ErrLoginRequired
	}
	client, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return SessionExchangeResponse{}, err
	}

	accessToken, expiresAt, err := s.issueAccessToken(client.ClientID, user)
	if err != nil {
		return SessionExchangeResponse{}, err
	}

	return SessionExchangeResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		ExpiresAt:   expiresAt,
		User:        toUserInfo(user),
	}, nil
}

func (s *service) issueAccessToken(clientID string, user dto.AuthUser) (string, time.Time, error) {
	return jwtutil.IssueSSO(jwtutil.SSOIssueParams{
		Secret:      s.cfg.Auth.JWTSecret,
		Issuer:      s.cfg.SSO.Issuer,
		Audience:    clientID,
		Subject:     user.ID,
		Email:       user.Email,
		Name:        strings.TrimSpace(user.FirstName + " " + user.LastName),
		Role:        user.Role,
		Permissions: user.Permissions,
		TTL:         s.cfg.SSO.AccessTokenTTL,
	})
}

func (s *service) authenticateClient(ctx context.Context, clientID, clientSecret string) (model.SSOClient, error) {
	client, err := s.findActiveClient(ctx, clientID)
	if err != nil {
		return model.SSOClient{}, err
	}
	if !secureEqual(client.ClientSecretHash, hashSecret(clientSecret)) {
		return model.SSOClient{}, ErrInvalidClient
	}
	return client, nil
}

func (s *service) findActiveClient(ctx context.Context, clientID string) (model.SSOClient, error) {
	var client model.SSOClient
	err := s.db.WithContext(ctx).Where("client_id = ?", strings.TrimSpace(clientID)).Take(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SSOClient{}, ErrInvalidClient
	}
	if err != nil {
		return model.SSOClient{}, err
	}
	if !client.IsActive {
		return model.SSOClient{}, ErrInactiveClient
	}
	return client, nil
}

func (s *service) findUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("id = ?", id).
		Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrInvalidGrant
	}
	return user, err
}

func clientAllowsRedirect(client model.SSOClient, redirectURI string) bool {
	var uris []string
	if err := json.Unmarshal(client.RedirectURIs, &uris); err != nil {
		return false
	}
	for _, uri := range uris {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

func verifyPKCE(code model.SSOAuthorizationCode, verifier string) error {
	if code.CodeChallenge == "" {
		return nil
	}
	verifier = strings.TrimSpace(verifier)
	if verifier == "" {
		return ErrPKCEInvalid
	}
	switch code.CodeChallengeMethod {
	case "S256":
		sum := sha256.Sum256([]byte(verifier))
		encoded := base64.RawURLEncoding.EncodeToString(sum[:])
		if !secureEqual(encoded, code.CodeChallenge) {
			return ErrPKCEInvalid
		}
	case "plain", "":
		if !secureEqual(verifier, code.CodeChallenge) {
			return ErrPKCEInvalid
		}
	default:
		return ErrPKCEInvalid
	}
	return nil
}

func normalizeScopes(raw string) string {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) == 0 {
		return defaultScope
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return strings.Join(out, " ")
}

func hashSecret(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func secureEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func toAuthUser(user model.User) dto.AuthUser {
	return dto.AuthUser{
		ID:          user.ID.String(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Username:    user.Username,
		Phone:       user.Phone,
		Role:        user.Role.Slug,
		Permissions: user.Role.PermissionCodes(),
		Status:      string(user.Status),
	}
}

func toUserInfo(user dto.AuthUser) UserInfoResponse {
	return UserInfoResponse{
		Sub:         user.ID,
		Email:       user.Email,
		Name:        strings.TrimSpace(user.FirstName + " " + user.LastName),
		GivenName:   user.FirstName,
		FamilyName:  user.LastName,
		Role:        user.Role,
		Permissions: user.Permissions,
		Status:      user.Status,
	}
}

// EncodeRedirectURIs marshals redirect URIs for persistence.
func EncodeRedirectURIs(uris []string) (datatypes.JSON, error) {
	raw, err := json.Marshal(uris)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(raw), nil
}

// HashClientSecret exposes secret hashing for seeding/admin use.
func HashClientSecret(secret string) string {
	return hashSecret(secret)
}
