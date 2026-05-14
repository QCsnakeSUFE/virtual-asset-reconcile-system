package model

import "time"

type OutboxMessage struct {
	ID          int64      `gorm:"primaryKey"`
	TenantID    string     `gorm:"column:tenant_id;size:32;not null"`
	BizNo       string     `gorm:"column:biz_no;size:64;not null"`
	EventType   string     `gorm:"column:event_type;size:64;not null"`
	Status      string     `gorm:"column:status;size:20;not null;default:PENDING"`
	Payload     string     `gorm:"column:payload;type:json;not null"`
	RetryCount  int        `gorm:"column:retry_count;not null;default:0"`
	MaxRetries  int        `gorm:"column:max_retries;not null;default:3"`
	NextRetryAt time.Time  `gorm:"column:next_retry_at;not null;autoCreateTime"`
	TraceID     string     `gorm:"column:trace_id;size:64;not null;default:''"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	SentAt      *time.Time `gorm:"column:sent_at"`
	LastRetryAt *time.Time `gorm:"column:last_retry_at"`
}

func (OutboxMessage) TableName() string {
	return "outbox_messages"
}
