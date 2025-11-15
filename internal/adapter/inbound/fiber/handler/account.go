package handler

import (
	"errors"
	"strconv"
	"txn-processor/internal/core/model"
	"txn-processor/internal/core/service"
	"txn-processor/internal/port"

	"github.com/gofiber/fiber/v2"
)

type AccountHandler struct {
	accountService port.AccountService
}

func NewAccountHandler(accountService port.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

func (h *AccountHandler) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req model.AccountCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "invalid request"})
	}

	res, err := h.accountService.CreateAccount(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, service.ErrConflict):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *AccountHandler) Get(c *fiber.Ctx) error {
	ctx := c.UserContext()

	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "invalid account id"})
	}

	res, err := h.accountService.GetAccount(ctx, id)
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

	return c.Status(fiber.StatusOK).JSON(res)
}
