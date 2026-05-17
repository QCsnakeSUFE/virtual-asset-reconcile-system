package service

import (
	"context"
	"encoding/json"
	"virtual-asset-reconcile-system/internal/transaction/model"

	"gorm.io/gorm"
)

func OutboxAssetHandler(ctx context.Context, db *gorm.DB, msg model.OutboxMessage) error {
	type payload struct {
		TenantID      string `json:"tenant_id"`
		UserID        string `json:"user_id"`
		ItemCode      string `json:"item_code"`
		Quantity      int    `json:"quantity"`
		SourceOrderNo string `json:"source_order_no"`
	}
	var pl payload
	err := json.Unmarshal([]byte(msg.Payload), &pl)
	if err != nil {
		return err
	}
	req := GrantAssetRequest{
		TenantID:      pl.TenantID,
		UserID:        pl.UserID,
		ItemCode:      pl.ItemCode,
		Quantity:      int64(pl.Quantity),
		SourceOrderNo: pl.SourceOrderNo,
	}
	_, err = GrantAsset(ctx, db, req)
	if err != nil {
		return err
	}
	return nil
}

func OutboxDeductHandler(ctx context.Context, db *gorm.DB, msg model.OutboxMessage) error {
	type payload struct {
		TenantID      string `json:"tenant_id"`
		UserID        string `json:"user_id"`
		ItemCode      string `json:"item_code"`
		Quantity      int    `json:"quantity"`
		SourceOrderNo string `json:"source_order_no"`
	}
	var pl payload
	err := json.Unmarshal([]byte(msg.Payload), &pl)
	if err != nil {
		return err
	}
	req := GrantAssetRequest{
		TenantID:      pl.TenantID,
		UserID:        pl.UserID,
		ItemCode:      pl.ItemCode,
		Quantity:      int64(pl.Quantity),
		SourceOrderNo: pl.SourceOrderNo,
	}
	_, err = DeductAsset(ctx, db, req)
	if err != nil {
		return err
	}
	return nil
}
