package booking

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"

	"github.com/google/uuid"
)

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
	var p UpdateBookingParams
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Start a transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Validate ID
	id, err := uuid.Parse(bookingID)
	if err != nil {
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
		// No fields to update, fetch the current details and return them
		return GetBookingTx(tx, bookingID) // Using a transactional Get to be safe
	}

	// 3. Construct the final SQL
	updateSQL := fmt.Sprintf(`
		UPDATE booking
		SET %s
		WHERE booking_id = $1
		RETURNING booking_id, booking_email, booking_status, payment_details_id, 
				  receipt_image, seat_quantity, seat_id, total_amount, 
				  participant_ids, created_at, seat_type`, // Assuming seat_type is now in the Booking struct
		strings.Join(sets, ", "))

	// 4. Execute and scan the returned row
	bk := &Booking{}
	var receiptImage []byte
	var participantIDsJSON []byte

	// Use tx.QueryRow
	row := tx.QueryRow(updateSQL, args...)

	if err := row.Scan(
		&bk.BookingID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptImage, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount,
		&participantIDsJSON, &bk.CreatedAt, &bk.SeatType,
	); err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, fmt.Errorf("booking with ID %s not found", bookingID)
		}
		tx.Rollback()
		return nil, fmt.Errorf("database update error: %w", err)
	}

	// Convert back from DB formats
	bk.ReceiptImage = receiptImage
	if err := json.Unmarshal(participantIDsJSON, &bk.ParticipantIDs); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to unmarshal participant IDs: %w", err)
	}

	// 5. Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return bk, nil
}
