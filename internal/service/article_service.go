package service

import (
	"errors"
	"test/internal/dto"
	"test/internal/model"
	"time"

	"gorm.io/gorm"
)

type ArticleService interface {
	CreateArticle(dto.CreateArticleRequest) (dto.ArticleResponse, error)

	SecretKey() string

	GetArticles() ([]dto.ArticleResponse, error)

	GetArticleByID(id string) (dto.ArticleResponse, error)
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
	case model.ArticleStatusDraft:
		return nil
	case model.ArticleStatusPublished:
		return nil
	case model.ArticleStatusArchived:
		return nil
	default:
		return errors.New("invalid article status")
	}
}

func (s *articleService) CreateArticle(bodyData dto.CreateArticleRequest) (dto.ArticleResponse, error) {
	if err := s.validateArticleStatus(bodyData.Status); err != nil {
		return dto.ArticleResponse{}, err
	}

	publishedAt := bodyData.PublishedAt
	if publishedAt == nil {
		now := time.Now()
		publishedAt = &now
	}

	newArticle := model.Article{
		Title:       bodyData.Title,
		Slug:        bodyData.Slug,
		Excerpt:     bodyData.Excerpt,
		Content:     bodyData.Content,
		CoverImage:  bodyData.CoverImage,
		Status:      bodyData.Status,
		PublishedAt: publishedAt,
		Meta:        bodyData.Meta,
		AuthorID:    bodyData.AuthorID,
	}

	if err := s.db.Create(&newArticle).Error; err != nil {
		return dto.ArticleResponse{}, err
	}

	return dto.ArticleResponse{
		ID: newArticle.ID,
	}, nil
}

func (s *articleService) GetArticles() ([]dto.ArticleResponse, error) {
	var articles []model.Article
	if err := s.db.Order("created_at DESC").Find(&articles).Error; err != nil {
		return nil, err
	}

	var response []dto.ArticleResponse
	for _, article := range articles {
		response = append(response, dto.ArticleResponse{
			ID:          article.ID,
			Title:       article.Title,
			Slug:        article.Slug,
			Excerpt:     article.Excerpt,
			Content:     article.Content,
			CoverImage:  article.CoverImage,
			Status:      article.Status,
			PublishedAt: article.PublishedAt,
			Meta:        article.Meta,
			AuthorID:    article.AuthorID,
		})
	}
	return response, nil
}

func (s *articleService) GetArticleByID(id string) (dto.ArticleResponse, error) {
	var article model.Article
	if err := s.db.Debug().Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ArticleResponse{}, errors.New("article not found")
		}
		return dto.ArticleResponse{}, err
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
		Meta:        article.Meta,
		AuthorID:    article.AuthorID,
	}, nil
}
