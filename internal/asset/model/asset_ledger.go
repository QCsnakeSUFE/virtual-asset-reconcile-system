package model

import "time"

type AssetLedger struct {
	ID            int64     `gorm:"primaryKey"`
	TenantID      string    `gorm:"column:tenant_id;size:32;not null"`
	UserID        string    `gorm:"column:user_id;size:64;not null"`
	ItemCode      string    `gorm:"column:item_code;size:64;not null"`
	ChangeType    string    `gorm:"column:change_type;size:20;not null"`
	ChangeAmount  int64     `gorm:"column:change_amount;not null;default:0"`
	BalanceBefore int64     `gorm:"column:balance_before;not null;default:0"`
	BalanceAfter  int64     `gorm:"column:balance_after;not null;default:0"`
	SourceOrderNo string    `gorm:"column:source_order_no;size:64;not null;default:''"`
	Remark        string    `gorm:"column:remark;size:255;not null;default:''"`
	TraceID       string    `gorm:"column:trace_id;size:64;not null;default:''"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

func (AssetLedger) TableName() string {
	return "asset_ledger"
}
