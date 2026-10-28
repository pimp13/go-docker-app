package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBanned   UserStatus = "banned"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	FirstName string            `gorm:"size:100;not null" json:"first_name"`
	LastName  string            `gorm:"size:100;not null" json:"last_name"`
	Email     string            `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Username  string            `gorm:"size:50;uniqueIndex" json:"username,omitempty"`
	Password  string            `gorm:"size:255;not null" json:"-"`
	Phone     string            `gorm:"size:20" json:"phone,omitempty"`
	Avatar    string            `gorm:"size:500" json:"avatar,omitempty"`
	Role      UserRole          `gorm:"type:varchar(20);default:'user';not null" json:"role"`
	Status    UserStatus        `gorm:"type:varchar(20);default:'active';not null" json:"status"`
	LastLogin *time.Time        `json:"last_login,omitempty"`
	Meta      datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'" json:"meta"`
}

// TableName تنظیم نام جدول (اختیاری)
func (User) TableName() string {
	return "users"
}

// BeforeCreate هش کردن پسورد قبل از ذخیره
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashed)
	}
	return nil
}

// CheckPassword بررسی صحت پسورد
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// FullName برگرداندن نام کامل
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}
