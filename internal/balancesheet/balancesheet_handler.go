package balancesheet

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateBalanceSheet(c fiber.Ctx) error {
	var req struct {
		TenantID string `json:"tenant_id"`
		UserID   string `json:"user_id"`
		Year     int    `json:"year"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	bs, err := h.service.CreateBalanceSheet(req.TenantID, req.UserID, req.Year).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(bs)
}

func (h *Handler) GetBalanceSheetByID(c fiber.Ctx) error {
	id := c.Params("id")
	bs, err := h.service.GetBalanceSheetByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "balancesheet not found"})
	}
	return c.JSON(bs)
}

func (h *Handler) UpdateBalanceSheet(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Year int `json:"year"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	bs, err := h.service.UpdateBalanceSheet(id, req.Year).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(bs)
}
