package debt

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RecordDebt(c fiber.Ctx) error {
	var req struct {
		TenantID    string  `json:"tenant_id"`
		Type        string  `json:"type"`
		Description string  `json:"description"`
		UserID      string  `json:"user_id"`
		Amount      float64 `json:"amount"`
		Interest    float64 `json:"interest"`
		MinimumPay  float64 `json:"minimum_pay"`
		Priority    int     `json:"priority"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	debt, err := h.service.RecordDebt(req.TenantID, req.Type, req.Description, req.UserID, req.Amount, req.Interest, req.MinimumPay, req.Priority).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(debt)
}

func (h *Handler) GetDebtByID(c fiber.Ctx) error {
	id := c.Params("id")
	debt, err := h.service.GetDebtByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "debt not found"})
	}
	return c.JSON(debt)
}

func (h *Handler) ChangeAmount(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	debt, err := h.service.ChangeAmount(id, req.Amount).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(debt)
}

func (h *Handler) AssignToBalanceSheet(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		BalanceSheetID string `json:"balance_sheet_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	debt, err := h.service.AssignToBalanceSheet(id, req.BalanceSheetID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(debt)
}

func (h *Handler) UnassignFromBalanceSheet(c fiber.Ctx) error {
	id := c.Params("id")
	debt, err := h.service.UnassignFromBalanceSheet(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(debt)
}
