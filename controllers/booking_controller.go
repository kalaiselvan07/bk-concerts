package controllers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"supra/applications/booking"
	"supra/logger"

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

// GetAllBookingsController handles GET /api/v1/bookings
// It retrieves *only* the bookings for the logged-in user.
func GetAllBookingsController(c echo.Context) error {
	// 1. Get the user's email from the context (set by JWTAuthMiddleware)
	userEmail, ok := c.Get("userEmail").(string)
	if !ok || userEmail == "" {
		logger.Log.Error("[booking] Failed to get userEmail from context in GetAllBookingsController.")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or missing token claims"})
	}

	logger.Log.Info(fmt.Sprintf("[booking] Fetching booking history for user: %s", userEmail))

	// 2. Call the use case, passing the user's email for filtering
	bookingsList, err := booking.GetAllBookings(userEmail)

	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Error fetching booking history for %s: %v", userEmail, err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve booking data: " + err.Error(),
		})
	}

	logger.Log.Info(fmt.Sprintf("[booking] Successfully retrieved %d bookings for %s.", len(bookingsList), userEmail))

	// 3. Return the filtered list
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

// UpdateBookingReceiptController handles PATCH /bookings/:bookingID/receipt
// It allows the user to re-upload or replace their payment receipt image.
func UpdateBookingReceiptController(c echo.Context) error {
	bookingID := c.Param("bookingID")

	// Read request body
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Failed to read payload for booking %s: %v", bookingID, err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload.",
		})
	}

	// Call use case
	updatedBooking, err := booking.UpdateBookingReceiptUC(bookingID, payload)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Failed to update receipt for booking %s: %v", bookingID, err))

		switch {
		case strings.Contains(err.Error(), "invalid booking ID"):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID."})
		case strings.Contains(err.Error(), "invalid base64"):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receipt image format."})
		case strings.Contains(err.Error(), "not found"):
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking not found."})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Receipt update failed: " + err.Error()})
		}
	}

	logger.Log.Info(fmt.Sprintf("[booking] Receipt updated successfully for booking %s.", bookingID))
	return c.JSON(http.StatusOK, updatedBooking)
}

func GetBookingReceiptController(c echo.Context) error {
	bookingID := c.Param("bookingID")

	logger.Log.Info(fmt.Sprintf("[booking] Fetching receipt for bookingID: %s", bookingID))

	// Call use case
	bk, err := booking.GetBookingReceiptUC(bookingID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Failed to fetch booking receipt for %s: %v", bookingID, err))

		switch {
		case err.Error() == "invalid booking ID format":
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID."})
		case err.Error() == "booking not found":
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking not found."})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to fetch booking receipt: %v", err),
			})
		}
	}

	logger.Log.Info(fmt.Sprintf("[booking] Booking receipt fetched successfully for %s", bookingID))
	return c.JSON(http.StatusOK, bk)
}

// 1️⃣ Fetch all bookings (admin only)
func GetAllBookingsAdminController(c echo.Context) error {
	logger.Log.Info("[booking-controller] Fetching all bookings (Admin)")
	bookings, err := booking.GetAllBookingsAdminUC()
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking-controller] Failed to fetch all bookings: %v", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch bookings"})
	}

	return c.JSON(http.StatusOK, bookings)
}

func GetAllBookingsByConcertIDController(c echo.Context) error {
	// 1. Get the user's email from the context (set by JWTAuthMiddleware)
	concertID := c.Param("concertID")
	status := c.Param("status")

	logger.Log.Info(fmt.Sprintf("[booking] Fetching booking history for concertID: %s", concertID))

	// 2. Call the use case, passing the user's email for filtering
	bookingsList, err := booking.GetAllBookingsByConcertID(concertID, status)

	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Error fetching booking history for %s: %v", concertID, err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve booking data: " + err.Error(),
		})
	}

	logger.Log.Info(fmt.Sprintf("[booking] Successfully retrieved %d bookings for %s.", len(bookingsList), concertID))

	// 3. Return the filtered list
	return c.JSON(http.StatusOK, bookingsList)
}

// VerifyBookingController handles PATCH /admin/bookings/:bookingID/verify?action=approve|reject
func VerifyBookingController(c echo.Context) error {
	bookingID := c.Param("bookingID")
	action := strings.ToLower(c.QueryParam("action"))

	if bookingID == "" || action == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing booking ID or action"})
	}

	var payload struct {
		Reason string `json:"reason"`
	}
	_ = c.Bind(&payload) // optional; only relevant for rejection

	logger.Log.Info(fmt.Sprintf("[booking-controller] Verification requested: %s → %s", bookingID, action))

	switch action {
	case "approve":
		bk, err := booking.ApproveBookingUC(bookingID)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[booking-controller] Approval failed for %s: %v", bookingID, err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to approve booking"})
		}
		return c.JSON(http.StatusOK, bk)

	case "reject":
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = "Rejected by admin"
		}
		bk, err := booking.RejectBookingUC(bookingID, reason)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[booking-controller] Rejection failed for %s: %v", bookingID, err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to reject booking"})
		}
		return c.JSON(http.StatusOK, bk)

	default:
		logger.Log.Warn(fmt.Sprintf("[booking-controller] Invalid action %q for booking %s", action, bookingID))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid action; must be approve or reject"})
	}
}

func UpdateParticipantBookingReceiptController(c echo.Context) error {
	bookingID := c.Param("bookingID")

	// Read request body
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Failed to read payload for booking %s: %v", bookingID, err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload.",
		})
	}

	// Call use case
	updatedBooking, err := booking.UpdateBookingReceiptUC(bookingID, payload)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Failed to update receipt for booking %s: %v", bookingID, err))

		switch {
		case strings.Contains(err.Error(), "invalid booking ID"):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid booking ID."})
		case strings.Contains(err.Error(), "invalid base64"):
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receipt image format."})
		case strings.Contains(err.Error(), "not found"):
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Booking not found."})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Receipt update failed: " + err.Error()})
		}
	}

	logger.Log.Info(fmt.Sprintf("[booking] Receipt updated successfully for booking %s.", bookingID))
	return c.JSON(http.StatusOK, updatedBooking)
}

func GetAllParicipantsByBookingIDIDController(c echo.Context) error {
	// 1. Get the user's email from the context (set by JWTAuthMiddleware)
	bookingID := c.Param("bookingID")
	logger.Log.Info(fmt.Sprintf("[booking] Fetching booking history for bookingID: %s", bookingID))

	// 2. Call the use case, passing the user's email for filtering
	participants, err := booking.GetAllParticipantByBookingID(bookingID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[booking] Error fetching participants history for %s: %v", bookingID, err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve paticipants data: " + err.Error(),
		})
	}

	logger.Log.Info(fmt.Sprintf("[booking] Successfully retrieved %d participants for %s.", len(participants), bookingID))

	// 3. Return the filtered list
	return c.JSON(http.StatusOK, participants)
}
