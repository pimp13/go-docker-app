package controller

import (
	"errors"
	"strings"
	"test/internal/config"
	"test/internal/service"
	"test/pkg/response"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type UploadController struct {
	uploads service.UploadService
	auth    *AuthController
	cfg     config.UploadConfig
}

func NewUploadController(uploads service.UploadService, auth *AuthController, cfg config.UploadConfig) *UploadController {
	return &UploadController{uploads: uploads, auth: auth, cfg: cfg}
}

func (c *UploadController) Register(api *echo.Group) {
	uploads := api.Group("/uploads", c.auth.RequireAuth, middleware.BodyLimit(c.cfg.MaxBytes))
	uploads.POST("", c.Upload)
}

// Upload stores a multipart file and returns its public URL.
//
//	@Summary		Upload file
//	@Description	Accepts multipart form field "file" and optional "folder" (avatars|articles|general)
//	@Tags			uploads
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			file	formData	file	true	"File to upload"
//	@Param			folder	formData	string	false	"Target folder"
//	@Success		201		{object}	response.APIResponse
//	@Failure		400		{object}	response.APIResponse
//	@Failure		401		{object}	response.APIResponse
//	@Router			/api/uploads [post]
func (c *UploadController) Upload(ctx *echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return response.BadRequest(ctx, service.ErrUploadMissingFile)
	}

	folder := strings.TrimSpace(ctx.FormValue("folder"))
	result, err := c.uploads.SaveImage(file, folder)
	if err != nil {
		return writeUploadError(ctx, err)
	}

	return response.Created(ctx, result)
}

func writeUploadError(ctx *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrUploadMissingFile),
		errors.Is(err, service.ErrUploadInvalidType),
		errors.Is(err, service.ErrUploadInvalidFolder):
		return response.BadRequest(ctx, err)
	case errors.Is(err, service.ErrUploadTooLarge):
		return response.Failure(ctx, 413, "PAYLOAD_TOO_LARGE", err.Error())
	default:
		return response.InternalServerError(ctx, err)
	}
}
