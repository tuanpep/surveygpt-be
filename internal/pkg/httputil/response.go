package httputil

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// OK responds with 200 and wraps data in {"data": ...}.
func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, echo.Map{"data": data})
}

// Created responds with 201 and wraps data in {"data": ...}.
func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, echo.Map{"data": data})
}

// Paginated responds with 200, data, and pagination metadata.
func Paginated(c echo.Context, data any, nextCursor string, total int) error {
	return c.JSON(http.StatusOK, echo.Map{
		"data":        data,
		"next_cursor": nextCursor,
		"total":       total,
	})
}

// Message responds with 200 and a message.
func Message(c echo.Context, msg string) error {
	return c.JSON(http.StatusOK, echo.Map{"message": msg})
}
