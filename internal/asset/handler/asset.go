package handler

import (
	"net/http"

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

	response.Error(c, http.StatusNotImplemented, 2002, "GetUserAssets not implemented - 由你来实现")
}

func GetGrantStatus(c *gin.Context, db *gorm.DB) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, 2003, "order_no is required")
		return
	}

	response.Error(c, http.StatusNotImplemented, 2004, "GetGrantStatus not implemented - 由你来实现")
}

func GrantAsset(c *gin.Context, db *gorm.DB) {
	response.Error(c, http.StatusNotImplemented, 2005, "GrantAsset not implemented - 由你来实现")
}
