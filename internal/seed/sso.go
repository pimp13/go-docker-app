package seed

import (
	"fmt"
	"strings"
	"test/internal/config"
	"test/internal/model"
	"test/internal/sso"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SetupSSO migrates SSO tables and optionally seeds the default client from config.
func SetupSSO(db *gorm.DB, cfg config.SSOConfig) error {
	if err := db.AutoMigrate(
		&model.SSOClient{},
		&model.SSOAuthorizationCode{},
	); err != nil {
		return fmt.Errorf("auto migrate sso: %w", err)
	}

	if strings.TrimSpace(cfg.DefaultClientID) == "" || strings.TrimSpace(cfg.DefaultClientSecret) == "" {
		return nil
	}

	uris, err := sso.EncodeRedirectURIs([]string{cfg.DefaultClientRedirectURI})
	if err != nil {
		return err
	}

	client := model.SSOClient{
		ID:               uuid.New(),
		ClientID:         cfg.DefaultClientID,
		ClientSecretHash: sso.HashClientSecret(cfg.DefaultClientSecret),
		Name:             cfg.DefaultClientName,
		RedirectURIs:     uris,
		IsActive:         true,
		RequirePKCE:      cfg.DefaultClientRequirePKCE,
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "client_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"client_secret_hash",
			"name",
			"redirect_uris",
			"is_active",
			"require_pkce",
			"updated_at",
		}),
	}).Create(&client).Error; err != nil {
		return fmt.Errorf("seed sso client: %w", err)
	}

	return nil
}
