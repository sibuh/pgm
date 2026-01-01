package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pgm/internal/domain"
)

type paymentPostgresRepository struct {
	db *sql.DB
}

func NewPaymentPostgresRepository(db *sql.DB) domain.PaymentRepo {
	return &paymentPostgresRepository{db: db}
}

func (r *paymentPostgresRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (amount, currency, reference, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query, p.Amount, p.Currency, p.Reference, p.Status, p.CreatedAt, p.UpdatedAt).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (r *paymentPostgresRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `SELECT id, amount, currency, reference, status, created_at, updated_at FROM payments WHERE id = $1`
	return r.fetchOne(ctx, query, id)
}

func (r *paymentPostgresRepository) GetByReference(ctx context.Context, reference string) (*domain.Payment, error) {
	query := `SELECT id, amount, currency, reference, status, created_at, updated_at FROM payments WHERE reference = $1`
	return r.fetchOne(ctx, query, reference)
}

func (r *paymentPostgresRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	return nil
}

func (r *paymentPostgresRepository) GetByIDWithLock(ctx context.Context, id string) (*domain.Payment, error) {
	query := `SELECT id, amount, currency, reference, status, created_at, updated_at FROM payments WHERE id = $1 FOR UPDATE`
	return r.fetchOne(ctx, query, id)
}

func (r *paymentPostgresRepository) fetchOne(ctx context.Context, query string, args ...interface{}) (*domain.Payment, error) {
	p := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&p.ID, &p.Amount, &p.Currency, &p.Reference, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or return a custom error like domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}
	return p, nil
}
