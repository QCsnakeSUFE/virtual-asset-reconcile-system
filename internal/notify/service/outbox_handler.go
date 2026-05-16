package service

import (
	"context"
	"encoding/json"
	"virtual-asset-reconcile-system/internal/transaction/model"

	"gorm.io/gorm"
)

func OutboxNotifyHandler(ctx context.Context, db *gorm.DB, msg model.OutboxMessage) error {
	type payload struct {
		TenantID string `json:"tenant_id"`
		UserID   string `json:"user_id"`
		OrderNo  string `json:"order_no"`
		Msg      string `json:"msg"`
	}
	var pl payload
	err := json.Unmarshal([]byte(msg.Payload), &pl)
	if err != nil {
		return err
	}
	req := SendNotifyRequest{
		TenantID: pl.TenantID,
		UserID:   pl.UserID,
		OrderNo:  pl.OrderNo,
		Title:    "",
		Content:  pl.Msg,
		Channel:  "",
	}
	_, err = SendNotify(ctx, db, req)
	if err != nil {
		return err
	}
	return nil
}
