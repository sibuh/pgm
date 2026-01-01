package http

import (
	"net/http"

	"pgm/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	svc domain.PaymentService
}

func NewPaymentHandler(e *echo.Echo, uc domain.PaymentService) {
	handler := &PaymentHandler{
		svc: uc,
	}
	e.Group("v1")
	e.POST("/payments", handler.Create)
	e.GET("/payments/:id", handler.GetByID)
}

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

func (h *PaymentHandler) Create(c echo.Context) error {
	var p domain.Payment
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := c.Validate(&p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	res, err := h.svc.Create(c.Request().Context(), &p)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, res)
}

func (h *PaymentHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	res, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if res == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "payment not found"})
	}

	return c.JSON(http.StatusOK, res)
}
