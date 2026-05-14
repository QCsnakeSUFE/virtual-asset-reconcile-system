package model

import "time"

type Asset struct {
	ID            int64     `gorm:"primaryKey"`
	TenantID      string    `gorm:"column:tenant_id;size:32;not null"`
	UserID        string    `gorm:"column:user_id;size:64;not null"`
	ItemCode      string    `gorm:"column:item_code;size:64;not null"`
	Quantity      int64     `gorm:"column:quantity;not null;default:0"`
	Frozen        int64     `gorm:"column:frozen;not null;default:0"`
	TotalGranted  int64     `gorm:"column:total_granted;not null;default:0"`
	TotalConsumed int64     `gorm:"column:total_consumed;not null;default:0"`
	Version       int       `gorm:"column:version;not null;default:0"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (Asset) TableName() string {
	return "assets"
}
