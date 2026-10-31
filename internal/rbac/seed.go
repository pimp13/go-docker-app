package rbac

import (
	"fmt"
	"test/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type permissionDef struct {
	Code        string
	Name        string
	Resource    string
	Action      string
	Description string
}

type roleDef struct {
	Slug        string
	Name        string
	Description string
	Permissions []string
}

var catalog = []permissionDef{
	{model.PermArticlesRead, "Read articles", "articles", "read", "List and view articles"},
	{model.PermArticlesCreate, "Create articles", "articles", "create", "Create new articles"},
	{model.PermArticlesUpdate, "Update articles", "articles", "update", "Edit existing articles"},
	{model.PermArticlesDelete, "Delete articles", "articles", "delete", "Delete articles"},
	{model.PermUsersRead, "Read users", "users", "read", "List and view users"},
	{model.PermUsersCreate, "Create users", "users", "create", "Create users"},
	{model.PermUsersUpdate, "Update users", "users", "update", "Edit users"},
	{model.PermUsersDelete, "Delete users", "users", "delete", "Delete users"},
	{model.PermRolesManage, "Manage roles", "roles", "manage", "Manage roles and permissions"},
}

var roles = []roleDef{
	{
		Slug:        model.RoleSlugUser,
		Name:        "User",
		Description: "Default authenticated member",
		Permissions: []string{
			model.PermArticlesRead,
		},
	},
	{
		Slug:        model.RoleSlugEditor,
		Name:        "Editor",
		Description: "Can create and edit content",
		Permissions: []string{
			model.PermArticlesRead,
			model.PermArticlesCreate,
			model.PermArticlesUpdate,
		},
	},
	{
		Slug:        model.RoleSlugAdmin,
		Name:        "Admin",
		Description: "Full access",
		Permissions: []string{
			model.PermArticlesRead,
			model.PermArticlesCreate,
			model.PermArticlesUpdate,
			model.PermArticlesDelete,
			model.PermUsersRead,
			model.PermUsersCreate,
			model.PermUsersUpdate,
			model.PermUsersDelete,
			model.PermRolesManage,
		},
	},
}

// Setup migrates RBAC tables, seeds built-in roles/permissions, and upgrades users.
func Setup(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
	); err != nil {
		return fmt.Errorf("auto migrate rbac: %w", err)
	}

	if err := seed(db); err != nil {
		return fmt.Errorf("seed rbac: %w", err)
	}

	if err := ensureUserRoleColumn(db); err != nil {
		return fmt.Errorf("prepare user roles: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Article{},
		&model.Tag{},
	); err != nil {
		return fmt.Errorf("auto migrate domain: %w", err)
	}

	if err := migrateLegacyUserRoles(db); err != nil {
		return fmt.Errorf("migrate legacy roles: %w", err)
	}

	return nil
}

func seed(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		permsByCode := make(map[string]model.Permission, len(catalog))
		for _, def := range catalog {
			perm := model.Permission{
				ID:          uuid.New(),
				Code:        def.Code,
				Name:        def.Name,
				Resource:    def.Resource,
				Action:      def.Action,
				Description: def.Description,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "code"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "resource", "action", "description", "updated_at"}),
			}).Create(&perm).Error; err != nil {
				return err
			}

			var stored model.Permission
			if err := tx.Where("code = ?", def.Code).Take(&stored).Error; err != nil {
				return err
			}
			permsByCode[def.Code] = stored
		}

		for _, def := range roles {
			role := model.Role{
				ID:          uuid.New(),
				Slug:        def.Slug,
				Name:        def.Name,
				Description: def.Description,
				IsSystem:    true,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "slug"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "description", "is_system", "updated_at"}),
			}).Create(&role).Error; err != nil {
				return err
			}

			var stored model.Role
			if err := tx.Where("slug = ?", def.Slug).Take(&stored).Error; err != nil {
				return err
			}

			wanted := make([]model.Permission, 0, len(def.Permissions))
			for _, code := range def.Permissions {
				perm, ok := permsByCode[code]
				if !ok {
					return fmt.Errorf("unknown permission %q for role %q", code, def.Slug)
				}
				wanted = append(wanted, perm)
			}

			if err := tx.Model(&stored).Association("Permissions").Replace(wanted); err != nil {
				return err
			}
		}

		return nil
	})
}

func ensureUserRoleColumn(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.User{}) {
		return nil
	}
	if db.Migrator().HasColumn(&model.User{}, "role_id") {
		return nil
	}
	return db.Exec(`ALTER TABLE users ADD COLUMN role_id uuid`).Error
}

func migrateLegacyUserRoles(db *gorm.DB) error {
	var defaultRole model.Role
	if err := db.Where("slug = ?", model.RoleSlugUser).Take(&defaultRole).Error; err != nil {
		return err
	}

	if db.Migrator().HasColumn(&model.User{}, "role") {
		if err := db.Exec(`
			UPDATE users AS u
			SET role_id = r.id
			FROM roles AS r
			WHERE (u.role_id IS NULL OR u.role_id = '00000000-0000-0000-0000-000000000000')
			  AND r.slug = u.role
		`).Error; err != nil {
			return err
		}

		if err := db.Migrator().DropColumn(&model.User{}, "role"); err != nil {
			return err
		}
	}

	if err := db.Exec(`
		UPDATE users
		SET role_id = ?
		WHERE role_id IS NULL OR role_id = '00000000-0000-0000-0000-000000000000'
	`, defaultRole.ID).Error; err != nil {
		return err
	}

	return db.Exec(`ALTER TABLE users ALTER COLUMN role_id SET NOT NULL`).Error
}
