package middleware

import (
	"cfk/internal/identity/tenant"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func TenantMiddleware(tenantService *tenant.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		slug := c.Get("X-Tenant-Slug")
		if slug == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "X-Tenant-Slug header is required",
			})
		}

		t, err := tenantService.GetTenantBySlug(slug).Get()
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tenant not found",
			})
		}

		if !t.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "tenant is deactivated",
			})
		}

		c.Locals("tenant_id", t.ID)
		c.Locals("tenant_slug", t.Slug)
		c.Locals("tenant_plan", string(t.Plan))
		return c.Next()
	}
}

func FeatureGuard(tenantService *tenant.Service, feature string) fiber.Handler {
	return func(c fiber.Ctx) error {
		tenantID, ok := c.Locals("tenant_id").(string)
		if !ok || tenantID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "tenant context not found",
			})
		}

		if !tenantService.HasFeature(tenantID, feature) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "feature not enabled for tenant: " + feature,
			})
		}

		c.Locals("feature_"+feature, true)
		return c.Next()
	}
}

func AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header with Bearer token is required",
			})
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		c.Locals("token", token)
		return c.Next()
	}
}

func RequestLoggerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		return err
	}
}
