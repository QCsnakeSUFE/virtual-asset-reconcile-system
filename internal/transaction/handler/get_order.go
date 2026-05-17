package handler

import (
	"net/http"

	"virtual-asset-reconcile-system/internal/transaction/model"
	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetOrder(c *gin.Context, db *gorm.DB) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.Error(c, http.StatusBadRequest, 1005, "order_no is required")
		return
	}

	var order model.Order
	if err := db.WithContext(c.Request.Context()).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, 1006, "order not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, 1007, err.Error())
		return
	}

	var items []model.OrderItem
	db.WithContext(c.Request.Context()).Where("order_id = ?", order.ID).Find(&items)

	response.Success(c, gin.H{
		"order": order,
		"items": items,
	})
}
