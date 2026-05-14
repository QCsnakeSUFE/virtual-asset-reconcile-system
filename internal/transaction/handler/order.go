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

type PurchaseRequest struct {
	TenantID      string         `json:"tenant_id" binding:"required"`
	UserID        string         `json:"user_id" binding:"required"`
	IdempotentKey string         `json:"idempotent_key" binding:"required"`
	Items         []PurchaseItem `json:"items" binding:"required,min=1"`
}

func Purchase(c *gin.Context, db *gorm.DB) {
	var purReq PurchaseRequest
	err := c.ShouldBindJSON(&purReq)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 1001, "invalid request:"+err.Error())
		return
	}

	// 调用 service
	svcReq := service.CreateOrderRequest{
		TenantID:      purReq.TenantID,
		UserID:        purReq.UserID,
		IdempotentKey: purReq.IdempotentKey,
		Items:         make([]service.OrderItemParam, len(purReq.Items)),
	}
	for i, item := range purReq.Items {
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
