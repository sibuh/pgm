package http

import (
	"net/http"

	"pgm/internal/domain"

	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	svc domain.PaymentService
}

func NewPaymentHandler(g *echo.Group, uc domain.PaymentService) {
	handler := &PaymentHandler{
		svc: uc,
	}
	g.POST("/payments", handler.CreatePayment)
	g.GET("/payments/:id", handler.GetPaymentByID)
}

func (h *PaymentHandler) CreatePayment(c echo.Context) error {
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

	if err := pr.Validate(); err != nil {
		return domain.NewError(
			http.StatusBadRequest,
			"validation failed",
			"payment request validation failed",
			err,
			map[string]interface{}{"req": pr},
		)
	}

	res, err := h.svc.CreatePayment(c.Request().Context(), &pr)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h *PaymentHandler) GetPaymentByID(c echo.Context) error {
	id := c.Param("id")
	res, err := h.svc.GetPaymentByID(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
