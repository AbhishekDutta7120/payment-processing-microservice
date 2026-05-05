package handler

import (
	"net/http"
	"payment-service/internal/model"
	"payment-service/internal/service"
	"payment-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service service.PaymentService
	logger  logger.Logger
}

func NewPaymentHandler(service service.PaymentService, logger logger.Logger) *PaymentHandler {
	return &PaymentHandler{
		service: service,
		logger:  logger,
	}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req model.CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request payload: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	h.logger.Info("Received payment creation request for user: " + req.UserID)

	resp, err := h.service.CreatePayment(&req)
	if err != nil {
		h.logger.Error("Payment creation failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment processing failed", "details": err.Error()})
		return
	}

	statusCode := http.StatusCreated
	if resp.Status == model.StatusFailed {
		statusCode = http.StatusOK
	}

	c.JSON(statusCode, resp)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")

	h.logger.Info("Received payment query for ID: " + paymentID)

	resp, err := h.service.GetPayment(paymentID)
	if err != nil {
		h.logger.Error("Payment query failed: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}