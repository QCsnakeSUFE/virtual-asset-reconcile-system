package main

import (
	"log"
	"net/http"
	"os"

	"virtual-asset-reconcile-system/internal/db"
	"virtual-asset-reconcile-system/internal/middleware"
	"virtual-asset-reconcile-system/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	serviceName := "asset-service"
	logger.Init(serviceName)
	defer logger.L.Sync()

	database, err := db.InitDB()
	if err != nil {
		logger.L.Fatal("failed to init db", zap.Error(err))
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	logger.L.Info("starting server", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}
