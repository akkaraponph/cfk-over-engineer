package pocket

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreatePocket(c fiber.Ctx) error {
	var req struct {
		TenantID string `json:"tenant_id"`
		Name     string `json:"name"`
		UserID   string `json:"user_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	pocket, err := h.service.CreatePocket(req.TenantID, req.Name, req.UserID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(pocket)
}

func (h *Handler) GetPocketByID(c fiber.Ctx) error {
	id := c.Params("id")
	pocket, err := h.service.GetPocketByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "pocket not found"})
	}
	return c.JSON(pocket)
}

func (h *Handler) ListPocketsByUser(c fiber.Ctx) error {
	tenantID := c.Get("X-Tenant-ID")
	userID := c.Params("userId")
	pockets, err := h.service.ListPocketsByUser(tenantID, userID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(pockets)
}

func (h *Handler) ChangeName(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	pocket, err := h.service.ChangeName(id, req.Name).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(pocket)
}

func (h *Handler) ChangeBalance(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	pocket, err := h.service.ChangeBalance(id, req.Amount).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(pocket)
}

func (h *Handler) DeletePocket(c fiber.Ctx) error {
	id := c.Params("id")
	pocket, err := h.service.DeletePocket(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(pocket)
}
