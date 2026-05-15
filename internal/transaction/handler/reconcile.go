package handler

import (
	"net/http"

	"virtual-asset-reconcile-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ReconcileRun(c *gin.Context, db *gorm.DB) {
	response.Error(c, http.StatusNotImplemented, 1008, "reconcile worker not implemented yet - 由你来实现")
}
