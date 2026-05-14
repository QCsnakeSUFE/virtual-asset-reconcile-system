package handler

import (
	"net/http"
	"virtual-asset-reconcile-system/internal/transaction/service"
	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PaymentCallback(c *gin.Context, db *gorm.DB) {
	var req struct {
		OrderNo  string `json:"order_no" binding:"required"`
		TenantID string `json:"tenant_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 1003, "invalid request: "+err.Error())
		return
	}

	svcReq := service.PaymentCallbackRequest{
		OrderNo:  req.OrderNo,
		TenantID: req.TenantID,
	}

	if err := service.ProcessPaymentCallback(c.Request.Context(), db, svcReq); err != nil {
		response.Error(c, http.StatusInternalServerError, 1004, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "payment processed successfully"})
}
