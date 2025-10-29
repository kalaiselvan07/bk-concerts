package booking

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bk-concerts/applications/auth"
	"bk-concerts/db"
	"bk-concerts/logger"

	"github.com/google/uuid"
)

// RejectBookingUC rejects a pending booking, updates status in DB, and notifies user via email.
func RejectBookingUC(bookingID string, reason string) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[reject-booking-uc] Starting rejection for booking: %s", bookingID))

	// Step 1: Validate input
	if strings.TrimSpace(bookingID) == "" {
		return nil, fmt.Errorf("bookingID cannot be empty")
	}
	if strings.TrimSpace(reason) == "" {
		reason = "No reason provided"
	}

	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[reject-booking-uc] Invalid booking ID: %s", bookingID))
		return nil, fmt.Errorf("invalid booking ID: %w", err)
	}

	// Step 2: Start transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[reject-booking-uc] Failed to start transaction: %v", err))
		return nil, fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback()

	// Step 3: Fetch booking details
	var (
		bk                Booking
		receiptBytes      []byte
		participantIDsRaw []byte
	)
	querySelect := `
		SELECT booking_id, booking_email, booking_status, payment_details_id,
		       receipt_image, seat_quantity, seat_id, total_amount, seat_type,
		       participant_ids, created_at
		FROM booking
		WHERE booking_id = $1
	`
	err = tx.QueryRow(querySelect, id).Scan(
		&bk.BookingID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptBytes, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount,
		&bk.SeatType, &participantIDsRaw, &bk.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[reject-booking-uc] Booking not found: %s", bookingID))
			return nil, fmt.Errorf("booking not found")
		}
		logger.Log.Error(fmt.Sprintf("[reject-booking-uc] Failed to fetch booking %s: %v", bookingID, err))
		return nil, err
	}

	// Step 4: Deserialize participants if available
	if len(participantIDsRaw) > 0 {
		_ = json.Unmarshal(participantIDsRaw, &bk.ParticipantIDs)
	}
	bk.ReceiptImage = receiptBytes

	// Step 5: Update booking → REJECTED
	queryUpdate := `
		UPDATE booking
		SET booking_status = 'REJECTED',
		    updated_at = $2
		WHERE booking_id = $1
	`
	_, err = tx.Exec(queryUpdate, id, time.Now())
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[reject-booking-uc] Failed to update booking %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	// Step 6: Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Log.Error(fmt.Sprintf("[reject-booking-uc] Commit failed for %s: %v", bookingID, err))
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[reject-booking-uc] Booking %s successfully marked as REJECTED.", bookingID))

	// Step 7: Send rejection email
	var base64Receipt string
	if len(receiptBytes) > 0 {
		base64Receipt = base64.StdEncoding.EncodeToString(receiptBytes)
	}

	emailErr := auth.SendBookingVerificationMail(
		bk.BookingEmail,
		"REJECTED",
		bk.BookingID.String(),
		base64Receipt,
		reason, // Optional rejection reason
	)
	if emailErr != nil {
		logger.Log.Warn(fmt.Sprintf("[reject-booking-uc] Booking %s rejected, but email sending failed: %v", bookingID, emailErr))
	} else {
		logger.Log.Info(fmt.Sprintf("[reject-booking-uc] Rejection email sent to %s.", bk.BookingEmail))
	}

	return &bk, nil
}
