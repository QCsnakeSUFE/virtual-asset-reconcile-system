package main

import (
	"log"
	"net/http"
	"os"

	"virtual-asset-reconcile-system/internal/db"
	"virtual-asset-reconcile-system/internal/middleware"
	"virtual-asset-reconcile-system/internal/transaction/handler"
	"virtual-asset-reconcile-system/internal/transaction/model"
	"virtual-asset-reconcile-system/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	serviceName := "transaction-service"
	logger.Init(serviceName)
	defer logger.L.Sync()

	database, err := db.InitDB()
	if err != nil {
		logger.L.Fatal("failed to init db", zap.Error(err))
	}

	if err := autoMigrate(database); err != nil {
		logger.L.Fatal("failed to migrate db", zap.Error(err))
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Trace())

	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := database.DB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "db check failed"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "db ping failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": serviceName})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api/v1")
	{
		api.POST("/orders/purchase", func(c *gin.Context) {
			handler.Purchase(c, database)
		})
		api.POST("/orders/grant", func(c *gin.Context) {
			handler.Grant(c, database)
		})
		api.POST("/orders/gift", func(c *gin.Context) {
			handler.Gift(c, database)
		})
		api.POST("/orders/payment/callback", func(c *gin.Context) {
			handler.PaymentCallback(c, database)
		})
		api.GET("/orders/:order_no", func(c *gin.Context) {
			handler.GetOrder(c, database)
		})
		api.POST("/reconcile/run", func(c *gin.Context) {
			handler.ReconcileRun(c, database)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.L.Info("starting server", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Order{},
		&model.OrderItem{},
		&model.OutboxMessage{},
	)
}
