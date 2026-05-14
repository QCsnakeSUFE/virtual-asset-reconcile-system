package model

import "time"

type Order struct {
	ID            int64     `gorm:"primaryKey"`
	TenantID      string    `gorm:"column:tenant_id;size:32;not null"`
	UserID        string    `gorm:"column:user_id;size:64;not null"`
	OrderNo       string    `gorm:"column:order_no;size:64;not null"`
	TotalAmount   int64     `gorm:"column:total_amount;not null;default:0"`
	Status        string    `gorm:"column:status;size:20;not null;default:CREATED"`
	IdempotentKey string    `gorm:"column:idempotent_key;size:128;not null"`
	TraceID       string    `gorm:"column:trace_id;size:64;not null;default:''"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (Order) TableName() string {
	return "orders"
}
