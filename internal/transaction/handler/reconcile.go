package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"virtual-asset-reconcile-system/internal/transaction/model"
	"virtual-asset-reconcile-system/internal/transaction/service"
	"virtual-asset-reconcile-system/pkg/idgen"
	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ReconcileRun(c *gin.Context, db *gorm.DB) {
	var orders []model.Order
	err := db.Raw(`
		SELECT o.* FROM orders o
		LEFT JOIN outbox_messages m 
			ON o.order_no = m.biz_no 
			AND m.event_type = 'ASSET_GRANT' 
			AND m.status = 'SENT'
		WHERE o.status = 'PAID' AND m.id IS NULL		
	`).Scan(&orders).Error
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 1010, "reconcile query failed: "+err.Error())
		return
	}

	type ReconcileReport struct {
		TotalInconsistent int `json:"total_inconsistent"`
		Fixed             int `json:"fixed"`
	}

	report := ReconcileReport{
		TotalInconsistent: len(orders),
	}

	for _, order := range orders {
		var items []model.OrderItem
		if err := db.WithContext(c.Request.Context()).Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			continue
		}

		for _, item := range items {
			payload := service.OutboxPayload{
				TenantID:      order.TenantID,
				UserID:        order.UserID,
				ItemCode:      item.ItemCode,
				Quantity:      item.Quantity,
				SourceOrderNo: order.OrderNo,
			}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				continue
			}

			msg := model.OutboxMessage{
				ID:          idgen.NextID(),
				TenantID:    order.TenantID,
				BizNo:       order.OrderNo,
				EventType:   "ASSET_GRANT",
				Payload:     string(payloadBytes),
				Status:      "PENDING",
				MaxRetries:  3,
				NextRetryAt: time.Now(),
			}
			if err := db.Create(&msg).Error; err != nil {
				continue
			}
			report.Fixed++
		}
	}

	response.Success(c, report)
}
