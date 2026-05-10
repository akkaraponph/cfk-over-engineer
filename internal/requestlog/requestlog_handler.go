package requestlog

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RecordRequestLog(c fiber.Ctx) error {
	var req struct {
		TenantID       string `json:"tenant_id"`
		UserID         string `json:"user_id"`
		Method         string `json:"method"`
		Path           string `json:"path"`
		QueryParams    string `json:"query_params"`
		RequestHeaders string `json:"request_headers"`
		RequestBody    string `json:"request_body"`
		ResponseStatus int    `json:"response_status"`
		ResponseBody   string `json:"response_body"`
		ResponseTimeMs int    `json:"response_time_ms"`
		IPAddress      string `json:"ip_address"`
		UserAgent      string `json:"user_agent"`
		ErrorMessage   string `json:"error_message"`
		ErrorStack     string `json:"error_stack"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rl, err := h.service.RecordRequestLog(
		req.TenantID, req.UserID, req.Method, req.Path,
		req.QueryParams, req.RequestHeaders, req.RequestBody,
		req.ResponseStatus, req.ResponseBody, req.ResponseTimeMs,
		req.IPAddress, req.UserAgent, req.ErrorMessage, req.ErrorStack,
	).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(rl)
}

func (h *Handler) GetRequestLogByID(c fiber.Ctx) error {
	id := c.Params("id")
	rl, err := h.service.GetRequestLogByID(id).Get()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request log not found"})
	}
	return c.JSON(rl)
}

func (h *Handler) ListRequestLogs(c fiber.Ctx) error {
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	logs, err := h.service.ListRequestLogs(limit, offset).Get()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(logs)
}
