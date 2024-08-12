package api

import (
	"errors"

	"github.com/brontomc/bronto/agent/instance"
	"github.com/labstack/echo"
)

func ErrorHandler(defaultHandler func(err error, c echo.Context)) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		if errors.Is(err, instance.ErrInstanceDoesNotExist) {
			c.JSON(404, map[string]string{"error": "instance does not exist"})
			return
		}
		if errors.Is(err, instance.ErrInstanceAlreadyExists) {
			c.JSON(409, map[string]string{"error": "instance already exists"})
			return
		}
		if errors.Is(err, instance.ErrInstanceIsRunning) {
			c.JSON(409, map[string]string{"error": "instance is running"})
			return
		}
		if errors.Is(err, instance.ErrInstanceIsNotRunning) {
			c.JSON(409, map[string]string{"error": "instance is not running"})
			return
		}

		defaultHandler(err, c)
	}
}
