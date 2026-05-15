package handler

import (
	"net/http"

	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetNotifyStatus(c *gin.Context, db *gorm.DB) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, 3001, "order_no is required")
		return
	}

	response.Error(c, http.StatusNotImplemented, 3002, "GetNotifyStatus not implemented - 由你来实现")
}

func SendNotify(c *gin.Context, db *gorm.DB) {
	response.Error(c, http.StatusNotImplemented, 3003, "SendNotify not implemented - 由你来实现")
}
