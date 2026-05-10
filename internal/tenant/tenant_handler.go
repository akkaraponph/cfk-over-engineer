package tenant

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateTenant(c fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	t, err := h.svc.CreateTenant(req.Name, req.Slug).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(t)
}

func (h *Handler) GetTenantBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	t, err := h.svc.GetTenantBySlug(slug).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "tenant not found"})
	}
	return c.JSON(t)
}

func (h *Handler) ActivateTenant(c fiber.Ctx) error {
	id := c.Params("id")
	t, err := h.svc.ActivateTenant(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}

func (h *Handler) DeactivateTenant(c fiber.Ctx) error {
	id := c.Params("id")
	t, err := h.svc.DeactivateTenant(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}
