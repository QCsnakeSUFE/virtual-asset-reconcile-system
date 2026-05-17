package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"virtual-asset-reconcile-system/internal/transaction/model"
	"virtual-asset-reconcile-system/pkg/idgen"
	"virtual-asset-reconcile-system/pkg/metrics"

	"gorm.io/gorm"
)

type CreateOrderRequest struct {
	TenantID      string
	UserID        string
	IdempotentKey string
	Items         []OrderItemParam
}

type CreateDirectOrderRequest struct {
	TenantID      string
	UserID        string
	SourceUserID  string
	IdempotentKey string
	OrderType     string
	Items         []NoPaymentOrderItemParam
}

type OrderItemParam struct {
	ItemCode  string
	ItemName  string
	Quantity  int
	UnitPrice int64
}

type NoPaymentOrderItemParam struct {
	ItemCode string
	ItemName string
	Quantity int
}

type OrderResult struct {
	OrderNo     string `json:"order_no"`
	TotalAmount int64  `json:"total_amount"`
	Status      string `json:"status"`
}

func CreateOrder(ctx context.Context, db *gorm.DB, req CreateOrderRequest) (*OrderResult, error) {
	var existing model.Order

	err := db.Where("tenant_id = ? AND idempotent_key = ?", req.TenantID, req.IdempotentKey).First(&existing).Error
	// existing
	if err == nil {
		res := &OrderResult{
			OrderNo:     existing.OrderNo,
			TotalAmount: existing.TotalAmount,
			Status:      existing.Status,
		}
		return res, nil
	}
	// err != nil, check error type
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// create one, OrderTotal++
		orderID := idgen.NextID()
		var totalAmount int64
		var orderItems []model.OrderItem
		for _, item := range req.Items {
			totalPrice := int64(item.Quantity) * item.UnitPrice
			totalAmount += totalPrice
			orderItem := model.OrderItem{
				ID:         idgen.NextID(),
				TenantID:   req.TenantID,
				OrderID:    orderID,
				ItemCode:   item.ItemCode,
				ItemName:   item.ItemName,
				Quantity:   item.Quantity,
				UnitPrice:  item.UnitPrice,
				TotalPrice: totalPrice,
			}
			orderItems = append(orderItems, orderItem)
		}
		order := &model.Order{
			ID:            orderID,
			TenantID:      req.TenantID,
			UserID:        req.UserID,
			OrderNo:       fmt.Sprintf("%d", orderID),
			TotalAmount:   totalAmount,
			Status:        "CREATED",
			IdempotentKey: req.IdempotentKey,
			TraceID:       "",
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(order).Error; err != nil {
				return err
			}
			if err := tx.Create(orderItems).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			metrics.OrderFailedTotal.Inc()
			return nil, err
		}
		res := &OrderResult{
			OrderNo:     fmt.Sprintf("%d", orderID),
			TotalAmount: totalAmount,
		}
		metrics.OrderTotal.Inc()
		return res, nil
	} else {
		metrics.OrderFailedTotal.Inc()
		return nil, err
	}
}

func CreateDirectOrder(ctx context.Context, db *gorm.DB, req CreateDirectOrderRequest) (*OrderResult, error) {
	var existing model.Order
	err := db.WithContext(ctx).Where("tenant_id = ? AND idempotent_key = ?", req.TenantID, req.IdempotentKey).First(&existing).Error
	if err == nil {
		return &OrderResult{
			OrderNo:     existing.OrderNo,
			TotalAmount: existing.TotalAmount,
			Status:      existing.Status,
		}, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	orderID := idgen.NextID()
	var orderItems []model.OrderItem
	for _, item := range req.Items {
		orderItems = append(orderItems, model.OrderItem{
			ID:         idgen.NextID(),
			TenantID:   req.TenantID,
			OrderID:    orderID,
			ItemCode:   item.ItemCode,
			ItemName:   item.ItemName,
			Quantity:   item.Quantity,
			UnitPrice:  0,
			TotalPrice: 0,
		})
	}

	order := &model.Order{
		ID:            orderID,
		TenantID:      req.TenantID,
		UserID:        req.UserID,
		OrderNo:       fmt.Sprintf("%d", orderID),
		TotalAmount:   0,
		Status:        "PAID",
		IdempotentKey: req.IdempotentKey,
		TraceID:       "",
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		if err := tx.Create(orderItems).Error; err != nil {
			return err
		}
		for _, item := range req.Items {
			payload := OutboxPayload{
				TenantID:      req.TenantID,
				UserID:        req.UserID,
				ItemCode:      item.ItemCode,
				Quantity:      item.Quantity,
				SourceOrderNo: order.OrderNo,
			}
			payloadBytes, _ := json.Marshal(payload)
			msg := model.OutboxMessage{
				ID:          idgen.NextID(),
				TenantID:    req.TenantID,
				BizNo:       order.OrderNo,
				EventType:   "ASSET_GRANT",
				Payload:     string(payloadBytes),
				Status:      "PENDING",
				MaxRetries:  3,
				NextRetryAt: time.Now(),
			}
			if err := tx.Create(&msg).Error; err != nil {
				return err
			}
		}

		if req.OrderType == "GIFT" {
			for _, item := range req.Items {
				deductPayload := OutboxPayload{
					TenantID:      req.TenantID,
					UserID:        req.SourceUserID,
					ItemCode:      item.ItemCode,
					Quantity:      item.Quantity,
					SourceOrderNo: order.OrderNo,
				}
				deductBytes, _ := json.Marshal(deductPayload)
				deductMsg := model.OutboxMessage{
					ID:          idgen.NextID(),
					TenantID:    req.TenantID,
					BizNo:       order.OrderNo,
					EventType:   "ASSET_DEDUCT",
					Payload:     string(deductBytes),
					Status:      "PENDING",
					MaxRetries:  3,
					NextRetryAt: time.Now(),
				}
				if err := tx.Create(&deductMsg).Error; err != nil {
					return err
				}
			}
		}
		noticePayload, _ := json.Marshal(map[string]string{
			"tenant_id": req.TenantID,
			"user_id":   req.UserID,
			"order_no":  order.OrderNo,
			"msg":       "您收到一笔资产发放。",
		})
		noticeMsg := model.OutboxMessage{
			ID:          idgen.NextID(),
			TenantID:    req.TenantID,
			BizNo:       order.OrderNo,
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

	if err != nil {
		metrics.OrderFailedTotal.Inc()
		return nil, err
	}

	metrics.OrderTotal.Inc()
	return &OrderResult{
		OrderNo:     order.OrderNo,
		TotalAmount: 0,
		Status:      "PAID",
	}, nil
}
