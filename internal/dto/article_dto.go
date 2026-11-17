package dto

import (
	"test/internal/model"
	"time"

	"github.com/google/uuid"
)

type CreateArticleRequest struct {
	Title       string              `json:"title" validate:"required,min=3,max=255"`
	Slug        string              `json:"slug" validate:"omitempty,min=3,max=255"`
	Excerpt     string              `json:"excerpt" validate:"required,min=3,max=500"`
	Content     string              `json:"content" validate:"required,min=3"`
	CoverImage  string              `json:"cover_image" validate:"omitempty,min=3,max=500"`
	Status      model.ArticleStatus `json:"status" validate:"required,oneof=draft published archived"`
	PublishedAt *time.Time          `json:"published_at"`
	Meta        map[string]any      `json:"meta"`
	AuthorID    uuid.UUID           `json:"author_id" validate:"omitempty,uuid"`
}

type UpdateArticleRequest struct {
	Title       *string              `json:"title" validate:"omitempty,min=3,max=255"`
	Slug        *string              `json:"slug" validate:"omitempty,min=3,max=255"`
	Excerpt     *string              `json:"excerpt" validate:"omitempty,min=3,max=500"`
	Content     *string              `json:"content" validate:"omitempty,min=3"`
	CoverImage  *string              `json:"cover_image" validate:"omitempty,min=3,max=500"`
	Status      *model.ArticleStatus `json:"status" validate:"omitempty,oneof=draft published archived"`
	PublishedAt *time.Time           `json:"published_at"`
	Meta        map[string]any       `json:"meta"`
}

type ListArticlesQuery struct {
	Status string `query:"status"`
	Search string `query:"search"`
}

type ArticleResponse struct {
	ID          uuid.UUID           `json:"id"`
	Title       string              `json:"title"`
	Slug        string              `json:"slug"`
	Excerpt     string              `json:"excerpt"`
	Content     string              `json:"content"`
	CoverImage  string              `json:"cover_image"`
	Status      model.ArticleStatus `json:"status"`
	PublishedAt *time.Time          `json:"published_at"`
	Meta        map[string]any      `json:"meta"`
	AuthorID    uuid.UUID           `json:"author_id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	ViewCount   int64               `json:"view_count"`
}

type ListPublicArticlesQuery struct {
	Search string `query:"search"`
}

// PublicArticleSummary is returned by the public list endpoint (no full body).
type PublicArticleSummary struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	CoverImage  string     `json:"cover_image"`
	PublishedAt *time.Time `json:"published_at"`
	AuthorName  string     `json:"author_name"`
	ViewCount   int64      `json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PublicArticleDetail is the full published article for readers.
type PublicArticleDetail struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Content     string     `json:"content"`
	CoverImage  string     `json:"cover_image"`
	PublishedAt *time.Time `json:"published_at"`
	AuthorName  string     `json:"author_name"`
	ViewCount   int64      `json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ArticleViewResponse struct {
	Slug      string `json:"slug"`
	ViewCount int64  `json:"view_count"`
}
