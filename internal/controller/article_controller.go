package controller

import (
	"errors"
	"test/internal/dto"
	"test/internal/service"
	"test/pkg/response"
	"test/pkg/token"

	"github.com/labstack/echo/v5"
)

type ArticleController struct {
	articleService service.ArticleService
}

func NewArticleController(articleService service.ArticleService) *ArticleController {
	return &ArticleController{articleService}
}

func (c *ArticleController) Register(api *echo.Group) {
	api.GET("/articles", c.GetArticles)
	api.POST("/articles/create", c.CreateArticle)
	api.GET("/articles/:id", c.GetArticleByID)
}

func (c *ArticleController) CreateArticle(ctx *echo.Context) error {
	tk := token.ExtractToken(ctx)
	if tk == "" {
		return response.Unauthorized(ctx, errors.New("token is required"))
	}
	if err := token.ValidateToken(tk, c.articleService.SecretKey()); err != nil {
		return response.Unauthorized(ctx, err)
	}

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

func (c *ArticleController) GetArticles(ctx *echo.Context) error {
	tk := token.ExtractToken(ctx)
	if tk == "" {
		return response.Unauthorized(ctx, errors.New("token is required"))
	}
	if err := token.ValidateToken(tk, c.articleService.SecretKey()); err != nil {
		return response.Unauthorized(ctx, err)
	}

	result, err := c.articleService.GetArticles()
	if err != nil {
		return response.InternalServerError(ctx, err)
	}

	return response.OK(ctx, result)
}

func (c *ArticleController) GetArticleByID(ctx *echo.Context) error {
	tk := token.ExtractToken(ctx)
	if tk == "" {
		return response.Unauthorized(ctx, errors.New("token is required"))
	}
	if err := token.ValidateToken(tk, c.articleService.SecretKey()); err != nil {
		return response.Unauthorized(ctx, err)
	}

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
