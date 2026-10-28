package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"
)

type Article struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// محتوا
	Title      string `gorm:"size:255;not null;index" json:"title"`
	Slug       string `gorm:"size:255;not null;uniqueIndex" json:"slug"`
	Excerpt    string `gorm:"type:text" json:"excerpt,omitempty"`
	Content    string `gorm:"type:text;not null" json:"content"`
	CoverImage string `gorm:"size:500" json:"cover_image,omitempty"`

	// وضعیت و انتشار
	Status      ArticleStatus `gorm:"type:varchar(20);default:'draft';not null;index" json:"status"`
	PublishedAt *time.Time    `gorm:"index" json:"published_at,omitempty"`

	// آمار
	ViewCount int64 `gorm:"default:0" json:"view_count"`

	// ارتباط با نویسنده
	AuthorID uuid.UUID `gorm:"type:uuid;not null;index" json:"author_id"`
	Author   User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`

	// SEO (اختیاری ولی مفید)
	MetaTitle       string            `gorm:"size:255" json:"meta_title,omitempty"`
	MetaDescription string            `gorm:"size:500" json:"meta_description,omitempty"`
	Meta            datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'" json:"meta"`

	// تگ‌ها (Many-to-Many)
	Tags []Tag `gorm:"many2many:article_tags;" json:"tags,omitempty"`
}

// TableName
func (Article) TableName() string {
	return "articles"
}

// BeforeCreate تولید خودکار Slug در صورت خالی بودن
func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.Slug == "" && a.Title != "" {
		a.Slug = generateSlug(a.Title)
	}
	return nil
}

// IsPublished چک کردن وضعیت انتشار
func (a *Article) IsPublished() bool {
	return a.Status == ArticleStatusPublished && a.PublishedAt != nil
}

// ==================== مدل Tag (اختیاری ولی توصیه‌شده) ====================

type Tag struct {
	ID        uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Meta      datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'" json:"meta"`

	Name string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Slug string `gorm:"size:100;not null;uniqueIndex" json:"slug"`
}

func (Tag) TableName() string {
	return "tags"
}
