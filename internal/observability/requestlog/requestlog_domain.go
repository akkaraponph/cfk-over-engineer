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
	CreatedAt      time.Time
}

const (
	EventRecorded = "requestlog.recorded"
)
