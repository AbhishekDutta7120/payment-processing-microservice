package repository

import (
	"errors"
	"payment-service/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *model.Payment) error
	GetByID(id uuid.UUID) (*model.Payment, error)
	GetByIdempotencyKey(key string) (*model.Payment, error)
	Update(payment *model.Payment) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.Payment{})
}

func (r *paymentRepository) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByID(id uuid.UUID) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("id = ?", id).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByIdempotencyKey(key string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("idempotency_key = ?", key).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *model.Payment) error {
	return r.db.Save(payment).Error
}