package requestlog

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	ErrInvalidMethod = errors.New("invalid HTTP method")
	ErrInvalidPath   = errors.New("invalid path")
	ErrNotFound      = errors.New("request log not found")
)

type Service struct {
	repo     Repository
	eventBus *event.Bus
}

func NewService(repo Repository, eventBus *event.Bus) *Service {
	return &Service{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *Service) RecordRequestLog(tenantID, userID, method, path, queryParams, requestHeaders, requestBody string, responseStatus int, responseBody string, responseTimeMs int, ipAddress, userAgent, errorMessage, errorStack string) mo.Result[RequestLog] {
	if method == "" {
		return mo.Err[RequestLog](ErrInvalidMethod)
	}
	if path == "" {
		return mo.Err[RequestLog](ErrInvalidPath)
	}

	id := uuid.New().String()
	now := time.Now()

	evt := event.Event{
		AggregateType: "requestlog",
		AggregateID:   id,
		EventType:     EventRecorded,
		Version:       1,
		Payload: RequestLogRecordedPayload{
			ID:             id,
			TenantID:       tenantID,
			UserID:         userID,
			Method:         method,
			Path:           path,
			QueryParams:    queryParams,
			RequestHeaders: requestHeaders,
			RequestBody:    requestBody,
			ResponseStatus: responseStatus,
			ResponseBody:   responseBody,
			ResponseTimeMs: responseTimeMs,
			IPAddress:      ipAddress,
			UserAgent:      userAgent,
			ErrorMessage:   errorMessage,
			ErrorStack:     errorStack,
			CreatedAt:      now,
		},
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[RequestLog](r.Error())
	}

	return mo.Ok(RequestLog{
		ID:             id,
		TenantID:       tenantID,
		UserID:         userID,
		Method:         method,
		Path:           path,
		QueryParams:    queryParams,
		RequestHeaders: requestHeaders,
		RequestBody:    requestBody,
		ResponseStatus: responseStatus,
		ResponseBody:   responseBody,
		ResponseTimeMs: responseTimeMs,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ErrorMessage:   errorMessage,
		ErrorStack:     errorStack,
		CreatedAt:      now,
	})
}

func (s *Service) GetRequestLogByID(id string) mo.Result[RequestLog] {
	opt := s.repo.FindByID(id)
	rl, ok := opt.Get()
	if !ok {
		return mo.Err[RequestLog](ErrNotFound)
	}
	return mo.Ok(rl)
}

func (s *Service) ListRequestLogs(limit, offset int) mo.Result[[]RequestLog] {
	return s.repo.FindAll(limit, offset)
}
