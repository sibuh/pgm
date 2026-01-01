package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"pgm/internal/domain"
)

type PaymentService struct {
	paymentRepo domain.PaymentRepo
	publisher   domain.MessagePublisher
}

func NewPaymentService(repo domain.PaymentRepo, publisher domain.MessagePublisher) domain.PaymentService {
	return &PaymentService{
		paymentRepo: repo,
		publisher:   publisher,
	}
}

func (u *PaymentService) Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	// Check if reference already exists
	existing, err := u.paymentRepo.GetByReference(ctx, p.Reference)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("payment with this reference already exists")
	}

	p.Status = domain.StatusPending
	err = u.paymentRepo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	// Publish to RabbitMQ
	err = u.publisher.PublishPaymentCreated(ctx, p.ID)
	if err != nil {
		// In a real-world scenario, we might want to use an outbox pattern here
		// to ensure the message is eventually published.
		fmt.Printf("failed to publish message: %v\n", err)
	}

	return p, nil
}

func (u *PaymentService) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	return u.paymentRepo.GetByID(ctx, id)
}

func (u *PaymentService) Process(ctx context.Context, id string) error {
	// Use row-level locking to prevent race conditions
	p, err := u.paymentRepo.GetByIDWithLock(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return errors.New("payment not found")
	}

	// Idempotency check: only process if PENDING
	if p.Status != domain.StatusPending {
		fmt.Printf("payment %s already processed with status %s\n", id, p.Status)
		return nil
	}

	// Simulate processing
	time.Sleep(2 * time.Second)

	newStatus := domain.StatusSuccess
	if rand.Float32() < 0.3 { // 30% failure rate for simulation
		newStatus = domain.StatusFailed
	}

	err = u.paymentRepo.UpdateStatus(ctx, id, newStatus)
	if err != nil {
		return err
	}

	fmt.Printf("payment %s processed with status %s\n", id, newStatus)
	return nil
}
