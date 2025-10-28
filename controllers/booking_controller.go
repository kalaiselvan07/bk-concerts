package controllers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"bk-concerts/applications/booking" // Using the correct module path

	"github.com/labstack/echo/v4"
)

func BookNowController(c echo.Context) error {
	payload, err := io.ReadAll(c.Request().Body)

	if err != nil {
		log.Printf("Error reading payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload.",
		})
	}

	newBooking, err := booking.BookNow(payload)

	if err != nil {
		log.Printf("Booking failed: %v", err)

		// Check for specific business logic error (Not enough seats)
		if errors.Is(err, booking.ErrNotEnoughSeats) {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()}) // 409 Conflict
		}

		// Check for CANCELLED prefix from the use case error handling
		if strings.HasPrefix(err.Error(), booking.CANCELLED) {
			// Unmarshal, validation, or general use case failure
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// Everything else is an internal server error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal booking failure: " + err.Error()})
	}

	log.Println("Booking initiated successfully. Status:", newBooking.BookingStatus)
	return c.JSON(http.StatusCreated, newBooking)
}

// UpdateBookingController handles PUT requests to update booking details.
func UpdateBookingController(c echo.Context) error {
	bookingID := c.Param("bookingID")
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload."})
	}

	updatedBooking, err := booking.UpdateBooking(bookingID, payload)

	if err != nil {
		log.Printf("Error updating booking %s: %v", bookingID, err)

		// Check for Not Found or Invalid ID
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking not found."})
		}
		if strings.Contains(err.Error(), "invalid ID format") || strings.Contains(err.Error(), "invalid booking status") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// General error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Booking update failed: " + err.Error()})
	}

	log.Println("Booking updated successfully. Status:", updatedBooking.BookingStatus)
	return c.JSON(http.StatusOK, updatedBooking)
}

// GetBookingController handles GET /bookings/:bookingID (Read One)
func GetBookingController(c echo.Context) error {
	bookingID := c.Param("bookingID")

	bk, err := booking.GetBooking(bookingID) // Non-transactional read

	if err != nil {
		log.Printf("Error fetching booking %s: %v", bookingID, err)

		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking not found."})
		}
		if strings.Contains(err.Error(), "invalid ID format") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve booking: " + err.Error()})
	}

	return c.JSON(http.StatusOK, bk)
}

// GetAllBookingsController handles GET /bookings (Read All)
func GetAllBookingsController(c echo.Context) error {
	bookingsList, err := booking.GetAllBookings()

	if err != nil {
		log.Printf("Error fetching all bookings: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve booking data: " + err.Error()})
	}

	return c.JSON(http.StatusOK, bookingsList)
}

// DeleteBookingController handles DELETE /bookings/:bookingID (Cancel/Delete)
func DeleteBookingController(c echo.Context) error {
	bookingID := c.Param("bookingID")

	// Call the use case which handles the transactional status change and seat refund
	cancelledBooking, err := booking.DeleteBooking(bookingID)

	if err != nil {
		log.Printf("Error canceling booking %s: %v", bookingID, err)

		// Check for Not Found or Invalid ID
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking not found."})
		}
		// Check for business logic conflicts (e.g., already CANCELLED)
		if strings.Contains(err.Error(), "already cancelled") {
			// Use 409 Conflict as the state prevents the action
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}

		// General error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Cancellation failed: " + err.Error()})
	}

	log.Println("Booking successfully cancelled. ID:", cancelledBooking.BookingID)
	// Return the cancelled booking object with a 200 OK status
	return c.JSON(http.StatusOK, cancelledBooking)
}
