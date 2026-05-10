package category

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateCategory(c fiber.Ctx) error {
	var req struct {
		TenantID    string `json:"tenant_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		IsCustom    bool   `json:"is_custom"`
		UserID      string `json:"user_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	cat, err := h.service.CreateCategory(req.TenantID, req.Name, req.Description, req.Type, req.IsCustom, req.UserID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(cat)
}

func (h *Handler) GetCategoryByID(c fiber.Ctx) error {
	id := c.Params("id")
	cat, err := h.service.GetCategoryByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
	}
	return c.JSON(cat)
}

func (h *Handler) ListCategoriesByTenant(c fiber.Ctx) error {
	tenantID := c.Get("X-Tenant-ID")
	cats, err := h.service.ListCategoriesByTenant(tenantID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cats)
}

func (h *Handler) UpdateCategory(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	cat, err := h.service.UpdateCategory(id, req.Name, req.Description).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cat)
}

func (h *Handler) DeleteCategory(c fiber.Ctx) error {
	id := c.Params("id")
	cat, err := h.service.DeleteCategory(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cat)
}
