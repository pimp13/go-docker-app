package service

import (
	"errors"
	"strings"
	"test/internal/dto"
	"test/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrArticleNotFound    = errors.New("article not found")
	ErrInvalidArticleStatus = errors.New("invalid article status")
)

type ArticleService interface {
	CreateArticle(dto.CreateArticleRequest) (dto.ArticleResponse, error)
	UpdateArticle(id string, body dto.UpdateArticleRequest) (dto.ArticleResponse, error)
	DeleteArticle(id string) error
	GetArticles(query dto.ListArticlesQuery) ([]dto.ArticleResponse, error)
	GetArticleByID(id string) (dto.ArticleResponse, error)
	SecretKey() string
}

type articleService struct {
	db        *gorm.DB
	secretKey string
}

func NewArticleService(db *gorm.DB, secretKey string) ArticleService {
	return &articleService{db, secretKey}
}

func (s *articleService) SecretKey() string {
	return s.secretKey
}

func (s *articleService) validateArticleStatus(status model.ArticleStatus) error {
	switch status {
	case model.ArticleStatusDraft, model.ArticleStatusPublished, model.ArticleStatusArchived:
		return nil
	default:
		return ErrInvalidArticleStatus
	}
}

func toArticleResponse(article model.Article) dto.ArticleResponse {
	meta := map[string]any(article.Meta)
	if meta == nil {
		meta = map[string]any{}
	}

	return dto.ArticleResponse{
		ID:          article.ID,
		Title:       article.Title,
		Slug:        article.Slug,
		Excerpt:     article.Excerpt,
		Content:     article.Content,
		CoverImage:  article.CoverImage,
		Status:      article.Status,
		PublishedAt: article.PublishedAt,
		Meta:        meta,
		AuthorID:    article.AuthorID,
		CreatedAt:   article.CreatedAt,
		UpdatedAt:   article.UpdatedAt,
		ViewCount:   article.ViewCount,
	}
}

func resolvePublishedAt(status model.ArticleStatus, publishedAt *time.Time) *time.Time {
	if status != model.ArticleStatusPublished {
		return publishedAt
	}
	if publishedAt != nil {
		return publishedAt
	}
	now := time.Now()
	return &now
}

func (s *articleService) CreateArticle(bodyData dto.CreateArticleRequest) (dto.ArticleResponse, error) {
	if err := s.validateArticleStatus(bodyData.Status); err != nil {
		return dto.ArticleResponse{}, err
	}

	if bodyData.AuthorID == uuid.Nil {
		return dto.ArticleResponse{}, errors.New("author_id is required")
	}

	newArticle := model.Article{
		Title:       bodyData.Title,
		Slug:        bodyData.Slug,
		Excerpt:     bodyData.Excerpt,
		Content:     bodyData.Content,
		CoverImage:  bodyData.CoverImage,
		Status:      bodyData.Status,
		PublishedAt: resolvePublishedAt(bodyData.Status, bodyData.PublishedAt),
		Meta:        datatypes.JSONMap(bodyData.Meta),
		AuthorID:    bodyData.AuthorID,
	}
	if newArticle.Meta == nil {
		newArticle.Meta = datatypes.JSONMap{}
	}

	if err := s.db.Create(&newArticle).Error; err != nil {
		return dto.ArticleResponse{}, err
	}

	return toArticleResponse(newArticle), nil
}

func (s *articleService) GetArticles(query dto.ListArticlesQuery) ([]dto.ArticleResponse, error) {
	db := s.db.Model(&model.Article{}).Order("created_at DESC")

	if status := strings.TrimSpace(query.Status); status != "" && status != "all" {
		if err := s.validateArticleStatus(model.ArticleStatus(status)); err != nil {
			return nil, err
		}
		db = db.Where("status = ?", status)
	}

	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		db = db.Where("title ILIKE ? OR excerpt ILIKE ? OR slug ILIKE ?", like, like, like)
	}

	var articles []model.Article
	if err := db.Find(&articles).Error; err != nil {
		return nil, err
	}

	response := make([]dto.ArticleResponse, 0, len(articles))
	for _, article := range articles {
		response = append(response, toArticleResponse(article))
	}
	return response, nil
}

func (s *articleService) GetArticleByID(id string) (dto.ArticleResponse, error) {
	article, err := s.findByID(id)
	if err != nil {
		return dto.ArticleResponse{}, err
	}
	return toArticleResponse(article), nil
}

func (s *articleService) UpdateArticle(id string, body dto.UpdateArticleRequest) (dto.ArticleResponse, error) {
	article, err := s.findByID(id)
	if err != nil {
		return dto.ArticleResponse{}, err
	}

	if body.Title != nil {
		article.Title = *body.Title
	}
	if body.Slug != nil {
		article.Slug = *body.Slug
	}
	if body.Excerpt != nil {
		article.Excerpt = *body.Excerpt
	}
	if body.Content != nil {
		article.Content = *body.Content
	}
	if body.CoverImage != nil {
		article.CoverImage = *body.CoverImage
	}
	if body.Status != nil {
		if err := s.validateArticleStatus(*body.Status); err != nil {
			return dto.ArticleResponse{}, err
		}
		article.Status = *body.Status
		article.PublishedAt = resolvePublishedAt(*body.Status, body.PublishedAt)
	} else if body.PublishedAt != nil {
		article.PublishedAt = body.PublishedAt
	}
	if body.Meta != nil {
		article.Meta = datatypes.JSONMap(body.Meta)
	}

	if err := s.db.Save(&article).Error; err != nil {
		return dto.ArticleResponse{}, err
	}

	return toArticleResponse(article), nil
}

func (s *articleService) DeleteArticle(id string) error {
	article, err := s.findByID(id)
	if err != nil {
		return err
	}
	return s.db.Delete(&article).Error
}

func (s *articleService) findByID(id string) (model.Article, error) {
	var article model.Article
	if err := s.db.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Article{}, ErrArticleNotFound
		}
		return model.Article{}, err
	}
	return article, nil
}
