package service

import (
	"context"
	"errors"
	"fmt"
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

type OrderItemParam struct {
	ItemCode  string
	ItemName  string
	Quantity  int
	UnitPrice int64
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
