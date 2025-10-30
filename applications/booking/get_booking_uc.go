package booking

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bk-concerts/db"     // Assumes global DB instance
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: Booking struct definition is assumed here

// GetBooking retrieves a single booking record for general API reading (non-transactional).
func GetBooking(bookingID string) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[get-booking-uc] Starting standard read for BookingID: %s", bookingID))

	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-booking-uc] Read failed for %s: Invalid UUID format.", bookingID))
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	const selectSQL = `
		SELECT booking_id, booking_email, booking_status, payment_details_id, 
		       receipt_image, seat_quantity, seat_id, concert_id, total_amount, seat_type, 
		       participant_ids, created_at
		FROM booking
		WHERE booking_id = $1`

	logger.Log.Info(fmt.Sprintf("[get-booking-uc] Executing standard DB query for ID: %s", bookingID))
	row := db.DB.QueryRow(selectSQL, id)

	bk := &Booking{}
	var receiptImage []byte
	var participantIDsJSON []byte
	var bookingIDUUID uuid.UUID

	// Use db.DB.QueryRow() for non-transactional read
	err = row.Scan(
		&bookingIDUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.ConcertID, &bk.TotalAmount, &bk.SeatType,
		&participantIDsJSON, &bk.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[get-booking-uc] Read failed for %s: Booking not found.", bookingID))
			return nil, fmt.Errorf("booking with ID %s not found", bookingID)
		}
		logger.Log.Error(fmt.Sprintf("[get-booking-uc] Database query error for %s: %v", bookingID, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}

	bk.BookingID = bookingIDUUID
	bk.ReceiptImage = receiptImage

	if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-booking-uc] Failed to unmarshal participant IDs for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to unmarshal participant IDs from database: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-booking-uc] Booking %s retrieved successfully. Status: %s.", bookingID, bk.BookingStatus))
	return bk, nil
}

// GetBookingTx retrieves a single booking record using an existing transaction.
func GetBookingTx(tx *sql.Tx, bookingID string) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[get-booking-uc] Starting transactional read for BookingID: %s", bookingID))

	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-booking-uc] Transactional read failed for %s: Invalid UUID format.", bookingID))
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	const selectSQL = `
		SELECT booking_id, booking_email, booking_status, payment_details_id, 
		       receipt_image, seat_quantity, seat_id, concert_id, total_amount, seat_type, 
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
		&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.ConcertID, &bk.TotalAmount, &bk.SeatType,
		&participantIDsJSON, &bk.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[get-booking-uc] Transactional read failed for %s: Booking not found.", bookingID))
			return nil, fmt.Errorf("booking with ID %s not found", bookingID)
		}
		logger.Log.Error(fmt.Sprintf("[get-booking-uc] Transactional query error for %s: %v", bookingID, err))
		return nil, fmt.Errorf("transactional query error: %w", err)
	}

	bk.BookingID = bookingIDUUID
	bk.ReceiptImage = receiptImage

	if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-booking-uc] Transactional unmarshal failed for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to unmarshal participant IDs from database: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-booking-uc] Booking %s retrieved successfully within transaction.", bookingID))
	return bk, nil
}
