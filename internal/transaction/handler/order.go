package handler

// BindJSON + 调 service + 返回响应

import (
	"net/http"
	"virtual-asset-reconcile-system/internal/transaction/service"
	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PurchaseItem struct {
	ItemCode  string `json:"item_code" binding:"required"`
	ItemName  string `json:"item_name" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	UnitPrice int64  `json:"unit_price" binding:"required,min=1"`
}

type NoPaymentItem struct {
	ItemCode string `json:"item_code" binding:"required"`
	ItemName string `json:"item_name" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

type PurchaseRequest struct {
	TenantID      string         `json:"tenant_id" binding:"required"`
	UserID        string         `json:"user_id" binding:"required"`
	IdempotentKey string         `json:"idempotent_key" binding:"required"`
	Items         []PurchaseItem `json:"items" binding:"required,min=1"`
}

type GrantRequest struct {
	TenantID      string          `json:"tenant_id" binding:"required"`
	UserID        string          `json:"user_id" binding:"required"`
	IdempotentKey string          `json:"idempotent_key" binding:"required"`
	Items         []NoPaymentItem `json:"items" binding:"required,min=1"`
}

type GiftRequest struct {
	TenantID      string          `json:"tenant_id" binding:"required"`
	UserID        string          `json:"user_id" binding:"required"`
	SourceUserID  string          `json:"source_user_id" binding:"required"`
	IdempotentKey string          `json:"idempotent_key" binding:"required"`
	Items         []NoPaymentItem `json:"items" binding:"required,min=1"`
}

func Purchase(c *gin.Context, db *gorm.DB) {
	var req PurchaseRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 1001, "invalid request:"+err.Error())
		return
	}

	svcReq := service.CreateOrderRequest{
		TenantID:      req.TenantID,
		UserID:        req.UserID,
		IdempotentKey: req.IdempotentKey,
		Items:         make([]service.OrderItemParam, len(req.Items)),
	}
	for i, item := range req.Items {
		svcReq.Items[i] = service.OrderItemParam{
			ItemCode:  item.ItemCode,
			ItemName:  item.ItemName,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}
	result, err := service.CreateOrder(c, db, svcReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 1002, err.Error())
		return
	}
	response.Success(c, result)
}

func Grant(c *gin.Context, db *gorm.DB) {
	var req GrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 1003, "invalid request: "+err.Error())
		return
	}

	svcReq := service.CreateDirectOrderRequest{
		TenantID:      req.TenantID,
		UserID:        req.UserID,
		SourceUserID:  "",
		IdempotentKey: req.IdempotentKey,
		OrderType:     "GRANT",
		Items:         make([]service.NoPaymentOrderItemParam, len(req.Items)),
	}
	for i, item := range req.Items {
		svcReq.Items[i] = service.NoPaymentOrderItemParam{
			ItemCode: item.ItemCode,
			ItemName: item.ItemName,
			Quantity: item.Quantity,
		}
	}
	result, err := service.CreateDirectOrder(c, db, svcReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 1004, err.Error())
		return
	}
	response.Success(c, result)
}

func Gift(c *gin.Context, db *gorm.DB) {
	var req GiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 1005, "invalid request: "+err.Error())
		return
	}

	svcReq := service.CreateDirectOrderRequest{
		TenantID:      req.TenantID,
		UserID:        req.UserID,
		SourceUserID:  req.SourceUserID,
		IdempotentKey: req.IdempotentKey,
		OrderType:     "GIFT",
		Items:         make([]service.NoPaymentOrderItemParam, len(req.Items)),
	}
	for i, item := range req.Items {
		svcReq.Items[i] = service.NoPaymentOrderItemParam{
			ItemCode: item.ItemCode,
			ItemName: item.ItemName,
			Quantity: item.Quantity,
		}
	}
	result, err := service.CreateDirectOrder(c, db, svcReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 1006, err.Error())
		return
	}
	response.Success(c, result)
}
