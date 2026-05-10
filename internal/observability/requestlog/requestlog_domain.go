package requestlog

import "time"

type RequestLog struct {
	ID             string
	TenantID       string
	UserID         string
	Method         string
	Path           string
	QueryParams    string
	RequestHeaders string
	RequestBody    string
	ResponseStatus int
	ResponseBody   string
	ResponseTimeMs int
	IPAddress      string
	UserAgent      string
	ErrorMessage   string
	ErrorStack     string
	Version        int
	CreatedAt      time.Time
}

const (
	EventRecorded = "requestlog.recorded"
)

type RequestLogRecordedPayload struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	UserID         string    `json:"user_id"`
	Method         string    `json:"method"`
	Path           string    `json:"path"`
	QueryParams    string    `json:"query_params"`
	RequestHeaders string    `json:"request_headers"`
	RequestBody    string    `json:"request_body"`
	ResponseStatus int       `json:"response_status"`
	ResponseBody   string    `json:"response_body"`
	ResponseTimeMs int       `json:"response_time_ms"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	ErrorMessage   string    `json:"error_message"`
	ErrorStack     string    `json:"error_stack"`
	CreatedAt      time.Time `json:"created_at"`
}
