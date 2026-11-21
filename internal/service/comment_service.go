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
	ErrCommentNotFound      = errors.New("comment not found")
	ErrInvalidCommentStatus = errors.New("invalid comment status")
	ErrCommentArticleGone   = errors.New("article is not available for comments")
	ErrInvalidCommentParent = errors.New("invalid parent comment")
)

type CommentService interface {
	ListPublicComments(articleSlug string) ([]dto.CommentResponse, error)
	CreatePublicComment(articleSlug string, body dto.CreatePublicCommentRequest, userID *uuid.UUID, authorFallback string) (dto.CommentResponse, error)
	ListComments(query dto.ListCommentsQuery) ([]dto.CommentResponse, error)
	UpdateComment(id string, body dto.UpdateCommentRequest) (dto.CommentResponse, error)
	DeleteComment(id string) error
}

type commentService struct {
	db *gorm.DB
}

func NewCommentService(db *gorm.DB) CommentService {
	return &commentService{db: db}
}

func (s *commentService) validateStatus(status model.CommentStatus) error {
	switch status {
	case model.CommentStatusPending, model.CommentStatusApproved, model.CommentStatusRejected, model.CommentStatusSpam:
		return nil
	default:
		return ErrInvalidCommentStatus
	}
}

func toCommentResponse(comment model.Comment) dto.CommentResponse {
	resp := dto.CommentResponse{
		ID:          comment.ID,
		ArticleID:   comment.ArticleID,
		ParentID:    comment.ParentID,
		UserID:      comment.UserID,
		AuthorName:  comment.AuthorName,
		AuthorEmail: comment.AuthorEmail,
		Body:        comment.Body,
		Status:      comment.Status,
		CreatedAt:   comment.CreatedAt,
		UpdatedAt:   comment.UpdatedAt,
	}
	if comment.Article.ID != uuid.Nil {
		resp.ArticleSlug = comment.Article.Slug
		resp.ArticleTitle = comment.Article.Title
	}
	return resp
}

func (s *commentService) findPublishedArticleBySlug(slug string) (model.Article, error) {
	var article model.Article
	err := s.db.Where("slug = ? AND status = ?", strings.TrimSpace(slug), model.ArticleStatusPublished).
		First(&article).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Article{}, ErrCommentArticleGone
		}
		return model.Article{}, err
	}
	return article, nil
}

func (s *commentService) ListPublicComments(articleSlug string) ([]dto.CommentResponse, error) {
	article, err := s.findPublishedArticleBySlug(articleSlug)
	if err != nil {
		return nil, err
	}

	var comments []model.Comment
	if err := s.db.
		Where("article_id = ? AND status = ?", article.ID, model.CommentStatusApproved).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, err
	}

	return nestComments(comments), nil
}

func nestComments(comments []model.Comment) []dto.CommentResponse {
	byID := make(map[uuid.UUID]*dto.CommentResponse, len(comments))
	order := make([]uuid.UUID, 0, len(comments))

	for _, comment := range comments {
		item := toCommentResponse(comment)
		item.Replies = []dto.CommentResponse{}
		copyItem := item
		byID[comment.ID] = &copyItem
		order = append(order, comment.ID)
	}

	// Attach replies onto the live pointers first, then materialize roots.
	// Copying roots before replies are attached drops nested children.
	rootIDs := make([]uuid.UUID, 0)
	for _, id := range order {
		item := byID[id]
		if item.ParentID != nil {
			if parent, ok := byID[*item.ParentID]; ok {
				parent.Replies = append(parent.Replies, *item)
				continue
			}
		}
		rootIDs = append(rootIDs, id)
	}

	roots := make([]dto.CommentResponse, 0, len(rootIDs))
	for _, id := range rootIDs {
		roots = append(roots, *byID[id])
	}
	return roots
}

func (s *commentService) CreatePublicComment(
	articleSlug string,
	body dto.CreatePublicCommentRequest,
	userID *uuid.UUID,
	authorFallback string,
) (dto.CommentResponse, error) {
	article, err := s.findPublishedArticleBySlug(articleSlug)
	if err != nil {
		return dto.CommentResponse{}, err
	}

	authorName := strings.TrimSpace(body.AuthorName)
	if authorName == "" {
		authorName = strings.TrimSpace(authorFallback)
	}
	if authorName == "" {
		return dto.CommentResponse{}, errors.New("author_name is required")
	}

	if body.ParentID != nil {
		var parent model.Comment
		if err := s.db.Where("id = ? AND article_id = ? AND status = ?", *body.ParentID, article.ID, model.CommentStatusApproved).
			First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return dto.CommentResponse{}, ErrInvalidCommentParent
			}
			return dto.CommentResponse{}, err
		}
		if parent.ParentID != nil {
			return dto.CommentResponse{}, errors.New("nested replies deeper than one level are not allowed")
		}
	}

	status := model.CommentStatusPending
	if userID != nil {
		status = model.CommentStatusApproved
	}

	comment := model.Comment{
		ArticleID:   article.ID,
		ParentID:    body.ParentID,
		UserID:      userID,
		AuthorName:  authorName,
		AuthorEmail: strings.TrimSpace(body.AuthorEmail),
		Body:        strings.TrimSpace(body.Body),
		Status:      status,
		Meta:        datatypes.JSONMap{},
	}

	if err := s.db.Create(&comment).Error; err != nil {
		return dto.CommentResponse{}, err
	}

	comment.Article = article
	return toCommentResponse(comment), nil
}

func (s *commentService) ListComments(query dto.ListCommentsQuery) ([]dto.CommentResponse, error) {
	db := s.db.Model(&model.Comment{}).
		Preload("Article").
		Order("created_at DESC")

	if status := strings.TrimSpace(query.Status); status != "" && status != "all" {
		if err := s.validateStatus(model.CommentStatus(status)); err != nil {
			return nil, err
		}
		db = db.Where("status = ?", status)
	}

	if articleID := strings.TrimSpace(query.ArticleID); articleID != "" {
		db = db.Where("article_id = ?", articleID)
	}

	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		db = db.Where("body ILIKE ? OR author_name ILIKE ? OR author_email ILIKE ?", like, like, like)
	}

	var comments []model.Comment
	if err := db.Find(&comments).Error; err != nil {
		return nil, err
	}

	out := make([]dto.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		out = append(out, toCommentResponse(comment))
	}
	return out, nil
}

func (s *commentService) UpdateComment(id string, body dto.UpdateCommentRequest) (dto.CommentResponse, error) {
	var comment model.Comment
	if err := s.db.Preload("Article").Where("id = ?", id).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CommentResponse{}, ErrCommentNotFound
		}
		return dto.CommentResponse{}, err
	}

	if body.Body != nil {
		comment.Body = strings.TrimSpace(*body.Body)
	}
	if body.Status != nil {
		if err := s.validateStatus(*body.Status); err != nil {
			return dto.CommentResponse{}, err
		}
		comment.Status = *body.Status
	}
	comment.UpdatedAt = time.Now()

	if err := s.db.Save(&comment).Error; err != nil {
		return dto.CommentResponse{}, err
	}

	return toCommentResponse(comment), nil
}

func (s *commentService) DeleteComment(id string) error {
	result := s.db.Where("id = ?", id).Delete(&model.Comment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}
