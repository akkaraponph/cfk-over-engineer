package cashflowout

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RecordCashflowOut(c fiber.Ctx) error {
	var req struct {
		TenantID    string  `json:"tenant_id"`
		UserID      string  `json:"user_id"`
		PocketID    string  `json:"pocket_id"`
		CategoryID  int     `json:"category_id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		Receipt     string  `json:"receipt"`
		Type        string  `json:"type"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	cf, err := h.service.RecordCashflowOut(req.TenantID, req.UserID, req.PocketID, req.CategoryID, req.Amount, req.Description, req.Receipt, req.Type).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(cf)
}

func (h *Handler) GetCashflowOutByID(c fiber.Ctx) error {
	id := c.Params("id")
	cf, err := h.service.GetCashflowOutByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "cashflowout not found"})
	}
	return c.JSON(cf)
}

func (h *Handler) ListCashflowOutsByPocket(c fiber.Ctx) error {
	tenantID := c.Get("X-Tenant-ID")
	pocketID := c.Params("pocketId")
	cfs, err := h.service.ListCashflowOutsByPocket(tenantID, pocketID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cfs)
}

func (h *Handler) UpdateCashflowOut(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		CategoryID  int     `json:"category_id"`
		Receipt     string  `json:"receipt"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	cf, err := h.service.UpdateCashflowOut(id, req.Amount, req.Description, req.CategoryID, req.Receipt).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cf)
}

func (h *Handler) DeleteCashflowOut(c fiber.Ctx) error {
	id := c.Params("id")
	cf, err := h.service.DeleteCashflowOut(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cf)
}
