package model

import "time"

type OrderItem struct {
	ID         int64     `gorm:"primaryKey"`
	TenantID   string    `gorm:"column:tenant_id;size:32;not null"`
	OrderID    int64     `gorm:"column:order_id;not null"`
	ItemCode   string    `gorm:"column:item_code;size:64;not null"`
	ItemName   string    `gorm:"column:item_name;size:128;not null;default:''"`
	Quantity   int       `gorm:"column:quantity;not null;default:1"`
	UnitPrice  int64     `gorm:"column:unit_price;not null;default:0"`
	TotalPrice int64     `gorm:"column:total_price;not null;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
