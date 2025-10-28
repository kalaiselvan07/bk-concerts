package booking

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bk-concerts/applications/participant"
	"bk-concerts/applications/seat"
	"bk-concerts/db"

	"github.com/google/uuid"
)

// NOTE: Booking struct and constant definitions are assumed to be defined in this file/package.

type CreateBookingParams struct {
	BookingEmail     string                 `json:"bookingEmail" validate:"required"`
	PaymentDetailsID string                 `json:"paymentDetailsID" validate:"required"`
	ReceiptImage     string                 `json:"receiptImage" validate:"required"`
	SeatQuantity     int                    `json:"seatQuantity" validate:"required"`
	SeatID           string                 `json:"seatID" validate:"required"`
	TotalAmount      float64                `json:"totalAmount" validate:"required"`
	Participants     []*participantsDetails `json:"participants"`
}

type participantsDetails struct {
	Name  string `json:"name" validate:"required"`
	WaNum string `json:"wpNum" validate:"required"`
	Email string `json:"email,omitempty"`
}

var ErrNotEnoughSeats = errors.New("not enough seats available")

// BookNow orchestrates the entire booking process within a database transaction.
func BookNow(payload []byte) (*Booking, error) {
	var p CreateBookingParams
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("%s: unmarshal error: %w", CANCELLED, err)
	}

	// Start a transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to start transaction: %w", CANCELLED, err)
	}
	defer tx.Rollback() // Rollback is safe to call even after Commit

	// --- 1. Validate and Update Seat Availability ---
	updatedSeat, err := validateAndUpdateSeatTx(tx, p.SeatID, p.SeatQuantity)
	if err != nil {
		return nil, fmt.Errorf("%s: seat update failed: %w", CANCELLED, err)
	}

	fmt.Printf("updated seat: %d", updatedSeat.Available)

	// --- 2. Save Participants and collect IDs ---
	participantIDs, err := addParticipantsTx(tx, p.Participants)
	if err != nil {
		return nil, fmt.Errorf("%s: adding participants failed: %w", CANCELLED, err)
	}

	// --- 3. Create and Save Booking Record ---
	// ⬅️ CRITICAL FIX: Pass the SeatType from the updatedSeat object
	bk, err := newBookingTx(tx, &p, updatedSeat.SeatType, participantIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: booking insertion failed: %w", CANCELLED, err)
	}

	// --- 4. Commit the transaction ---
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: failed to commit transaction: %w", CANCELLED, err)
	}

	return bk, nil
}

// validateAndUpdateSeatTx deducts seats and checks availability within a transaction.
func validateAndUpdateSeatTx(tx *sql.Tx, seatID string, quantity int) (*seat.Seat, error) {
	// Logic remains unchanged, relies on seat.GetSeatForUpdateTx and seat.UpdateAvailableTx
	currentSeat, err := seat.GetSeatForUpdateTx(tx, seatID)
	if err != nil {
		return nil, err
	}

	newVal := currentSeat.Available - quantity
	if newVal < 0 {
		return nil, fmt.Errorf("%w: requested %d, available %d",
			ErrNotEnoughSeats, quantity, currentSeat.Available)
	}

	updatedSeat, err := seat.UpdateAvailableTx(tx, seatID, newVal)
	if err != nil {
		return nil, fmt.Errorf("failed to update seat count: %w", err)
	}

	return updatedSeat, nil
}

// addParticipantsTx creates all participant records within a transaction.
func addParticipantsTx(tx *sql.Tx, details []*participantsDetails) ([]string, error) {
	var pts []string
	for _, detail := range details {
		toCreate := participant.CreateParticipantParams{
			Name:  detail.Name,
			WaNum: detail.WaNum,
			Email: detail.Email,
		}
		mSt, err := json.Marshal(toCreate)
		if err != nil {
			return nil, err
		}

		newPt, err := participant.AddParticipantTx(tx, mSt)
		if err != nil {
			return nil, err
		}
		// Note: Assuming newPt.UserID is now a string due to the implementation of AddParticipantTx
		pts = append(pts, newPt.UserID)
	}
	return pts, nil
}

// newBookingTx creates and inserts the booking record within a transaction.
// ⬅️ CRITICAL FIX: Added seatType parameter to the signature
func newBookingTx(tx *sql.Tx, p *CreateBookingParams, seatType string, participantIDs []string) (*Booking, error) {
	bkID := uuid.New()

	participantIDsJSON, err := json.Marshal(participantIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal participant IDs: %w", err)
	}

	receiptBytes := []byte(p.ReceiptImage)

	bk := &Booking{
		BookingID:        bkID,
		BookingEmail:     p.BookingEmail,
		BookingStatus:    VERIFYING,
		PaymentDetailsID: p.PaymentDetailsID,
		ReceiptImage:     receiptBytes,
		SeatQuantity:     p.SeatQuantity,
		SeatID:           p.SeatID,
		TotalAmount:      p.TotalAmount,
		SeatType:         seatType, // ⬅️ Storing the passed SeatType
		ParticipantIDs:   participantIDs,
		CreatedAt:        time.Now(),
	}

	const insertSQL = `
		INSERT INTO booking (booking_id, booking_email, booking_status, payment_details_id, 
		                     receipt_image, seat_quantity, seat_id, total_amount, seat_type, 
		                     participant_ids, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(
		insertSQL,
		bk.BookingID,
		bk.BookingEmail,
		bk.BookingStatus,
		p.PaymentDetailsID,
		bk.ReceiptImage,
		bk.SeatQuantity,
		bk.SeatID,
		bk.TotalAmount,
		bk.SeatType,
		participantIDsJSON,
		bk.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("transactional insert error: %w", err)
	}

	return bk, nil
}
