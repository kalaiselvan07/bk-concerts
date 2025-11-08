package controllers

import (
	"io"
	"net/http"
	"strings"

	"supra/applications/paymentdetails" // Using the correct module path

	"github.com/labstack/echo/v4"
)

// AddPaymentController handles POST /payments (Create)
func AddPaymentController(c echo.Context) error {
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload."})
	}
	newPayment, err := paymentdetails.AddPayment(payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to record payment: " + err.Error()})
	}
	return c.JSON(http.StatusCreated, newPayment)
}

// GetPaymentController handles GET /payments/:paymentID (Read One)
func GetPaymentController(c echo.Context) error {
	paymentID := c.Param("paymentID")
	p, err := paymentdetails.GetPayment(paymentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Payment not found."})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payment: " + err.Error()})
	}
	return c.JSON(http.StatusOK, p)
}

// GetAllPaymentsController handles GET /payments (Read All)
func GetAllPaymentsController(c echo.Context) error {
	paymentsList, err := paymentdetails.GetAllPayments()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve payments data: " + err.Error()})
	}
	return c.JSON(http.StatusOK, paymentsList)
}

// UpdatePaymentController handles PUT /payments/:paymentID (Update Details)
func UpdatePaymentController(c echo.Context) error {
	paymentID := c.Param("paymentID")
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload."})
	}
	p, err := paymentdetails.UpdatePayment(paymentID, payload)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Payment not found."})
		}
		if strings.Contains(err.Error(), "invalid ID format") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Payment update failed: " + err.Error()})
	}
	return c.JSON(http.StatusOK, p)
}

// DeletePaymentController handles DELETE /payments/:paymentID (Delete)
func DeletePaymentController(c echo.Context) error {
	paymentID := c.Param("paymentID")
	rowsAffected, err := paymentdetails.DeletePayment(paymentID)
	if err != nil {
		if rowsAffected == 0 || strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Payment not found."})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete payment: " + err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
