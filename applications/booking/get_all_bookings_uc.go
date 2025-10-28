package booking

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db" // Assumes global DB instance

	"github.com/google/uuid"
)

// GetAllBookings retrieves a slice of all booking records.
func GetAllBookings() ([]*Booking, error) {
	const selectAllSQL = `
		SELECT booking_id, booking_email, booking_status, payment_details_id, 
		       receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
		       participant_ids, created_at
		FROM booking
		ORDER BY created_at DESC`

	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	bookings := make([]*Booking, 0)

	for rows.Next() {
		bk := &Booking{}
		var participantIDsJSON []byte
		var bookingIDUUID uuid.UUID
		var receiptImage []byte

		err := rows.Scan(
			&bookingIDUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
			&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount, &bk.SeatType,
			&participantIDsJSON, &bk.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning booking row: %w", err)
		}

		bk.BookingID = bookingIDUUID
		bk.ReceiptImage = receiptImage

		if len(participantIDsJSON) > 0 && string(participantIDsJSON) != "null" {
			if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal participant IDs from database: %w", err)
			}
		}

		bookings = append(bookings, bk)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return bookings, nil
}
