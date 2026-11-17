package seed

import (
	"fmt"
	"strings"
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

// resourceCRUD describes a module and which actions it exposes.
// Empty Actions defaults to standard CRUD: read, create, update, delete.
type resourceCRUD struct {
	Resource    string
	Actions     []string
	Description string // optional base description; action-specific text is generated
}

var defaultCRUDActions = []string{"read", "create", "update", "delete"}

// modules is the single place to register new resources.
// Add a row here → permissions are created on next seed.
var modules = []resourceCRUD{
	{Resource: "articles", Description: "articles"},
	{Resource: "comments", Description: "comments"},
	{Resource: "users", Description: "users"},
	{Resource: "roles", Actions: []string{"manage"}, Description: "roles and permissions"},
	{Resource: "sso", Actions: []string{"manage"}, Description: "SSO clients"},
}

var catalog = buildCatalog(modules)

func buildCatalog(resources []resourceCRUD) []permissionDef {
	out := make([]permissionDef, 0, len(resources)*4)
	for _, r := range resources {
		actions := r.Actions
		if len(actions) == 0 {
			actions = defaultCRUDActions
		}
		label := r.Description
		if label == "" {
			label = r.Resource
		}
		for _, action := range actions {
			out = append(out, permissionDef{
				Code:        permCode(r.Resource, action),
				Name:        humanName(action, r.Resource),
				Resource:    r.Resource,
				Action:      action,
				Description: humanDescription(action, label),
			})
		}
	}
	return out
}

func permCode(resource, action string) string {
	return resource + ":" + action
}

func humanName(action, resource string) string {
	return strings.ToUpper(action[:1]) + action[1:] + " " + resource
}

func humanDescription(action, label string) string {
	switch action {
	case "read":
		return "List and view " + label
	case "create":
		return "Create " + label
	case "update":
		return "Edit " + label
	case "delete":
		return "Delete " + label
	case "manage":
		return "Manage " + label
	default:
		return strings.ToUpper(action[:1]) + action[1:] + " " + label
	}
}

// permCodes builds resource:action codes for role assignment.
func permCodes(resource string, actions ...string) []string {
	codes := make([]string, 0, len(actions))
	for _, a := range actions {
		codes = append(codes, permCode(resource, a))
	}
	return codes
}

// crudCodes is shorthand for full CRUD on a resource.
func crudCodes(resource string) []string {
	return permCodes(resource, defaultCRUDActions...)
}

// mergePerms concatenates permission code slices.
func mergePerms(groups ...[]string) []string {
	n := 0
	for _, g := range groups {
		n += len(g)
	}
	out := make([]string, 0, n)
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// allPermissionCodes returns every code in the catalog (useful for admin).
func allPermissionCodes() []string {
	codes := make([]string, 0, len(catalog))
	for _, p := range catalog {
		codes = append(codes, p.Code)
	}
	return codes
}

// builtinRoles maps which permissions each system role gets.
// وقتی ماژول جدید اضافه می‌کنی، اینجا هم به roleهای مورد نظر وصلش کن
// (admin با allPermissionCodes خودکار همه را می‌گیرد).
func builtinRoles() []roleDef {
	return []roleDef{
		{
			Slug:        model.RoleSlugUser,
			Name:        "User",
			Description: "Default authenticated member",
			Permissions: permCodes("articles", "read"),
		},
		{
			Slug:        model.RoleSlugEditor,
			Name:        "Editor",
			Description: "Can create and edit content",
			Permissions: mergePerms(
				permCodes("articles", "read", "create", "update"),
				permCodes("comments", "read", "update", "delete"),
			),
		},
		{
			Slug:        model.RoleSlugAdmin,
			Name:        "Admin",
			Description: "Full access including all comment permissions",
			Permissions: mergePerms(
				allPermissionCodes(),
				crudCodes("comments"),
			),
		},
		// Legacy / dashboard slug used by existing accounts.
		{
			Slug:        model.RoleSlugSuperAdmin,
			Name:        "Super Admin",
			Description: "Full access including all comment permissions",
			Permissions: mergePerms(
				allPermissionCodes(),
				crudCodes("comments"),
			),
		},
	}
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
		&model.Comment{},
	); err != nil {
		return fmt.Errorf("auto migrate domain: %w", err)
	}

	if err := migrateLegacyUserRoles(db); err != nil {
		return fmt.Errorf("migrate legacy roles: %w", err)
	}

	return nil
}

func seed(db *gorm.DB) error {
	roles := builtinRoles()

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

			// Replace: هر بار seed، permissionهای role سیستمی دقیقاً مطابق builtinRoles می‌شود.
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
