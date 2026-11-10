package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const PermSSOManage = "sso:manage"

// SSOClient is a registered relying party that can use this IdP.
type SSOClient struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	ClientID         string         `gorm:"size:100;not null;uniqueIndex" json:"client_id"`
	ClientSecretHash string         `gorm:"size:128;not null" json:"-"`
	Name             string         `gorm:"size:150;not null" json:"name"`
	RedirectURIs     datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"redirect_uris"`
	IsActive         bool           `gorm:"not null;default:true" json:"is_active"`
	RequirePKCE      bool           `gorm:"not null;default:false" json:"require_pkce"`
}

func (SSOClient) TableName() string {
	return "sso_clients"
}

// SSOAuthorizationCode is a one-time code issued after a successful SSO authorize.
type SSOAuthorizationCode struct {
	ID                  uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt           time.Time  `json:"created_at"`
	CodeHash            string     `gorm:"size:128;not null;uniqueIndex" json:"-"`
	ClientID            string     `gorm:"size:100;not null;index" json:"client_id"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	RedirectURI         string     `gorm:"size:500;not null" json:"redirect_uri"`
	Scopes              string     `gorm:"size:255;not null;default:''" json:"scopes"`
	CodeChallenge       string     `gorm:"size:128" json:"-"`
	CodeChallengeMethod string     `gorm:"size:16" json:"-"`
	ExpiresAt           time.Time  `gorm:"not null;index" json:"expires_at"`
	UsedAt              *time.Time `json:"used_at,omitempty"`
}

func (SSOAuthorizationCode) TableName() string {
	return "sso_authorization_codes"
}
