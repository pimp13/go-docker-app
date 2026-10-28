package dto

import (
	"test/internal/model"
	"time"

	"github.com/google/uuid"
)

type CreateArticleRequest struct {
	Title       string              `json:"title" validate:"required,min=3,max=255"`
	Slug        string              `json:"slug"`
	Excerpt     string              `json:"excerpt" validate:"required,min=3,max=255"`
	Content     string              `json:"content" validate:"required,min=3,max=255"`
	CoverImage  string              `json:"cover_image" validate:"required,min=3,max=255"`
	Status      model.ArticleStatus `json:"status" validate:"required,min=3,max=255"`
	PublishedAt *time.Time          `json:"published_at"`
	Meta        map[string]any      `json:"meta"`
	AuthorID    uuid.UUID           `json:"author_id" validate:"required,uuid"`
}

type UpdateArticleRequest struct {
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Excerpt     string         `json:"excerpt"`
	Content     string         `json:"content"`
	CoverImage  string         `json:"cover_image"`
	Status      string         `json:"status"`
	PublishedAt time.Time      `json:"published_at"`
	Meta        map[string]any `json:"meta"`
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
}
