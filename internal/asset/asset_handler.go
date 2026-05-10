package asset

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RecordAsset(c fiber.Ctx) error {
	var req struct {
		TenantID        string  `json:"tenant_id"`
		Type            string  `json:"type"`
		Description     string  `json:"description"`
		UserID          string  `json:"user_id"`
		Value           float64 `json:"value"`
		CashflowPerYear float64 `json:"cashflow_per_year"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	asset, err := h.service.RecordAsset(req.TenantID, req.Type, req.Description, req.UserID, req.Value, req.CashflowPerYear).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(asset)
}

func (h *Handler) GetAssetByID(c fiber.Ctx) error {
	id := c.Params("id")
	asset, err := h.service.GetAssetByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "asset not found"})
	}
	return c.JSON(asset)
}

func (h *Handler) ChangeValue(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Value           float64 `json:"value"`
		CashflowPerYear float64 `json:"cashflow_per_year"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	asset, err := h.service.ChangeValue(id, req.Value, req.CashflowPerYear).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(asset)
}

func (h *Handler) AssignToBalanceSheet(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		BalanceSheetID string `json:"balance_sheet_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	asset, err := h.service.AssignToBalanceSheet(id, req.BalanceSheetID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(asset)
}

func (h *Handler) UnassignFromBalanceSheet(c fiber.Ctx) error {
	id := c.Params("id")
	asset, err := h.service.UnassignFromBalanceSheet(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(asset)
}
