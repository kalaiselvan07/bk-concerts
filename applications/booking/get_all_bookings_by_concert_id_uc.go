package booking

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db"     // Assuming this exposes a global *sql.DB
	"bk-concerts/logger" // Your structured logger

	"github.com/google/uuid"
)

// GetAllBookingsByConcertID retrieves all bookings for a specific concert ID.
// If status == "all", it fetches every booking; otherwise, it filters for verification-related statuses.
func GetAllBookingsByConcertID(concertID, status string) ([]*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[get-all-booking-concertID-uc] Retrieving bookings for concert: %s", concertID))

	// Define SQL query based on the filter condition
	var selectAllSQL string
	if status == "all" {
		selectAllSQL = `
			SELECT booking_id, booking_email, booking_status, payment_details_id,
			       receipt_image, seat_quantity, seat_id, concert_id, total_amount, seat_type,
			       participant_ids, created_at
			FROM booking
			WHERE concert_id = $1
			ORDER BY created_at DESC`
	} else {
		selectAllSQL = `
			SELECT booking_id, booking_email, booking_status, payment_details_id,
			       receipt_image, seat_quantity, seat_id, concert_id, total_amount, seat_type,
			       participant_ids, created_at
			FROM booking
			WHERE concert_id = $1
			  AND booking_status IN ('VERIFYING', 'PENDING_VERIFICATION')
			ORDER BY created_at DESC`
	}

	logger.Log.Info("[get-all-booking-concertID-uc] Executing SELECT all query (filtered).")

	// Run the query
	rows, err := db.DB.Query(selectAllSQL, concertID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-booking-concertID-uc] Database query failed for %s: %v", concertID, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var bookings []*Booking
	recordCount := 0

	for rows.Next() {
		bk := &Booking{}
		var participantIDsJSON []byte
		var bookingIDUUID uuid.UUID
		var receiptImage []byte

		err := rows.Scan(
			&bookingIDUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
			&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.ConcertID,
			&bk.TotalAmount, &bk.SeatType, &participantIDsJSON, &bk.CreatedAt,
		)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[get-all-booking-concertID-uc] Error scanning booking row for %s: %v", concertID, err))
			return nil, fmt.Errorf("error scanning booking row: %w", err)
		}

		bk.BookingID = bookingIDUUID
		bk.ReceiptImage = receiptImage

		// Unmarshal JSON array for participant IDs
		if len(participantIDsJSON) > 0 && string(participantIDsJSON) != "null" {
			if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
				logger.Log.Warn(fmt.Sprintf("[get-all-booking-concertID-uc] Failed to unmarshal participant IDs for booking %s: %v", bk.BookingID, err))
			}
		}

		bookings = append(bookings, bk)
		recordCount++
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-booking-concertID-uc] Error during row iteration for %s: %v", concertID, err))
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-booking-concertID-uc] Successfully retrieved %d booking records for %s.", recordCount, concertID))
	return bookings, nil
}
