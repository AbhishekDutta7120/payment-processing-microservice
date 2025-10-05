package service

import (
	"errors"
	"fmt"
	"math/rand"
	"payment-service/internal/model"
	"payment-service/internal/repository"
	"payment-service/pkg/logger"
	"time"

	"github.com/google/uuid"
)

const (
	MaxRetries  = 3
	FailureRate = 0.3 // 30% chance of simulated failure
	MinAmount   = 1
	MaxAmount   = 1000000
)

type PaymentService interface {
	CreatePayment(req *model.CreatePaymentRequest) (*model.PaymentResponse, error)
	GetPayment(paymentID string) (*model.PaymentResponse, error)
}

type paymentService struct {
	repo   repository.PaymentRepository
	logger logger.Logger
}

func NewPaymentService(repo repository.PaymentRepository, logger logger.Logger) PaymentService {
	return &paymentService{
		repo:   repo,
		logger: logger,
	}
}

func (s *paymentService) CreatePayment(req *model.CreatePaymentRequest) (*model.PaymentResponse, error) {
	s.logger.Info(fmt.Sprintf("Processing payment request with idempotency key: %s", req.IdempotencyKey))

	// Validate request
	if err := s.validatePaymentRequest(req); err != nil {
		s.logger.Error(fmt.Sprintf("Validation failed: %v", err))
		return nil, err
	}

	// Check idempotency - if payment exists, return existing result
	existingPayment, err := s.repo.GetByIdempotencyKey(req.IdempotencyKey)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error checking idempotency: %v", err))
		return nil, err
	}

	if existingPayment != nil {
		s.logger.Info(fmt.Sprintf("Idempotent request detected. Returning existing payment: %s", existingPayment.ID))
		return s.toPaymentResponse(existingPayment), nil
	}

	// Create new payment
	payment := &model.Payment{
		UserID:         req.UserID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         model.StatusInitiated,
		IdempotencyKey: req.IdempotencyKey,
		RetryCount:     0,
	}

	if err := s.repo.Create(payment); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create payment: %v", err))
		return nil, err
	}

	// Process payment with retry logic
	if err := s.processPaymentWithRetry(payment); err != nil {
		s.logger.Error(fmt.Sprintf("Payment processing failed after retries: %v", err))
		payment.Status = model.StatusFailed
		payment.FailureReason = err.Error()
		s.repo.Update(payment)
		return s.toPaymentResponse(payment), nil
	}

	s.logger.Info(fmt.Sprintf("Payment processed successfully: %s", payment.ID))
	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) GetPayment(paymentID string) (*model.PaymentResponse, error) {
	s.logger.Info(fmt.Sprintf("Fetching payment: %s", paymentID))

	id, err := uuid.Parse(paymentID)
	if err != nil {
		return nil, errors.New("invalid payment ID format")
	}

	payment, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to fetch payment: %v", err))
		return nil, err
	}

	return s.toPaymentResponse(payment), nil
}

func (s *paymentService) processPaymentWithRetry(payment *model.Payment) error {
	var lastErr error

	for i := 0; i < MaxRetries; i++ {
		payment.RetryCount = i + 1
		s.logger.Info(fmt.Sprintf("Processing payment attempt %d/%d for payment: %s", i+1, MaxRetries, payment.ID))

		// Simulate payment processing
		if err := s.simulatePaymentProcessing(payment); err != nil {
			lastErr = err
			s.logger.Warn(fmt.Sprintf("Payment attempt %d failed: %v", i+1, err))

			// Exponential backoff
			if i < MaxRetries-1 {
				backoff := time.Duration(i+1) * 100 * time.Millisecond
				time.Sleep(backoff)
			}
			continue
		}

		// Success
		payment.Status = model.StatusSuccess
		s.repo.Update(payment)
		return nil
	}

	return fmt.Errorf("payment failed after %d retries: %v", MaxRetries, lastErr)
}

func (s *paymentService) simulatePaymentProcessing(payment *model.Payment) error {
	// Simulate random failure for testing
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() < FailureRate {
		return errors.New("payment gateway timeout")
	}

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	return nil
}

func (s *paymentService) validatePaymentRequest(req *model.CreatePaymentRequest) error {
	if req.Amount < MinAmount {
		return fmt.Errorf("amount must be at least %d", MinAmount)
	}

	if req.Amount > MaxAmount {
		return fmt.Errorf("amount cannot exceed %d", MaxAmount)
	}

	if len(req.Currency) != 3 {
		return errors.New("currency must be a 3-letter ISO code")
	}

	if req.IdempotencyKey == "" {
		return errors.New("idempotency key is required")
	}

	if req.UserID == "" {
		return errors.New("user ID is required")
	}

	return nil
}

func (s *paymentService) toPaymentResponse(payment *model.Payment) *model.PaymentResponse {
	return &model.PaymentResponse{
		PaymentID: payment.ID.String(),
		Status:    payment.Status,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		CreatedAt: payment.CreatedAt,
	}
}