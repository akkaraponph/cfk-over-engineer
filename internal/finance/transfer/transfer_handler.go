package transfer

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitiateTransfer(c fiber.Ctx) error {
	var req struct {
		TenantID     string  `json:"tenant_id"`
		UserID       string  `json:"user_id"`
		FromPocketID string  `json:"from_pocket_id"`
		ToPocketID   string  `json:"to_pocket_id"`
		Amount       float64 `json:"amount"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	t, err := h.service.InitiateTransfer(req.TenantID, req.UserID, req.FromPocketID, req.ToPocketID, req.Amount).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(t)
}

func (h *Handler) GetTransferByID(c fiber.Ctx) error {
	id := c.Params("id")
	t, err := h.service.GetTransferByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "transfer not found"})
	}
	return c.JSON(t)
}

func (h *Handler) CompleteTransfer(c fiber.Ctx) error {
	id := c.Params("id")
	t, err := h.service.CompleteTransfer(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}

func (h *Handler) FailTransfer(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	t, err := h.service.FailTransfer(id, req.Reason).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}

func (h *Handler) DeleteTransfer(c fiber.Ctx) error {
	id := c.Params("id")
	t, err := h.service.DeleteTransfer(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}
