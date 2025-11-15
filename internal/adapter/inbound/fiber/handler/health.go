package handler

import (
	"txn-processor/internal/port"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	healthService port.HealthService
}

func NewHealthHandler(healthService port.HealthService) *HealthHandler {
	return &HealthHandler{healthService}
}

func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	ctx := c.UserContext()

	err := h.healthService.Check(ctx)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": false, "err": err})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true})
}
