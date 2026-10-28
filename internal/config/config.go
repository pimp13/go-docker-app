package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config is the typed application configuration loaded from the environment.
type Config struct {
	App  AppConfig
	DB   DatabaseConfig
	Auth AuthConfig
}

// AppConfig holds HTTP server and application runtime settings.
type AppConfig struct {
	Port             uint
	Prefix           string
	Version          string
	IsProduction     bool
	HealthCheckRoute string
	ShutdownTimeout  time.Duration
	SecretKey        string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     uint
	User     string
	Password string
	Name     string
	SSLMode  string
}

// AuthConfig holds JWT signing settings and the auth cookie.
type AuthConfig struct {
	CookieName     string
	CookiePath     string
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite string
	JWTSecret      string
	JWTIssuer      string
	TokenTTL       time.Duration
}

// DSN builds a PostgreSQL connection string for GORM / pgx.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.Name,
		d.SSLMode,
	)
}

// Address returns the HTTP listen address, e.g. ":8080".
func (a AppConfig) Address() string {
	return fmt.Sprintf(":%d", a.Port)
}

// Load reads configuration from environment variables.
// If a .env file exists in the working directory it is loaded first
// (existing OS env values always win).
func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := envUint("APP_PORT", 8080)
	if err != nil {
		return nil, err
	}

	isProduction, err := envBool("APP_IS_PRODUCTION", false)
	if err != nil {
		return nil, err
	}

	shutdownTimeout, err := envDuration("APP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}

	dbPort, err := envUint("DB_PORT", 5432)
	if err != nil {
		return nil, err
	}

	dbPassword, err := envStringRequired("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	secretKey, err := envStringRequired("SECRET_KEY")
	if err != nil {
		return nil, err
	}

	authCookieName, err := envStringRequired("AUTH_COOKIE_NAME")
	if err != nil {
		return nil, err
	}

	tokenTTL, err := envDuration("AUTH_JWT_TTL", 24*time.Hour)
	if err != nil {
		return nil, err
	}

	cookieSecure, err := envBool("AUTH_COOKIE_SECURE", isProduction)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Port:             port,
			Prefix:           envString("APP_PREFIX", "/api"),
			Version:          envString("APP_VERSION", "0.1.0"),
			IsProduction:     isProduction,
			HealthCheckRoute: envString("APP_HEALTH_CHECK_ROUTE", "/health"),
			ShutdownTimeout:  shutdownTimeout,
			SecretKey:        secretKey,
		},
		DB: DatabaseConfig{
			Host:     envString("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     envString("DB_USER", "appuser"),
			Password: dbPassword,
			Name:     envString("DB_NAME", "appdb"),
			SSLMode:  envString("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			CookieName:     authCookieName,
			CookiePath:     envString("AUTH_COOKIE_PATH", "/"),
			CookieDomain:   envString("AUTH_COOKIE_DOMAIN", ""),
			CookieSecure:   cookieSecure,
			CookieSameSite: strings.ToLower(strings.TrimSpace(envString("AUTH_COOKIE_SAME_SITE", "lax"))),
			JWTSecret:      envString("AUTH_JWT_SECRET", secretKey),
			JWTIssuer:      envString("AUTH_JWT_ISSUER", "test"),
			TokenTTL:       tokenTTL,
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.App.Port == 0 || c.App.Port > 65535 {
		return fmt.Errorf("APP_PORT must be between 1 and 65535")
	}
	if c.App.ShutdownTimeout <= 0 {
		return fmt.Errorf("APP_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if c.DB.Host == "" {
		return fmt.Errorf("DB_HOST must not be empty")
	}
	if c.DB.Port == 0 || c.DB.Port > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535")
	}
	if c.DB.User == "" {
		return fmt.Errorf("DB_USER must not be empty")
	}
	if c.DB.Name == "" {
		return fmt.Errorf("DB_NAME must not be empty")
	}
	if c.App.SecretKey == "" {
		return fmt.Errorf("SECRET_KEY must not be empty")
	}
	if strings.TrimSpace(c.Auth.CookieName) == "" {
		return fmt.Errorf("AUTH_COOKIE_NAME must not be empty")
	}
	if strings.ContainsAny(c.Auth.CookieName, " ;\t,") {
		return fmt.Errorf("AUTH_COOKIE_NAME contains invalid characters")
	}
	if c.Auth.CookiePath == "" || !strings.HasPrefix(c.Auth.CookiePath, "/") {
		return fmt.Errorf("AUTH_COOKIE_PATH must start with /")
	}
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("AUTH_JWT_SECRET must be at least 32 characters")
	}
	if strings.TrimSpace(c.Auth.JWTIssuer) == "" {
		return fmt.Errorf("AUTH_JWT_ISSUER must not be empty")
	}
	if c.Auth.TokenTTL <= 0 {
		return fmt.Errorf("AUTH_JWT_TTL must be greater than zero")
	}
	switch c.Auth.CookieSameSite {
	case "lax", "strict", "none":
	default:
		return fmt.Errorf("AUTH_COOKIE_SAME_SITE must be one of lax, strict, none")
	}
	if c.Auth.CookieSameSite == "none" && !c.Auth.CookieSecure {
		return fmt.Errorf("AUTH_COOKIE_SECURE must be true when AUTH_COOKIE_SAME_SITE is none")
	}
	if c.App.IsProduction && !c.Auth.CookieSecure {
		return fmt.Errorf("AUTH_COOKIE_SECURE must be true when APP_IS_PRODUCTION is true")
	}
	return nil
}
