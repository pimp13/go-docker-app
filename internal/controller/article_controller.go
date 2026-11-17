package controller

import (
	"errors"
	"test/internal/dto"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/response"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type ArticleController struct {
	articleService service.ArticleService
	auth           *AuthController
}

func NewArticleController(articleService service.ArticleService, auth *AuthController) *ArticleController {
	return &ArticleController{articleService: articleService, auth: auth}
}

func (c *ArticleController) Register(api *echo.Group) {
	// Public reader endpoints (no auth)
	api.GET("/public/articles", c.GetPublishedArticles)
	api.GET("/public/articles/:slug", c.GetPublishedArticleBySlug)
	api.POST("/public/articles/:slug/view", c.RecordArticleView)

	// Admin / editorial endpoints
	api.GET("/articles", c.GetArticles, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesRead))
	api.POST("/articles/create", c.CreateArticle, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesCreate))
	api.GET("/articles/:id", c.GetArticleByID, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesRead))
	api.PUT("/articles/:id", c.UpdateArticle, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesUpdate))
	api.PATCH("/articles/:id", c.UpdateArticle, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesUpdate))
	api.DELETE("/articles/:id", c.DeleteArticle, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesDelete))
}

// CreateArticle creates an article.
//
//	@Summary		Create article
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			body	body		dto.CreateArticleRequest	true	"Article payload"
//	@Success		200		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Router			/api/articles/create [post]
func (c *ArticleController) CreateArticle(ctx *echo.Context) error {
	var bodyData dto.CreateArticleRequest
	if err := ctx.Bind(&bodyData); err != nil {
		return response.BadRequest(ctx, err)
	}

	if err := ctx.Validate(&bodyData); err != nil {
		return response.BadRequest(ctx, err)
	}

	if bodyData.AuthorID == uuid.Nil {
		user, err := c.auth.CurrentUser(ctx)
		if err != nil {
			return response.Unauthorized(ctx, err)
		}
		authorID, err := uuid.Parse(user.ID)
		if err != nil {
			return response.BadRequest(ctx, errors.New("invalid authenticated user id"))
		}
		bodyData.AuthorID = authorID
	}

	result, err := c.articleService.CreateArticle(bodyData)
	if err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, result)
}

// GetPublishedArticles lists published articles for readers.
//
//	@Summary		List published articles
//	@Tags			public-articles
//	@Produce		json
//	@Param			search	query		string	false	"Search title, excerpt, or slug"
//	@Success		200		{object}	response.APIResponse
//	@Router			/api/public/articles [get]
func (c *ArticleController) GetPublishedArticles(ctx *echo.Context) error {
	var query dto.ListPublicArticlesQuery
	if err := ctx.Bind(&query); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.articleService.GetPublishedArticles(query)
	if err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, result)
}

// GetPublishedArticleBySlug returns one published article.
//
//	@Summary		Get published article by slug
//	@Tags			public-articles
//	@Produce		json
//	@Param			slug	path		string	true	"Article slug"
//	@Success		200		{object}	response.APIResponse
//	@Failure		404		{object}	response.APIResponse
//	@Router			/api/public/articles/{slug} [get]
func (c *ArticleController) GetPublishedArticleBySlug(ctx *echo.Context) error {
	slug := ctx.Param("slug")
	if slug == "" {
		return response.BadRequest(ctx, errors.New("slug is required"))
	}

	result, err := c.articleService.GetPublishedArticleBySlug(slug)
	if err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, result)
}

// RecordArticleView increments the view counter for a published article.
//
//	@Summary		Record article view
//	@Tags			public-articles
//	@Produce		json
//	@Param			slug	path		string	true	"Article slug"
//	@Success		200		{object}	response.APIResponse
//	@Failure		404		{object}	response.APIResponse
//	@Router			/api/public/articles/{slug}/view [post]
func (c *ArticleController) RecordArticleView(ctx *echo.Context) error {
	slug := ctx.Param("slug")
	if slug == "" {
		return response.BadRequest(ctx, errors.New("slug is required"))
	}

	result, err := c.articleService.RecordArticleView(slug)
	if err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, result)
}

// GetArticles lists articles.
//
//	@Summary		List articles
//	@Tags			articles
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			status	query		string	false	"Filter by status (draft|published|archived|all)"
//	@Param			search	query		string	false	"Search title, excerpt, or slug"
//	@Success		200		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Router			/api/articles [get]
func (c *ArticleController) GetArticles(ctx *echo.Context) error {
	var query dto.ListArticlesQuery
	if err := ctx.Bind(&query); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.articleService.GetArticles(query)
	if err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, result)
}

// GetArticleByID returns one article.
//
//	@Summary		Get article by ID
//	@Tags			articles
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id	path		string	true	"Article ID"
//	@Success		200	{object}	response.APIResponse
//	@Failure		400	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Failure		404	{object}	response.APIResponse
//	@Router			/api/articles/{id} [get]
func (c *ArticleController) GetArticleByID(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}

	result, err := c.articleService.GetArticleByID(id)
	if err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, result)
}

// UpdateArticle updates an article.
//
//	@Summary		Update article
//	@Tags			articles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id		path		string					true	"Article ID"
//	@Param			body	body		dto.UpdateArticleRequest	true	"Article payload"
//	@Success		200		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Failure		403		{object}	response.APIResponse
//	@Failure		404		{object}	response.APIResponse
//	@Router			/api/articles/{id} [put]
func (c *ArticleController) UpdateArticle(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}

	var bodyData dto.UpdateArticleRequest
	if err := ctx.Bind(&bodyData); err != nil {
		return response.BadRequest(ctx, err)
	}
	if err := ctx.Validate(&bodyData); err != nil {
		return response.BadRequest(ctx, err)
	}

	result, err := c.articleService.UpdateArticle(id, bodyData)
	if err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, result)
}

// DeleteArticle deletes an article.
//
//	@Summary		Delete article
//	@Tags			articles
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			id	path		string	true	"Article ID"
//	@Success		200	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Failure		404	{object}	response.APIResponse
//	@Router			/api/articles/{id} [delete]
func (c *ArticleController) DeleteArticle(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}

	if err := c.articleService.DeleteArticle(id); err != nil {
		return writeArticleError(ctx, err)
	}

	return response.OK(ctx, map[string]string{"id": id})
}

func writeArticleError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrArticleNotFound):
		return response.NotFound(ctx, err)
	case errors.Is(err, service.ErrInvalidArticleStatus):
		return response.BadRequest(ctx, err)
	default:
		return response.InternalServerError(ctx, err)
	}
}
