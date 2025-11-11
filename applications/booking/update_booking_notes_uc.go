package booking

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"supra/db"
	"supra/logger"

	"github.com/google/uuid"
)

// UpdateBookingNotesUC updates only the user notes if booking is not approved.
func UpdateBookingNotesUC(bookingID string, payload []byte) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[update-booking-notes-uc] ✏️ Starting notes update for booking %s", bookingID))

	// Step 1️⃣ Parse request payload
	type notesPayload struct {
		UserNotes string `json:"userNotes" validate:"required"`
	}
	var p notesPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-notes-uc] ❌ Failed to unmarshal payload for %s: %v", bookingID, err))
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	// Step 2️⃣ Validate booking ID
	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-booking-notes-uc] ⚠️ Invalid booking ID format: %s", bookingID))
		return nil, fmt.Errorf("invalid booking ID: %w", err)
	}

	// Step 3️⃣ Sanitize user notes
	note := strings.TrimSpace(p.UserNotes)
	if len(note) > 2000 {
		return nil, fmt.Errorf("user notes too long (max 2000 characters)")
	}

	// Step 4️⃣ Start DB transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-notes-uc] ❌ Failed to start transaction: %v", err))
		return nil, fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback()

	// Step 5️⃣ Update DB record safely (only user_notes)
	query := `
		UPDATE booking
		SET
			user_notes = $2
		WHERE
			booking_id = $1
			AND booking_status IS DISTINCT FROM 'APPROVED'
		RETURNING booking_id, booking_email, booking_status, payment_details_id,
		          seat_quantity, seat_id, total_amount, seat_type,
		          participant_ids, created_at, user_notes;
	`

	var (
		bk                Booking
		participantIDsRaw []byte
		idUUID            uuid.UUID
	)

	row := tx.QueryRow(query, id, note)
	if err := row.Scan(
		&idUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount, &bk.SeatType,
		&participantIDsRaw, &bk.CreatedAt, &bk.UserNotes,
	); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[update-booking-notes-uc] ⚠️ Booking %s not found or already approved", bookingID))
			return nil, fmt.Errorf("booking not found or already approved")
		}
		logger.Log.Error(fmt.Sprintf("[update-booking-notes-uc] Database error: %v", err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Step 6️⃣ Deserialize participant IDs
	if len(participantIDsRaw) > 0 {
		if err := json.Unmarshal(participantIDsRaw, &bk.ParticipantIDs); err != nil {
			logger.Log.Warn(fmt.Sprintf("[update-booking-notes-uc] Failed to unmarshal participant IDs: %v", err))
		}
	}
	bk.BookingID = idUUID

	// Step 7️⃣ Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-notes-uc] ❌ Commit failed: %v", err))
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[update-booking-notes-uc] ✅ Notes updated successfully for booking %s", bookingID))
	return &bk, nil
}
