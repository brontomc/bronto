package api

import (
	"errors"
	"strconv"

	"github.com/brontomc/bronto/agent/instance"
	"github.com/labstack/echo"
)

const idKey = "instanceId"

type Handlers struct {
	Instances *instance.Instances
}

// VerifyInstanceExistsMiddleware verifies that the instance exists and attaches the instance id to the context
func (h *Handlers) VerifyInstanceExistsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(400, "instance not found")
		}
		id := uint32(id64)
		_, err = h.Instances.InstanceStatus(id)
		if err != nil && errors.Is(err, instance.ErrInstanceDoesNotExist) {
			return instance.ErrInstanceDoesNotExist
		} else if err != nil {
			return err
		}

		c.Set(idKey, id)

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}
