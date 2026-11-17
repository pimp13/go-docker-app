package app

import (
	"net/http"
	"net/url"
	"strings"

	"test/internal/config"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func corsMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	origins := cfg.App.CorsOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:3001"}
	}

	allowOrigin := corsAllowOriginFunc(cfg.App.IsProduction, origins)

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		// Reflect Access-Control-Request-Headers when empty — browsers often
		// send more than Content-Type (e.g. Authorization on later calls).
		AllowHeaders: nil,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowCredentials:      true,
		MaxAge:                600,
		UnsafeAllowOriginFunc: allowOrigin,
	})
}

func corsAllowOriginFunc(
	isProduction bool,
	configured []string,
) func(c *echo.Context, origin string) (string, bool, error) {
	allowed := make(map[string]struct{}, len(configured))
	for _, origin := range configured {
		key := strings.ToLower(strings.TrimRight(strings.TrimSpace(origin), "/"))
		if key != "" {
			allowed[key] = struct{}{}
		}
	}

	return func(_ *echo.Context, origin string) (string, bool, error) {
		normalized := strings.TrimRight(strings.TrimSpace(origin), "/")
		if normalized == "" {
			return "", false, nil
		}

		if _, ok := allowed[strings.ToLower(normalized)]; ok {
			return normalized, true, nil
		}

		// Local Next.js may be opened as localhost, 127.0.0.1, or a LAN IP.
		if !isProduction && isDevBrowserOrigin(normalized) {
			return normalized, true, nil
		}

		return "", false, nil
	}
}

func isDevBrowserOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	// Origins must not carry a path/query/fragment.
	if (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	switch {
	case host == "localhost", host == "127.0.0.1", host == "::1":
		return true
	case isPrivateIPv4(host):
		return true
	default:
		return false
	}
}

func isPrivateIPv4(host string) bool {
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}

	switch {
	case parts[0] == "10":
		return true
	case parts[0] == "192" && parts[1] == "168":
		return true
	case parts[0] == "172":
		// 172.16.0.0 – 172.31.255.255
		second := 0
		for _, ch := range parts[1] {
			second = second*10 + int(ch-'0')
		}
		return second >= 16 && second <= 31
	default:
		return false
	}
}
