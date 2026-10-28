package response

import (
	"github.com/labstack/echo/v5"
)

type APIResponse struct {
	Result Result `json:"result"`
}

type Result struct {
	Success bool           `json:"success"`
	Data    any            `json:"data"`
	Error   *ErrorDetails  `json:"error"`
	Request RequestDetails `json:"request"`
}

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type RequestDetails struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

// Success returns a standardized successful response.
func Success(c *echo.Context, status int, data any) error {
	return c.JSON(status, APIResponse{
		Result: Result{
			Success: true,
			Data:    data,
			Error:   nil,
			Request: getRequestDetails(c),
		},
	})
}

// Failure returns a standardized error response.
func Failure(
	c *echo.Context,
	status int,
	code string,
	message string,
) error {
	return c.JSON(status, APIResponse{
		Result: Result{
			Success: false,
			Data:    nil,
			Error: &ErrorDetails{
				Code:    code,
				Message: message,
			},
			Request: getRequestDetails(c),
		},
	})
}

// getRequestDetails extracts request metadata.
func getRequestDetails(c *echo.Context) RequestDetails {
	return RequestDetails{
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}
}
