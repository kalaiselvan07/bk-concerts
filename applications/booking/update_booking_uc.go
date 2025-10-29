package booking

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"     // Using the correct module path
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: Booking struct, GetBookingTx, CONFIRMED, and CANCELLED constants are assumed here.

// UpdateBookingParams defines fields that can be optionally updated for a booking record.
type UpdateBookingParams struct {
	BookingEmail     string `json:"bookingEmail,omitempty"`
	BookingStatus    string `json:"bookingStatus,omitempty"`
	PaymentDetailsID string `json:"paymentDetailsID,omitempty"`
	ReceiptImage     string `json:"receiptImage,omitempty"` // Base64 string
	// SeatQuantity, SeatID, TotalAmount, and ParticipantIDs are typically immutable or handled by separate UCs.
}

// UpdateBooking performs a general update of booking details within a transaction.
func UpdateBooking(bookingID string, payload []byte) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[update-booking-uc] Starting update process for BookingID: %s", bookingID))

	var p UpdateBookingParams
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-uc] Unmarshal failed for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Start a transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-uc] Failed to start transaction for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	logger.Log.Info(fmt.Sprintf("[update-booking-uc] Transaction started for %s.", bookingID))

	// 1. Validate ID
	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-booking-uc] Update failed for %s: Invalid UUID format.", bookingID))
		return nil, fmt.Errorf("invalid booking ID format: %w", err)
	}

	// 2. Build the dynamic SQL query
	sets := []string{}
	args := []interface{}{id} // Start with booking_id as the first argument ($1)
	argCounter := 2           // SQL placeholders start at $2

	if p.BookingEmail != "" {
		sets = append(sets, fmt.Sprintf("booking_email = $%d", argCounter))
		args = append(args, p.BookingEmail)
		argCounter++
	}
	if p.BookingStatus != "" {
		// Basic validation for status change
		if p.BookingStatus != CONFIRMED && p.BookingStatus != CANCELLED {
			logger.Log.Warn(fmt.Sprintf("[update-booking-uc] Update failed for %s: Invalid status attempted: %s", bookingID, p.BookingStatus))
			return nil, fmt.Errorf("invalid booking status: %s", p.BookingStatus)
		}
		sets = append(sets, fmt.Sprintf("booking_status = $%d", argCounter))
		args = append(args, p.BookingStatus)
		argCounter++
	}
	if p.PaymentDetailsID != "" {
		sets = append(sets, fmt.Sprintf("payment_details_id = $%d", argCounter))
		args = append(args, p.PaymentDetailsID)
		argCounter++
	}
	if p.ReceiptImage != "" {
		// Convert ReceiptImage string (Base64) to []byte
		receiptBytes := []byte(p.ReceiptImage)
		sets = append(sets, fmt.Sprintf("receipt_image = $%d", argCounter))
		args = append(args, receiptBytes)
		argCounter++
	}

	if len(sets) == 0 {
		logger.Log.Warn(fmt.Sprintf("[update-booking-uc] Update skipped for %s: No updatable fields provided in payload.", bookingID))
		// No fields to update, fetch the current details and return them
		return GetBookingTx(tx, bookingID) // Using a transactional Get to be safe
	}

	// 3. Construct the final SQL
	updateSQL := fmt.Sprintf(`
		UPDATE booking
		SET %s
		WHERE booking_id = $1
		RETURNING booking_id, booking_email, booking_status, payment_details_id, 
		           receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
		           participant_ids, created_at`,
		strings.Join(sets, ", "))

	logger.Log.Info(fmt.Sprintf("[update-booking-uc] Executing UPDATE for %s with %d fields modified.", bookingID, len(sets)))

	// 4. Execute and scan the returned row
	bk := &Booking{}
	var receiptImage []byte
	var participantIDsJSON []byte
	var bookingIDUUID uuid.UUID

	// Use tx.QueryRow
	row := tx.QueryRow(updateSQL, args...)

	if err := row.Scan(
		&bookingIDUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount,
		&participantIDsJSON, &bk.CreatedAt, &bk.SeatType,
	); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			logger.Log.Warn(fmt.Sprintf("[update-booking-uc] Update failed for %s: Booking not found (Rollback).", bookingID))
			return nil, fmt.Errorf("booking with ID %s not found", bookingID)
		}
		tx.Rollback()
		logger.Log.Error(fmt.Sprintf("[update-booking-uc] Database update error for %s (Rollback): %v", bookingID, err))
		return nil, fmt.Errorf("database update error: %w", err)
	}

	// Assign mapped fields
	bk.BookingID = bookingIDUUID
	bk.ReceiptImage = receiptImage

	// Convert back from DB formats
	if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
		tx.Rollback()
		logger.Log.Error(fmt.Sprintf("[update-booking-uc] Failed to unmarshal participant IDs for %s (Rollback): %v", bookingID, err))
		return nil, fmt.Errorf("failed to unmarshal participant IDs: %w", err)
	}

	// 5. Commit the transaction
	if err := tx.Commit(); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-uc] Failed to commit transaction for %s: %v", bookingID, err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[update-booking-uc] Booking %s updated successfully. New Status: %s.", bookingID, bk.BookingStatus))
	return bk, nil
}
