package http

import (
	"net/http"

	"pgm/internal/domain"

	"github.com/labstack/echo/v4"
)

// PaymentHandler handles HTTP requests for payments
type paymentHandler struct {
	svc domain.PaymentService
}

// NewPaymentHandler initializes the payment routes
func NewPaymentHandler(g *echo.Group, uc domain.PaymentService) domain.PaymentHandler {
	handler := &paymentHandler{
		svc: uc,
	}
	g.POST("/payments", handler.CreatePayment)
	g.GET("/payments/:id", handler.GetPaymentByID)
	return handler
}

// CreatePayment handles the creation of a new payment
// @Summary Create a new payment
// @Description Creates a new payment with the provided details
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body domain.PaymentRequest true "Payment details"
// @Success 201 {object} domain.Payment "Payment created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body or validation failed"
// @Failure 409 {object} ErrorResponse "Payment with this reference already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/payments [post]
func (h *paymentHandler) CreatePayment(c echo.Context) error {
	var pr domain.PaymentRequest
	if err := c.Bind(&pr); err != nil {
		return domain.NewError(
			http.StatusBadRequest,
			"invalid request body",
			"failed to bind request body",
			err,
			nil,
		)
	}

	res, err := h.svc.CreatePayment(c.Request().Context(), &pr)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

// GetPaymentByID retrieves a payment by its ID
// @Summary Get payment by ID
// @Description Retrieves payment details by payment ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} domain.Payment "Payment found"
// @Failure 400 {object} ErrorResponse "Invalid payment ID format"
// @Failure 404 {object} ErrorResponse "Payment not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/payments/{id} [get]
func (h *paymentHandler) GetPaymentByID(c echo.Context) error {
	id := c.Param("id")
	res, err := h.svc.GetPaymentByID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
