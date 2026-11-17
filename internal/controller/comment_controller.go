package controller

import (
	"errors"
	"strings"
	"test/internal/dto"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/response"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type CommentController struct {
	comments service.CommentService
	auth     *AuthController
}

func NewCommentController(comments service.CommentService, auth *AuthController) *CommentController {
	return &CommentController{comments: comments, auth: auth}
}

func (c *CommentController) Register(api *echo.Group) {
	api.GET("/public/articles/:slug/comments", c.ListPublicComments)
	api.POST("/public/articles/:slug/comments", c.CreatePublicComment, c.auth.OptionalAuth)

	api.GET("/comments", c.ListComments, c.auth.RequireAuth, c.auth.RequirePermission(model.PermCommentsRead))
	api.PATCH("/comments/:id", c.UpdateComment, c.auth.RequireAuth, c.auth.RequirePermission(model.PermCommentsUpdate))
	api.DELETE("/comments/:id", c.DeleteComment, c.auth.RequireAuth, c.auth.RequirePermission(model.PermCommentsDelete))
}

// ListPublicComments returns approved comments for a published article.
func (c *CommentController) ListPublicComments(ctx *echo.Context) error {
	slug := ctx.Param("slug")
	if slug == "" {
		return response.BadRequest(ctx, errors.New("slug is required"))
	}

	result, err := c.comments.ListPublicComments(slug)
	if err != nil {
		return writeCommentError(ctx, err)
	}
	return response.OK(ctx, result)
}

// CreatePublicComment creates a reader comment on a published article.
func (c *CommentController) CreatePublicComment(ctx *echo.Context) error {
	slug := ctx.Param("slug")
	if slug == "" {
		return response.BadRequest(ctx, errors.New("slug is required"))
	}

	var body dto.CreatePublicCommentRequest
	if err := ctx.Bind(&body); err != nil {
		return response.BadRequest(ctx, err)
	}

	var userID *uuid.UUID
	authorFallback := ""
	if user, err := c.auth.CurrentUser(ctx); err == nil {
		if parsed, parseErr := uuid.Parse(user.ID); parseErr == nil {
			userID = &parsed
		}
		authorFallback = strings.TrimSpace(user.FirstName + " " + user.LastName)
		if authorFallback == "" {
			authorFallback = user.Username
		}
		if authorFallback == "" {
			authorFallback = user.Email
		}
		if strings.TrimSpace(body.AuthorName) == "" {
			body.AuthorName = authorFallback
		}
		if strings.TrimSpace(body.AuthorEmail) == "" {
			body.AuthorEmail = user.Email
		}
	}

	if err := ctx.Validate(&body); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.comments.CreatePublicComment(slug, body, userID, authorFallback)
	if err != nil {
		return writeCommentError(ctx, err)
	}
	return response.OK(ctx, result)
}

// ListComments lists comments for moderation.
func (c *CommentController) ListComments(ctx *echo.Context) error {
	var query dto.ListCommentsQuery
	if err := ctx.Bind(&query); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.comments.ListComments(query)
	if err != nil {
		return writeCommentError(ctx, err)
	}
	return response.OK(ctx, result)
}

// UpdateComment updates body/status for moderation.
func (c *CommentController) UpdateComment(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}

	var body dto.UpdateCommentRequest
	if err := ctx.Bind(&body); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&body); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.comments.UpdateComment(id, body)
	if err != nil {
		return writeCommentError(ctx, err)
	}
	return response.OK(ctx, result)
}

// DeleteComment removes a comment.
func (c *CommentController) DeleteComment(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}

	if err := c.comments.DeleteComment(id); err != nil {
		return writeCommentError(ctx, err)
	}
	return response.OK(ctx, map[string]string{"id": id})
}

func writeCommentError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrCommentNotFound),
		errors.Is(err, service.ErrCommentArticleGone):
		return response.NotFound(ctx, err)
	case errors.Is(err, service.ErrInvalidCommentStatus),
		errors.Is(err, service.ErrInvalidCommentParent):
		return response.BadRequest(ctx, err)
	default:
		msg := err.Error()
		if msg == "author_name is required" ||
			msg == "nested replies deeper than one level are not allowed" {
			return response.BadRequest(ctx, err)
		}
		return response.InternalServerError(ctx, err)
	}
}

