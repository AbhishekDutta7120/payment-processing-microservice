package main

import (
	"log"
	"net/http"
	"os"

	"payment-service/internal/handler"
	"payment-service/internal/repository"
	"payment-service/internal/service"
	"payment-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	appLogger := logger.NewLogger()

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=payments port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate tables
	if err := repository.AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize layers
	repo := repository.NewPaymentRepository(db)
	svc := service.NewPaymentService(repo, appLogger)
	h := handler.NewPaymentHandler(svc, appLogger)

	// Setup router
	r := gin.Default()
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Payment routes
	r.POST("/payments", h.CreatePayment)
	r.GET("/payments/:payment_id", h.GetPayment)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appLogger.Info("Starting payment service on port " + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}