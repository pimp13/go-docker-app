package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// CustomValidator adapts go-playground/validator to Echo's Validator interface.
type CustomValidator struct {
	validate *validator.Validate
}

// New returns an Echo-compatible validator.
func New() echo.Validator {
	return &CustomValidator{validate: validator.New()}
}

// Validate runs struct tag validation (e.g. `validate:"required"`).
func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validate.Struct(i); err != nil {
		return err
	}
	return nil
}
