package handler

import (
	"errors"
	"txn-processor/internal/core/model"
	"txn-processor/internal/core/service"
	"txn-processor/internal/port"

	"github.com/gofiber/fiber/v2"
)

type TransferHandler struct {
	transferService port.TransferService
}

func NewTransferHandler(transferService port.TransferService) *TransferHandler {
	return &TransferHandler{transferService: transferService}
}

func (h *TransferHandler) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req model.TransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "invalid request"})
	}

	res, err := h.transferService.ProcessTransfer(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "account not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}
