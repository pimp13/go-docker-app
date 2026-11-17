package seed

import (
	"fmt"
	"strings"
	"test/internal/config"
	"test/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SetupOwner ensures the configured owner / super-admin account exists with
// the super_admin role (full permission catalog).
// Skipped when SUPER_ADMIN_EMAIL or SUPER_ADMIN_PASSWORD is empty.
func SetupOwner(db *gorm.DB, cfg config.OwnerConfig) error {
	email := strings.ToLower(strings.TrimSpace(cfg.Email))
	password := cfg.Password
	if email == "" || password == "" {
		return nil
	}
	if len(password) < 8 {
		return fmt.Errorf("SUPER_ADMIN_PASSWORD must be at least 8 characters")
	}

	var role model.Role
	if err := db.Where("slug = ?", model.RoleSlugSuperAdmin).Take(&role).Error; err != nil {
		return fmt.Errorf("super_admin role missing (run RBAC seed first): %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash owner password: %w", err)
	}

	firstName, lastName := splitOwnerName(cfg.Name)

	var existing model.User
	err = db.Where("email = ?", email).Take(&existing).Error
	switch {
	case err == nil:
		updates := map[string]any{
			"first_name": firstName,
			"last_name":  lastName,
			"password":   string(hashed),
			"role_id":    role.ID,
			"status":     model.UserStatusActive,
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return fmt.Errorf("update owner user: %w", err)
		}
		return nil
	case err == gorm.ErrRecordNotFound:
		user := model.User{
			ID:        uuid.New(),
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Password:  string(hashed),
			RoleID:    role.ID,
			Status:    model.UserStatusActive,
			Meta:      datatypes.JSONMap{"seeded_owner": true},
		}
		// Password is already hashed — skip BeforeCreate re-hash.
		if err := db.Session(&gorm.Session{SkipHooks: true}).Omit("Username").Create(&user).Error; err != nil {
			return fmt.Errorf("create owner user: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("lookup owner user: %w", err)
	}
}

func splitOwnerName(name string) (firstName, lastName string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Super", "Admin"
	}
	parts := strings.Fields(name)
	if len(parts) == 1 {
		return parts[0], "Admin"
	}
	return parts[0], strings.Join(parts[1:], " ")
}
