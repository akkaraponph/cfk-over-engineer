package requestlog

import (
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type RequestLogProjection struct {
	ID             string    `gorm:"type:uuid;primaryKey"`
	TenantID       string    `gorm:"type:uuid;index"`
	UserID         string    `gorm:"type:uuid;index"`
	Method         string    `gorm:"size:10;not null"`
	Path           string    `gorm:"size:500;not null"`
	QueryParams    string    `gorm:"type:jsonb"`
	RequestHeaders string    `gorm:"type:jsonb"`
	RequestBody    string    `gorm:"type:jsonb"`
	ResponseStatus int       `gorm:"index"`
	ResponseBody   string    `gorm:"type:jsonb"`
	ResponseTimeMs int
	IPAddress      string    `gorm:"size:45"`
	UserAgent      string    `gorm:"type:text"`
	ErrorMessage   string    `gorm:"type:text"`
	ErrorStack     string    `gorm:"type:text"`
	Version        int       `gorm:"not null;default:1"`
	CreatedAt      time.Time `gorm:"not null;index"`
}

func (RequestLogProjection) TableName() string {
	return "request_log_projections"
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) FindByID(id string) mo.Option[RequestLog] {
	var proj RequestLogProjection
	if err := r.db.Where("id = ?", id).First(&proj).Error; err != nil {
		return mo.None[RequestLog]()
	}
	return mo.Some(toDomain(proj))
}

func (r *GORMRepository) FindAll(limit, offset int) mo.Result[[]RequestLog] {
	var projs []RequestLogProjection
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&projs).Error; err != nil {
		return mo.Err[[]RequestLog](err)
	}
	logs := make([]RequestLog, len(projs))
	for i, p := range projs {
		logs[i] = toDomain(p)
	}
	return mo.Ok(logs)
}

func toDomain(p RequestLogProjection) RequestLog {
	return RequestLog{
		ID:             p.ID,
		TenantID:       p.TenantID,
		UserID:         p.UserID,
		Method:         p.Method,
		Path:           p.Path,
		QueryParams:    p.QueryParams,
		RequestHeaders: p.RequestHeaders,
		RequestBody:    p.RequestBody,
		ResponseStatus: p.ResponseStatus,
		ResponseBody:   p.ResponseBody,
		ResponseTimeMs: p.ResponseTimeMs,
		IPAddress:      p.IPAddress,
		UserAgent:      p.UserAgent,
		ErrorMessage:   p.ErrorMessage,
		ErrorStack:     p.ErrorStack,
		Version:        p.Version,
		CreatedAt:      p.CreatedAt,
	}
}
