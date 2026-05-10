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
		Plan string `json:"plan"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Plan == "" {
		req.Plan = "free"
	}

	t, err := h.svc.CreateTenant(req.Name, req.Slug, req.Plan).Get()
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

func (h *Handler) ChangePlan(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Plan string `json:"plan"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	t, err := h.svc.ChangePlan(id, req.Plan).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}

func (h *Handler) EnableFeature(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Feature string `json:"feature"`
		UserID  string `json:"user_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tf, err := h.svc.EnableFeature(id, req.Feature, req.UserID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(tf)
}

func (h *Handler) DisableFeature(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Feature string `json:"feature"`
		UserID  string `json:"user_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tf, err := h.svc.DisableFeature(id, req.Feature, req.UserID).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tf)
}

func (h *Handler) CheckFeature(c fiber.Ctx) error {
	id := c.Params("id")
	feature := c.Query("feature")
	if feature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "feature query param required"})
	}

	has := h.svc.HasFeature(id, feature)
	return c.JSON(fiber.Map{"tenant_id": id, "feature": feature, "enabled": has})
}
