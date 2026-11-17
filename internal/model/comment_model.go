package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusRejected CommentStatus = "rejected"
	CommentStatusSpam     CommentStatus = "spam"
)

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ArticleID uuid.UUID `gorm:"type:uuid;not null;index" json:"article_id"`
	Article   Article   `gorm:"foreignKey:ArticleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"article,omitempty"`

	ParentID *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent   *Comment   `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`

	UserID *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User   *User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user,omitempty"`

	AuthorName  string `gorm:"size:120;not null" json:"author_name"`
	AuthorEmail string `gorm:"size:255" json:"author_email,omitempty"`
	Body        string `gorm:"type:text;not null" json:"body"`

	Status CommentStatus `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`

	Meta datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'" json:"meta"`
}

func (Comment) TableName() string {
	return "comments"
}
