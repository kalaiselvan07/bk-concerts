package booking

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"bk-concerts/applications/auth"
	"bk-concerts/applications/participant"
	"bk-concerts/applications/seat"
	"bk-concerts/db"
	"bk-concerts/logger"

	"github.com/google/uuid"
)

type CreateBookingParams struct {
	BookingEmail     string                 `json:"bookingEmail" validate:"required"`
	PaymentDetailsID string                 `json:"paymentDetailsID" validate:"required"`
	ReceiptImage     string                 `json:"receiptImage" validate:"required"` // base64 from client
	SeatQuantity     int                    `json:"seatQuantity" validate:"required"`
	ConcertID        string                 `json:"concertID" validate:"required"`
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

func BookNow(payload []byte) (*Booking, error) {
	logger.Log.Info("[create-booking-uc] 🟢 Starting booking process")

	var p CreateBookingParams
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[create-booking-uc] ❌ JSON unmarshal failed: %v", err))
		return nil, fmt.Errorf("%s: unmarshal error: %w", CANCELLED, err)
	}

	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[create-booking-uc] ❌ Failed to start DB transaction: %v", err))
		return nil, fmt.Errorf("%s: failed to start transaction: %w", CANCELLED, err)
	}
	defer tx.Rollback()

	updatedSeat, err := validateAndUpdateSeatTx(tx, p.SeatID, p.SeatQuantity)
	if err != nil {
		return nil, fmt.Errorf("%s: seat update failed: %w", CANCELLED, err)
	}

	participantIDs, err := addParticipantsTx(tx, p.Participants)
	if err != nil {
		return nil, fmt.Errorf("%s: adding participants failed: %w", CANCELLED, err)
	}

	bk, err := newBookingTx(tx, &p, updatedSeat.SeatType, participantIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: booking insertion failed: %w", CANCELLED, err)
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Error(fmt.Sprintf("[create-booking-uc] ❌ Commit failed: %v", err))
		return nil, fmt.Errorf("%s: failed to commit transaction: %w", CANCELLED, err)
	}
	logger.Log.Info(fmt.Sprintf("[create-booking-uc] ✅ Booking committed successfully: %v", bk.BookingID))

	// ---- Admin notification (non-blocking) ----
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		logger.Log.Warn("[create-booking-uc] ⚠️ Admin email not configured; skipping notification")
		return bk, nil
	}

	// Use the ORIGINAL base64 from the request to avoid any re-encoding surprises.
	receiptBase64 := p.ReceiptImage

	go func() {
		logger.Log.Info(fmt.Sprintf("[create-booking-uc] ✉️ Admin email: %s", adminEmail))
		err := auth.SendBookingNotificationEmail(
			adminEmail,
			bk.BookingID.String(),
			bk.BookingEmail,
			bk.SeatType,
			bk.TotalAmount,
			receiptBase64,
		)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[create-booking-uc] ❌ Failed to send admin notification: %v", err))
		} else {
			logger.Log.Info(fmt.Sprintf("[create-booking-uc] ✉️ Admin notified of new booking: %s", bk.BookingID))
		}
	}()

	return bk, nil
}

// ---------- helpers ----------

func validateAndUpdateSeatTx(tx *sql.Tx, seatID string, quantity int) (*seat.Seat, error) {
	currentSeat, err := seat.GetSeatForUpdateTx(tx, seatID)
	if err != nil {
		return nil, err
	}

	newVal := currentSeat.Available - quantity
	if newVal < 0 {
		return nil, fmt.Errorf("%w: requested %d, available %d", ErrNotEnoughSeats, quantity, currentSeat.Available)
	}

	return seat.UpdateAvailableTx(tx, seatID, newVal)
}

func addParticipantsTx(tx *sql.Tx, details []*participantsDetails) ([]string, error) {
	var pts []string
	for _, detail := range details {
		toCreate := participant.CreateParticipantParams{
			Name:  detail.Name,
			WaNum: detail.WaNum,
			Email: detail.Email,
		}
		mSt, _ := json.Marshal(toCreate)
		newPt, err := participant.AddParticipantTx(tx, mSt)
		if err != nil {
			return nil, err
		}
		pts = append(pts, newPt.UserID)
	}
	return pts, nil
}

func newBookingTx(tx *sql.Tx, p *CreateBookingParams, seatType string, participantIDs []string) (*Booking, error) {
	bkID := uuid.New()

	participantIDsJSON, _ := json.Marshal(participantIDs)
	receiptBytes, _ := base64.StdEncoding.DecodeString(p.ReceiptImage)

	bk := &Booking{
		BookingID:        bkID,
		BookingEmail:     p.BookingEmail,
		BookingStatus:    VERIFYING,
		PaymentDetailsID: p.PaymentDetailsID,
		ReceiptImage:     receiptBytes,
		SeatQuantity:     p.SeatQuantity,
		SeatID:           p.SeatID,
		TotalAmount:      p.TotalAmount,
		SeatType:         seatType,
		ParticipantIDs:   participantIDs,
		CreatedAt:        time.Now(),
	}

	const insertSQL = `
	INSERT INTO booking (
		booking_id, booking_email, booking_status, payment_details_id,
		receipt_image, seat_quantity, seat_id, concert_id, total_amount,
		seat_type, participant_ids, created_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
`

	_, err := tx.Exec(
		insertSQL,
		bk.BookingID,
		bk.BookingEmail,
		bk.BookingStatus,
		bk.PaymentDetailsID,
		bk.ReceiptImage,
		bk.SeatQuantity,
		bk.SeatID,
		p.ConcertID,
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
