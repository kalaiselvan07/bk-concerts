package booking

import (
	"context"
	"encoding/json"
	"fmt"

	"supra/db"     // exposes global *sql.DB or pgxpool.Pool
	"supra/logger" // structured logger

	"github.com/google/uuid"
)

// GetPendingBookingsByConcertID retrieves all bookings for a specific concert ID.
func GetPendingBookingsByConcertID(concertID string) ([]*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[get-all-booking-concertID-uc] Retrieving bookings for concert: %s", concertID))

	ctx := context.Background()

	selectAllSQL := `
			SELECT booking_id, booking_email, booking_status, payment_details_id,
			       receipt_image, seat_quantity, seat_id, concert_id, total_amount, seat_type,
			       participant_ids, created_at
			FROM booking
			WHERE concert_id = $1
			  AND booking_status IN ('VERIFYING', 'PENDING_VERIFICATION')
			ORDER BY created_at DESC`

	logger.Log.Info("[get-all-booking-concertID-uc] Executing SELECT all query (filtered).")

	// ✅ Run query with context (avoids stale connection issues)
	rows, err := db.DB.QueryContext(ctx, selectAllSQL, concertID)
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

		// Unmarshal participant IDs
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
