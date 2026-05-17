package handler

import (
	"net/http"

	"virtual-asset-reconcile-system/internal/notify/service"
	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetNotifyStatus(c *gin.Context, db *gorm.DB) {

	tenantID, _ := c.Get("tenant_id")
	tenantIDStr := tenantID.(string)

	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, 3001, "order_no is required")
		return
	}
	notification, err := service.GetNotifyStatus(c.Request.Context(), db, tenantIDStr, orderNo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 3002, err.Error())
		return
	}
	response.Success(c, notification)
}

func SendNotify(c *gin.Context, db *gorm.DB) {
	var req service.SendNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 3003, "invalid request: "+err.Error())
		return
	}

	result, err := service.SendNotify(c.Request.Context(), db, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 3004, err.Error())
		return
	}
	response.Success(c, result)
}
