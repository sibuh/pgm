package service_test

import (
	"context"
	"testing"

	"pgm/internal/domain"
	"pgm/internal/repo/db"
	"pgm/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockQuerier is a mock implementation of db.Querier
type mockQuerier struct {
	mock.Mock
}

func (m *mockQuerier) CheckExistence(ctx context.Context, reference string) (bool, error) {
	args := m.Called(ctx, reference)
	return args.Bool(0), args.Error(1)
}

func (m *mockQuerier) CreatePayment(ctx context.Context, arg db.CreatePaymentParams) (db.Payment, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Payment), args.Error(1)
}

func (m *mockQuerier) GetPaymentByID(ctx context.Context, id uuid.UUID) (db.Payment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Payment), args.Error(1)
}

func (m *mockQuerier) GetPaymentByIDWithLock(ctx context.Context, id uuid.UUID) (db.Payment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Payment), args.Error(1)
}

func (m *mockQuerier) GetPaymentByReference(ctx context.Context, reference string) (db.Payment, error) {
	args := m.Called(ctx, reference)
	return args.Get(0).(db.Payment), args.Error(1)
}

func (m *mockQuerier) UpdatePaymentStatus(ctx context.Context, arg db.UpdatePaymentStatusParams) (db.Payment, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Payment), args.Error(1)
}

// mockPublisher is a mock implementation of domain.MessagePublisher
type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) PublishPaymentCreated(ctx context.Context, paymentID string) error {
	args := m.Called(ctx, paymentID)
	return args.Error(0)
}

func TestCreatePayment(t *testing.T) {
	tests := []struct {
		name          string
		req           *domain.PaymentRequest
		setupMock     func(mq *mockQuerier, mp *mockPublisher)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful creation",
			req: &domain.PaymentRequest{
				Amount:    100.50,
				Currency:  "USD",
				Reference: "ref-123",
			},
			setupMock: func(mq *mockQuerier, mp *mockPublisher) {
				mq.On("CheckExistence", mock.Anything, "ref-123").Return(false, nil)
				mq.On("CreatePayment", mock.Anything, db.CreatePaymentParams{
					Amount:    decimal.NewFromFloat(100.50),
					Currency:  "USD",
					Reference: "ref-123",
				}).Return(db.Payment{
					ID:        uuid.New(),
					Amount:    decimal.NewFromFloat(100.50),
					Currency:  "USD",
					Reference: "ref-123",
					Status:    db.PaymentstatusPENDING,
				}, nil)
				mp.On("PublishPaymentCreated", mock.Anything, mock.Anything).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "validation failure",
			req: &domain.PaymentRequest{
				Amount: -1,
			},
			setupMock:     func(mq *mockQuerier, mp *mockPublisher) {},
			expectedError: "validation failed",
		},
		{
			name: "already exists",
			req: &domain.PaymentRequest{
				Amount:    100.50,
				Currency:  "USD",
				Reference: "ref-exists",
			},
			setupMock: func(mq *mockQuerier, mp *mockPublisher) {
				mq.On("CheckExistence", mock.Anything, "ref-exists").Return(true, nil)
			},
			expectedError: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mq := new(mockQuerier)
			mp := new(mockPublisher)
			tt.setupMock(mq, mp)

			svc := service.NewPaymentService(mq, mp)
			res, err := svc.CreatePayment(context.Background(), tt.req)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.req.Amount, res.Amount)
				assert.Equal(t, tt.req.Reference, res.Reference)
			}
			mq.AssertExpectations(t)
			mp.AssertExpectations(t)
		})
	}
}

func TestGetPaymentByID(t *testing.T) {
	paymentID := uuid.New()
	tests := []struct {
		name          string
		id            string
		setupMock     func(mq *mockQuerier)
		expectedError string
	}{
		{
			name: "successful retrieval",
			id:   paymentID.String(),
			setupMock: func(mq *mockQuerier) {
				mq.On("GetPaymentByID", mock.Anything, paymentID).Return(db.Payment{
					ID:        paymentID,
					Amount:    decimal.NewFromFloat(100),
					Currency:  "USD",
					Reference: "ref-1",
					Status:    db.PaymentstatusSUCCESS,
				}, nil)
			},
		},
		{
			name:          "invalid uuid",
			id:            "not-a-uuid",
			setupMock:     func(mq *mockQuerier) {},
			expectedError: "Invalid payment ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mq := new(mockQuerier)
			mp := new(mockPublisher)
			tt.setupMock(mq)

			svc := service.NewPaymentService(mq, mp)
			res, err := svc.GetPaymentByID(context.Background(), tt.id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, paymentID, res.ID)
			}
			mq.AssertExpectations(t)
		})
	}
}

func TestProcessPayment(t *testing.T) {
	paymentID := uuid.New()
	tests := []struct {
		name          string
		id            string
		setupMock     func(mq *mockQuerier)
		expectedError string
	}{
		{
			name: "successful processing",
			id:   paymentID.String(),
			setupMock: func(mq *mockQuerier) {
				mq.On("GetPaymentByIDWithLock", mock.Anything, paymentID).Return(db.Payment{
					ID:     paymentID,
					Status: db.PaymentstatusPENDING,
				}, nil)
				mq.On("UpdatePaymentStatus", mock.Anything, mock.MatchedBy(func(p db.UpdatePaymentStatusParams) bool {
					return p.ID == paymentID && (p.Status == db.PaymentstatusSUCCESS || p.Status == db.PaymentstatusFAILED)
				})).Return(db.Payment{}, nil)
			},
		},
		{
			name: "already processed",
			id:   paymentID.String(),
			setupMock: func(mq *mockQuerier) {
				mq.On("GetPaymentByIDWithLock", mock.Anything, paymentID).Return(db.Payment{
					ID:     paymentID,
					Status: db.PaymentstatusSUCCESS,
				}, nil)
			},
			expectedError: "already processed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mq := new(mockQuerier)
			mp := new(mockPublisher)
			tt.setupMock(mq)

			svc := service.NewPaymentService(mq, mp)
			
			// Note: ProcessPayment has a time.Sleep(2 * time.Second) which makes tests slow.
			// In a real scenario, we might want to inject a clock or use a shorter sleep for tests.
			err := svc.ProcessPayment(context.Background(), tt.id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			mq.AssertExpectations(t)
		})
	}
}
