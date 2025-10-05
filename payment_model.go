package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	StatusInitiated PaymentStatus = "INITIATED"
	StatusSuccess   PaymentStatus = "SUCCESS"
	StatusFailed    PaymentStatus = "FAILED"
)

type Payment struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          string        `gorm:"not null;index" json:"user_id"`
	Amount          int64         `gorm:"not null" json:"amount"`
	Currency        string        `gorm:"not null" json:"currency"`
	Status          PaymentStatus `gorm:"not null;default:'INITIATED'" json:"status"`
	IdempotencyKey  string        `gorm:"uniqueIndex;not null" json:"idempotency_key"`
	RetryCount      int           `gorm:"default:0" json:"retry_count"`
	FailureReason   string        `json:"failure_reason,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type CreatePaymentRequest struct {
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
	Amount         int64  `json:"amount" binding:"required,gt=0"`
	Currency       string `json:"currency" binding:"required,len=3"`
	UserID         string `json:"user_id" binding:"required"`
}

type PaymentResponse struct {
	PaymentID string        `json:"payment_id"`
	Status    PaymentStatus `json:"status"`
	Amount    int64         `json:"amount,omitempty"`
	Currency  string        `json:"currency,omitempty"`
	CreatedAt time.Time     `json:"created_at,omitempty"`
}