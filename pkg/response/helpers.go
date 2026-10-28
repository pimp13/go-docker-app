package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// OK is a shortcut for HTTP 200.
func OK(c *echo.Context, data any) error {
	return Success(c, http.StatusOK, data)
}

// Created is a shortcut for HTTP 201.
func Created(c *echo.Context, data any) error {
	return Success(c, http.StatusCreated, data)
}

// InternalServerError is a shortcut for HTTP 500.
func InternalServerError(c *echo.Context, err error) error {
	return Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", err.Error())
}

// BadRequest is a shortcut for HTTP 400.
func BadRequest(c *echo.Context, err error) error {
	return Failure(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
}

// Unauthorized is a shortcut for HTTP 401.
func Unauthorized(c *echo.Context, err error) error {
	return Failure(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
}

// Forbidden is a shortcut for HTTP 403.
func Forbidden(c *echo.Context, err error) error {
	return Failure(c, http.StatusForbidden, "FORBIDDEN", err.Error())
}

// NotFound is a shortcut for HTTP 404.
func NotFound(c *echo.Context, err error) error {
	return Failure(c, http.StatusNotFound, "NOT_FOUND", err.Error())
}

// MethodNotAllowed is a shortcut for HTTP 405.
func MethodNotAllowed(c *echo.Context, err error) error {
	return Failure(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", err.Error())
}

// RequestTimeout is a shortcut for HTTP 408.
func RequestTimeout(c *echo.Context, err error) error {
	return Failure(c, http.StatusRequestTimeout, "REQUEST_TIMEOUT", err.Error())
}

// Conflict is a shortcut for HTTP 409.
func Conflict(c *echo.Context, err error) error {
	return Failure(c, http.StatusConflict, "CONFLICT", err.Error())
}
