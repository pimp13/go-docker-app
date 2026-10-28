package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"test/internal/config"
	"test/internal/controller"
	"test/internal/dto"
	"test/internal/service"
	appvalidator "test/pkg/validator"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

const testCookieName = "session_jwt"

type stubAuth struct {
	session dto.Session
	err     error
}

func (s *stubAuth) Register(context.Context, dto.RegisterRequest) (dto.AuthUser, error) {
	return s.session.User, s.err
}

func (s *stubAuth) Login(context.Context, dto.LoginRequest) (dto.Session, error) {
	return s.session, s.err
}

func (s *stubAuth) Authenticate(context.Context, string) (dto.AuthUser, error) {
	if s.err != nil {
		return dto.AuthUser{}, s.err
	}
	return s.session.User, nil
}

func (s *stubAuth) ChangePassword(context.Context, string, dto.ChangePasswordRequest) (dto.Session, error) {
	return s.session, s.err
}

func newAuthEngine(t *testing.T, auth service.AuthService) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.Validator = appvalidator.New()
	controller.NewAuthController(auth, config.AuthConfig{
		CookieName:     testCookieName,
		CookiePath:     "/",
		CookieSameSite: "lax",
		TokenTTL:       time.Hour,
	}).Register(e.Group("/api"))
	return e
}

func TestLoginSetsConfiguredAuthCookie(t *testing.T) {
	const token = "header.payload.signature"
	expiresAt := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	engine := newAuthEngine(t, &stubAuth{
		session: dto.Session{
			Token:     token,
			ExpiresAt: expiresAt,
			User: dto.AuthUser{
				ID:        "user-1",
				FirstName: "Ada",
				LastName:  "Lovelace",
				Email:     "ada@example.com",
				Role:      "user",
				Status:    "active",
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"ada@example.com","password":"password1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), token) {
		t.Fatalf("token leaked into JSON body: %s", rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %#v", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != testCookieName || cookie.Value != token {
		t.Fatalf("cookie = %#v", cookie)
	}
	if !cookie.HttpOnly || cookie.Secure || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookie flags = %#v", cookie)
	}
	if cookie.MaxAge < 1 {
		t.Fatalf("max-age = %d", cookie.MaxAge)
	}
}

func TestLoginFailureDoesNotSetCookie(t *testing.T) {
	engine := newAuthEngine(t, &stubAuth{err: service.ErrInvalidCredentials})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"ada@example.com","password":"password1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("unexpected cookies: %#v", rec.Result().Cookies())
	}
}

func TestLogoutClearsAuthCookie(t *testing.T) {
	engine := newAuthEngine(t, &stubAuth{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "header.payload.signature"})
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %#v", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != testCookieName || cookie.Value != "" || cookie.MaxAge >= 0 || !cookie.HttpOnly {
		t.Fatalf("cookie = %#v", cookie)
	}
}

func TestMeRequiresAuthCookie(t *testing.T) {
	engine := newAuthEngine(t, &stubAuth{
		session: dto.Session{User: dto.AuthUser{ID: "user-1", Email: "ada@example.com"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "header.payload.signature"})
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "ada@example.com") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestChangePasswordRefreshesAuthCookie(t *testing.T) {
	const token = "new.header.payload"
	engine := newAuthEngine(t, &stubAuth{
		session: dto.Session{
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour),
			User:      dto.AuthUser{ID: "user-1", Email: "ada@example.com", Role: "user", Status: "active"},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", strings.NewReader(`{"current_password":"password1","new_password":"password2"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: testCookieName, Value: "old.token.value"})
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != testCookieName || cookies[0].Value != token || !cookies[0].HttpOnly {
		t.Fatalf("cookies = %#v", cookies)
	}
	if strings.Contains(rec.Body.String(), token) {
		t.Fatalf("token leaked into JSON body: %s", rec.Body.String())
	}
}
