package controller

import (
	"errors"
	"test/internal/dto"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/response"

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
	api.GET("/articles", c.GetArticles, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesRead))
	api.POST("/articles/create", c.CreateArticle, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesCreate))
	api.GET("/articles/:id", c.GetArticleByID, c.auth.RequireAuth, c.auth.RequirePermission(model.PermArticlesRead))
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

	result, err := c.articleService.CreateArticle(bodyData)
	if err != nil {
		return response.InternalServerError(ctx, err)
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
//	@Success		200	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Router			/api/articles [get]
func (c *ArticleController) GetArticles(ctx *echo.Context) error {
	result, err := c.articleService.GetArticles()
	if err != nil {
		return response.InternalServerError(ctx, err)
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
//	@Router			/api/articles/{id} [get]
func (c *ArticleController) GetArticleByID(ctx *echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return response.BadRequest(ctx, errors.New("id is required"))
	}

	result, err := c.articleService.GetArticleByID(id)
	if err != nil {
		return response.InternalServerError(ctx, err)
	}

	return response.OK(ctx, result)
}
