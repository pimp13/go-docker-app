package funmod

import (
	"test/internal/config"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type API struct {
	service Service
	cfg     *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *API {
	return &API{service: NewService(db), cfg: cfg}
}

func (a *API) Service() Service {
	return a.service
}

func (a *API) Register(api *echo.Group) {
	NewController(a.service).Register(api)
}
