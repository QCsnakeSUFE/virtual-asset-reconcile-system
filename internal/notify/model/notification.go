package model

import "time"

type Notification struct {
	ID        int64      `gorm:"primaryKey"`
	TenantID  string     `gorm:"column:tenant_id;size:32;not null"`
	UserID    string     `gorm:"column:user_id;size:64;not null"`
	BizNo     string     `gorm:"column:biz_no;size:64;not null"`
	EventType string     `gorm:"column:event_type;size:64;not null"`
	Channel   string     `gorm:"column:channel;size:32;not null;default:''"`
	Status    string     `gorm:"column:status;size:20;not null;default:PENDING"`
	Title     string     `gorm:"column:title;size:128;not null;default:''"`
	Content   string     `gorm:"column:content;size:512;not null;default:''"`
	Result    string     `gorm:"column:result;size:255;not null;default:''"`
	TraceID   string     `gorm:"column:trace_id;size:64;not null;default:''"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	SentAt    *time.Time `gorm:"column:sent_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
