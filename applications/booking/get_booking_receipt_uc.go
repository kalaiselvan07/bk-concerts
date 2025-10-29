package booking

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"bk-concerts/db"
	"bk-concerts/logger"

	"github.com/google/uuid"
)

// GetBookingReceiptUC retrieves the booking details including the latest receipt image.
func GetBookingReceiptUC(bookingID string) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[get-booking-receipt-uc] Fetching booking receipt for BookingID: %s", bookingID))

	// Step 1: Validate booking ID
	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-booking-receipt-uc] Invalid booking ID format: %s", bookingID))
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	// Step 2: Query booking data
	query := `
		SELECT 
			booking_id, booking_email, booking_status, payment_details_id,
			receipt_image, seat_quantity, seat_id, seat_type, total_amount,
			participant_ids, created_at
		FROM booking
		WHERE booking_id = $1
	`
	row := db.DB.QueryRowContext(context.Background(), query, id)

	bk := &Booking{}
	var (
		receiptBytes      []byte
		participantIDsRaw []byte
		idUUID            uuid.UUID
	)

	if err := row.Scan(
		&idUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptBytes, &bk.SeatQuantity, &bk.SeatID, &bk.SeatType,
		&bk.TotalAmount, &participantIDsRaw, &bk.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[get-booking-receipt-uc] Booking %s not found.", bookingID))
			return nil, fmt.Errorf("booking with ID %s not found", bookingID)
		}
		logger.Log.Error(fmt.Sprintf("[get-booking-receipt-uc] Database error for %s: %v", bookingID, err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Step 3: Unmarshal participant IDs
	if err := json.Unmarshal(participantIDsRaw, &bk.ParticipantIDs); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-booking-receipt-uc] Failed to unmarshal participant IDs for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to unmarshal participant IDs: %w", err)
	}

	bk.BookingID = idUUID
	bk.ReceiptImage = receiptBytes

	logger.Log.Info(fmt.Sprintf("[get-booking-receipt-uc] Booking receipt successfully retrieved for %s", bookingID))
	return bk, nil
}
