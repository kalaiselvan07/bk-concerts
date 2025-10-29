package booking

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db"     // Assumes global DB instance
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: Booking struct definition is assumed here

// GetAllBookings retrieves all bookings associated with a specific user email.
func GetAllBookings(userEmail string) ([]*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[get-all-booking-uc] Retrieving bookings for user: %s", userEmail))

	// 1. SQL query filters by booking_email
	const selectAllSQL = `
		SELECT booking_id, booking_email, booking_status, payment_details_id, 
		       receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
		       participant_ids, created_at
		FROM booking
		WHERE booking_email = $1  -- Filter added
		ORDER BY created_at DESC`

	// 2. Execute the query
	logger.Log.Info("[get-all-booking-uc] Executing SELECT all query (filtered).")
	// Pass userEmail as the argument
	rows, err := db.DB.Query(selectAllSQL, userEmail)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-booking-uc] Database query failed for %s: %v", userEmail, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	bookings := make([]*Booking, 0)
	recordCount := 0

	// 3. Iterate
	for rows.Next() {
		bk := &Booking{}
		var participantIDsJSON []byte
		var bookingIDUUID uuid.UUID
		var receiptImage []byte // Use sql.NullBytes for potentially null images

		err := rows.Scan(
			&bookingIDUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
			&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount, &bk.SeatType,
			&participantIDsJSON, &bk.CreatedAt,
		)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[get-all-booking-uc] Error scanning booking row for %s: %v", userEmail, err))
			return nil, fmt.Errorf("error scanning booking row: %w", err)
		}

		bk.BookingID = bookingIDUUID
		bk.ReceiptImage = receiptImage

		// 4. JSON unmarshal for Participant IDs
		if len(participantIDsJSON) > 0 && string(participantIDsJSON) != "null" {
			if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
				logger.Log.Error(fmt.Sprintf("[get-all-booking-uc] Failed to unmarshal participant IDs for booking %s: %v", bk.BookingID, err))
				// We can choose to just log this error or fail the whole query
				// return nil, fmt.Errorf("failed to unmarshal participant IDs: %w", err)
			}
		}

		bookings = append(bookings, bk)
		recordCount++
	}

	// 5. Check for errors during iteration
	if err = rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-booking-uc] Error during row iteration for %s: %v", userEmail, err))
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-booking-uc] Successfully retrieved %d booking records for %s.", recordCount, userEmail))
	return bookings, nil
}
