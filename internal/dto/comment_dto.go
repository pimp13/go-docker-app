package dto

import (
	"test/internal/model"
	"time"

	"github.com/google/uuid"
)

type CreatePublicCommentRequest struct {
	AuthorName  string     `json:"author_name" validate:"required,min=2,max=120"`
	AuthorEmail string     `json:"author_email" validate:"omitempty,email,max=255"`
	Body        string     `json:"body" validate:"required,min=2,max=4000"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
}

type UpdateCommentRequest struct {
	Body   *string              `json:"body" validate:"omitempty,min=2,max=4000"`
	Status *model.CommentStatus `json:"status" validate:"omitempty,oneof=pending approved rejected spam"`
}

type ListCommentsQuery struct {
	Status    string `query:"status"`
	ArticleID string `query:"article_id"`
	Search    string `query:"search"`
}

type CommentResponse struct {
	ID           uuid.UUID           `json:"id"`
	ArticleID    uuid.UUID           `json:"article_id"`
	ArticleSlug  string              `json:"article_slug,omitempty"`
	ArticleTitle string              `json:"article_title,omitempty"`
	ParentID     *uuid.UUID          `json:"parent_id,omitempty"`
	UserID       *uuid.UUID          `json:"user_id,omitempty"`
	AuthorName   string              `json:"author_name"`
	AuthorEmail  string              `json:"author_email,omitempty"`
	Body         string              `json:"body"`
	Status       model.CommentStatus `json:"status"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Replies      []CommentResponse   `json:"replies,omitempty"`
}
