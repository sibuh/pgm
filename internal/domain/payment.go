package domain

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

type PaymentStatus string

const (
	StatusPending PaymentStatus = "PENDING"
	StatusSuccess PaymentStatus = "SUCCESS"
	StatusFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID        string        `json:"id"`
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
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	GetByReference(ctx context.Context, reference string) (*Payment, error)
	UpdateStatus(ctx context.Context, id string, status PaymentStatus) error
	// For row-level locking and idempotency
	GetByIDWithLock(ctx context.Context, id string) (*Payment, error)
}

type PaymentService interface {
	Create(ctx context.Context, payment *Payment) (*Payment, error)
	GetByID(ctx context.Context, id string) (*Payment, error)
	Process(ctx context.Context, id string) error
}

type MessagePublisher interface {
	PublishPaymentCreated(ctx context.Context, paymentID string) error
}
