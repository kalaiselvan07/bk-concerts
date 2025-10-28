package booking

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bk-concerts/db" // Assumes global DB instance

	"github.com/google/uuid"
)

// GetBooking retrieves a single booking record for general API reading (non-transactional).
func GetBooking(bookingID string) (*Booking, error) {
	id, err := uuid.Parse(bookingID)
	if err != nil {
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	const selectSQL = `
		SELECT booking_id, booking_email, booking_status, payment_details_id, 
		       receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
		       participant_ids, created_at
		FROM booking
		WHERE booking_id = $1`

	bk := &Booking{}
	var receiptImage []byte
	var participantIDsJSON []byte
	var bookingIDUUID uuid.UUID

	// Use db.DB.QueryRow() for non-transactional read
	row := db.DB.QueryRow(selectSQL, id)

	err = row.Scan(
		&bookingIDUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount, &bk.SeatType,
		&participantIDsJSON, &bk.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking with ID %s not found", bookingID)
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	bk.BookingID = bookingIDUUID
	bk.ReceiptImage = receiptImage

	if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participant IDs from database: %w", err)
	}

	return bk, nil
}

// GetBookingTx retrieves a single booking record using an existing transaction.
// This uses the exact same select/scan logic but receives *sql.Tx as context.
func GetBookingTx(tx *sql.Tx, bookingID string) (*Booking, error) {
	id, err := uuid.Parse(bookingID)
	if err != nil {
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	const selectSQL = `
		SELECT booking_id, booking_email, booking_status, payment_details_id, 
		       receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
		       participant_ids, created_at
		FROM booking
		WHERE booking_id = $1`

	bk := &Booking{}
	var receiptImage []byte
	var participantIDsJSON []byte
	var bookingIDUUID uuid.UUID

	// Use tx.QueryRow()
	row := tx.QueryRow(selectSQL, id)

	err = row.Scan(
		&bookingIDUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount, &bk.SeatType,
		&participantIDsJSON, &bk.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking with ID %s not found", bookingID)
		}
		return nil, fmt.Errorf("transactional query error: %w", err)
	}

	bk.BookingID = bookingIDUUID
	bk.ReceiptImage = receiptImage

	if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participant IDs from database: %w", err)
	}

	return bk, nil
}
