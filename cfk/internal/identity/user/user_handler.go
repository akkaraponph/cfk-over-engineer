package user

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterUser(c fiber.Ctx) error {
	var req struct {
		TenantID  string `json:"tenant_id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		Role      string `json:"role"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.service.RegisterUser(req.TenantID, req.Username, req.Email, req.Password, req.FirstName, req.LastName, req.Phone, req.Role).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *Handler) GetUserByEmail(c fiber.Ctx) error {
	tenantID := c.Get("X-Tenant-ID")
	email := c.Params("email")
	user, err := h.service.GetUserByEmail(tenantID, email).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(user)
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req struct {
		TenantID string `json:"tenant_id"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	tokenResp, err := h.service.Login(req.TenantID, req.Email, req.Password).Get()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tokenResp)
}

func (h *Handler) ActivateUser(c fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.service.ActivateUser(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}

func (h *Handler) DeactivateUser(c fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.service.DeactivateUser(id).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}

func (h *Handler) ChangeRole(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Role string `json:"role"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.service.ChangeRole(id, req.Role).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}

func (h *Handler) UpdateProfile(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.service.UpdateProfile(id, req.FirstName, req.LastName, req.Phone).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}
