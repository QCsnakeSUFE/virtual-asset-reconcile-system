package handler

import (
	"net/http"

	"virtual-asset-reconcile-system/internal/asset/service"
	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserAssets(c *gin.Context, db *gorm.DB) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, http.StatusBadRequest, 2001, "user_id is required")
		return
	}
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)

	items, err := service.GetUserAssets(c.Request.Context(), db, tenantIDStr, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 2002, err.Error())
		return
	}
	response.Success(c, items)
}

func GetGrantStatus(c *gin.Context, db *gorm.DB) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, 2003, "order_no is required")
		return
	}
	tenantID, _ := c.Get("tenant_id")
	tenantIDStr, _ := tenantID.(string)
	result, err := service.GetGrantStatus(c.Request.Context(), db, tenantIDStr, orderNo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 2004, err.Error())
	}
	response.Success(c, result)
}

func GrantAsset(c *gin.Context, db *gorm.DB) {
	var req struct {
		TenantID      string `json:"tenant_id"`
		UserID        string `json:"user_id" binding:"required"`
		ItemCode      string `json:"item_code" binding:"required"`
		Quantity      int64  `json:"quantity" binding:"required,min=1"`
		SourceOrderNo string `json:"source_order_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 2005, "invalid request: "+err.Error())
		return
	}
	if req.TenantID == "" {
		tenantID, _ := c.Get("tenant_id")
		req.TenantID, _ = tenantID.(string)
	}

	traceID, _ := c.Get("trace_id")
	traceIDStr, _ := traceID.(string)

	svcReq := service.GrantAssetRequest{
		TenantID:      req.TenantID,
		UserID:        req.UserID,
		ItemCode:      req.ItemCode,
		Quantity:      req.Quantity,
		SourceOrderNo: req.SourceOrderNo,
		TraceID:       traceIDStr,
	}
	result, err := service.GrantAsset(c.Request.Context(), db, svcReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 2006, err.Error())
		return
	}
	response.Success(c, result)
}
