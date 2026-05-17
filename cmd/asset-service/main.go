package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"virtual-asset-reconcile-system/internal/asset/handler"
	"virtual-asset-reconcile-system/internal/asset/model"
	"virtual-asset-reconcile-system/internal/asset/service"
	"virtual-asset-reconcile-system/internal/db"
	"virtual-asset-reconcile-system/internal/middleware"
	"virtual-asset-reconcile-system/internal/outbox"
	"virtual-asset-reconcile-system/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	serviceName := "asset-service"
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
		api.GET("/users/:user_id/assets", func(c *gin.Context) {
			handler.GetUserAssets(c, database)
		})
		api.GET("/assets/grant/status", func(c *gin.Context) {
			handler.GetGrantStatus(c, database)
		})
		api.POST("/assets/grant", func(c *gin.Context) {
			handler.GrantAsset(c, database)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	// outbox handler registeration & start goroutine
	consumer := outbox.NewConsumer(database, 5*time.Second)
	consumer.RegisterHandler("ASSET_GRANT", service.OutboxAssetHandler)
	go consumer.Start(context.Background())

	logger.L.Info("starting server", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Asset{},
		&model.AssetLedger{},
	)
}
