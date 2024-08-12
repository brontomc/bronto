package api

import (
	"github.com/brontomc/bronto/agent/instance/state"
	"github.com/labstack/echo"
)

func (h *Handlers) HandleGetInstance(c echo.Context) error {
	id := c.Get(idKey).(uint32)

	instance, err := h.Instances.GetInstance(id)
	if err != nil {
		return err
	}

	resp := echo.Map{
		"status": instance.Status,
	}

	if instance.Status == state.Running || instance.Status == state.Starting {
		resp["startTime"] = instance.StartTime
	}

	return c.JSON(200, resp)
}

func (h *Handlers) HandleStartInstance(c echo.Context) error {
	id := c.Get(idKey).(uint32)

	err := h.Instances.StartInstance(id)
	if err != nil {
		return err
	}

	return c.NoContent(204)
}

func (h *Handlers) HandleStopInstance(c echo.Context) error {
	id := c.Get(idKey).(uint32)

	err := h.Instances.StopInstance(id)
	if err != nil {
		return err
	}

	return c.NoContent(204)
}

func (h *Handlers) HandleRemoveInstance(c echo.Context) error {
	id := c.Get(idKey).(uint32)

	err := h.Instances.RemoveInstance(id)
	if err != nil {
		return err
	}

	return c.NoContent(204)
}
