package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"virtual-asset-reconcile-system/internal/transaction/model"
	"virtual-asset-reconcile-system/pkg/idgen"

	"gorm.io/gorm"
)

type PaymentCallbackRequest struct {
	// 用户支付之后，调用POST /api/v1/payment/callback，通知后端“已付款”
	OrderNo  string
	TenantID string
}

type OutboxPayload struct {
	// ASSET_GRANT
	TenantID      string `json:"tenant_id"`
	UserID        string `json:"user_id"`
	ItemCode      string `json:"item_code"`
	Quantity      int    `json:"quantity"`
	SourceOrderNo string `json:"source_order_no"`
}

func ProcessPaymentCallback(ctx context.Context, db *gorm.DB, req PaymentCallbackRequest) error {
	var order model.Order
	err := db.WithContext(ctx).Where("tenant_id = ? AND order_no = ?", req.TenantID, req.OrderNo).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order not found")
		}
		return err
	}
	// 查到了，继续检查订单状态
	if order.Status != "CREATED" {
		return errors.New("invalid order status: " + order.Status)
	}

	// 第二步：查订单商品信息，用来构建 outbox payload
	var items []model.OrderItem
	if err = db.WithContext(ctx).Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
		return err
	}

	// 第三步：本地事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 3.1 更新订单状态 CREATED->PAID
		if err := tx.Model(&order).Update("status", "PAID").Error; err != nil {
			return err
		}

		// 3.2 遍历每一条明细，生成 ASSET_GRANT 消息
		for _, item := range items {
			payload := OutboxPayload{
				TenantID:      req.TenantID,
				UserID:        order.UserID,
				ItemCode:      item.ItemCode,
				Quantity:      item.Quantity,
				SourceOrderNo: req.OrderNo,
			}
			payloadBytes, _ := json.Marshal(payload)

			msg := model.OutboxMessage{
				ID:          idgen.NextID(),
				TenantID:    req.TenantID,
				BizNo:       req.OrderNo,
				EventType:   "ASSET_GRANT",
				Payload:     string(payloadBytes),
				Status:      "PENDING",
				MaxRetries:  3,
				NextRetryAt: time.Now(),
			}
			if err = tx.Create(&msg).Error; err != nil {
				return err
			}
		}
		// 3.3 再加一条 NOTICE_SEND 消息
		noticePayload, _ := json.Marshal(map[string]string{
			"tenant_id": req.TenantID,
			"user_id":   order.UserID,
			"order_no":  req.OrderNo,
			"msg":       "您的订单已支付成功，资产已发放。",
		})
		noticeMsg := model.OutboxMessage{
			ID:          idgen.NextID(),
			TenantID:    req.TenantID,
			BizNo:       req.OrderNo,
			EventType:   "NOTICE_SEND",
			Payload:     string(noticePayload),
			Status:      "PENDING",
			MaxRetries:  3,
			NextRetryAt: time.Now(),
		}
		if err := tx.Create(&noticeMsg).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}
