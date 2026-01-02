package domain

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

type PaymentStatus string

const (
	StatusPending PaymentStatus = "PENDING"
	StatusSuccess PaymentStatus = "SUCCESS"
	StatusFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID        uuid.UUID     `json:"id"`
	Amount    float64       `json:"amount" validate:"required,gt=0"`
	Currency  string        `json:"currency" validate:"required,oneof=ETB USD"`
	Reference string        `json:"reference" validate:"required"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
type PaymentRequest struct {
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Currency  string  `json:"currency" validate:"required,oneof=ETB USD"`
	Reference string  `json:"reference" validate:"required"`
}

func (pr PaymentRequest) Validate() error {
	return validation.ValidateStruct(&pr,
		validation.Field(&pr.Amount, validation.Required.Error("payment amount is required"), validation.Min(0.01)),
		validation.Field(&pr.Currency, validation.Required.Error("currency is required"), validation.In("ETB", "USD")),
		validation.Field(&pr.Reference, validation.Required.Error("payment reference is required")))
}

type PaymentRepo interface {
	CreatePayment(ctx context.Context, payment *Payment) error
	GetPaymentByID(ctx context.Context, id string) (*Payment, error)
	GetPaymentByReference(ctx context.Context, reference string) (*Payment, error)
	UpdatePaymentStatus(ctx context.Context, id string, status PaymentStatus) error
	// For row-level locking and idempotency
	GetPaymentByIDWithLock(ctx context.Context, id string) (*Payment, error)
}

type PaymentService interface {
	CreatePayment(ctx context.Context, pr *PaymentRequest) (*Payment, error)
	GetPaymentByID(ctx context.Context, id string) (*Payment, error)
	ProcessPayment(ctx context.Context, id string) error
}

type MessagePublisher interface {
	PublishPaymentCreated(ctx context.Context, paymentID string) error
}
