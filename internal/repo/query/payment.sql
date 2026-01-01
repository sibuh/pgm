-- name: CreatePayment: one
INSERT INTO payments (amount, currency, reference, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
-- name: GetPaymentByID :one
SELECT id, amount, currency, reference, status, created_at, updated_at FROM payments WHERE id = $1;

-- name: GetPaymentByReference: one
SELECT id, amount, currency, reference, status, created_at, updated_at FROM payments WHERE reference = $1;
-- name: UpdatePaymentStatus: one
UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3