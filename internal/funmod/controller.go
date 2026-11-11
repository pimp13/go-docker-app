package funmod

import (
	"test/pkg/response"

	"github.com/labstack/echo/v5"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service}
}

func (c *Controller) Register(api *echo.Group) {
	group := api.Group("/funmod")
	group.GET("/:param", c.Funmod)
}

// Funmod funmod.
//
//	@Summary		Funmod
//	@Tags			  funmod
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			param	path		string	true	"Param"
//	@Success		200	{object}	response.APIResponse
//	@Failure		401	{object}	response.APIResponse
//	@Failure		403	{object}	response.APIResponse
//	@Router			/api/funmod/{param} [get]
func (c *Controller) Funmod(ctx *echo.Context) error {
	param := ctx.Param("param")
	result, err := c.service.Funmod(param)
	if err != nil {
		return response.InternalServerError(ctx, err)
	}

	return response.OK(ctx, result)
}
