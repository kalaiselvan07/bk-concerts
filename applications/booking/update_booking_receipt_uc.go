package booking

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"bk-concerts/applications/auth"
	"bk-concerts/db"
	"bk-concerts/logger"

	"github.com/google/uuid"
)

// UpdateBookingReceiptUC updates the receipt image and sends admin a notification email.
func UpdateBookingReceiptUC(bookingID string, payload []byte) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[update-booking-receipt-uc] 🚀 Starting receipt re-upload for booking %s", bookingID))

	// Step 1️⃣ Parse request payload
	type receiptPayload struct {
		ReceiptImage string `json:"receiptImage" validate:"required"`
	}
	var p receiptPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-receipt-uc] ❌ Failed to unmarshal payload for %s: %v", bookingID, err))
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if strings.TrimSpace(p.ReceiptImage) == "" {
		return nil, fmt.Errorf("receiptImage cannot be empty")
	}

	// Step 2️⃣ Validate booking ID
	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-booking-receipt-uc] ⚠️ Invalid booking ID format: %s", bookingID))
		return nil, fmt.Errorf("invalid booking ID: %w", err)
	}

	// Step 3️⃣ Decode Base64 → binary image bytes
	decodedBytes, err := base64.StdEncoding.DecodeString(p.ReceiptImage)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-receipt-uc] ❌ Invalid base64 image for %s: %v", bookingID, err))
		return nil, fmt.Errorf("invalid base64 image: %w", err)
	}

	// Step 4️⃣ Start DB transaction
	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-receipt-uc] ❌ Failed to start transaction: %v", err))
		return nil, fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback()

	logger.Log.Info(fmt.Sprintf("[update-booking-receipt-uc] 🧾 Transaction started for booking %s", bookingID))

	// Step 5️⃣ Update DB record
	query := `
		UPDATE booking
		SET receipt_image = $2,
		    booking_status = 'PENDING_VERIFICATION'
		WHERE booking_id = $1
		RETURNING booking_id, booking_email, booking_status, payment_details_id,
		          receipt_image, seat_quantity, seat_id, total_amount, seat_type,
		          participant_ids, created_at
	`

	var (
		bk                Booking
		receiptBytes      []byte
		participantIDsRaw []byte
		idUUID            uuid.UUID
	)

	row := tx.QueryRow(query, id, decodedBytes)
	if err := row.Scan(
		&idUUID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptBytes, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount,
		&bk.SeatType, &participantIDsRaw, &bk.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[update-booking-receipt-uc] Booking not found for ID: %s", bookingID))
			return nil, fmt.Errorf("booking not found: %s", bookingID)
		}
		logger.Log.Error(fmt.Sprintf("[update-booking-receipt-uc] Database error: %v", err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Step 6️⃣ Deserialize participant IDs
	if len(participantIDsRaw) > 0 {
		if err := json.Unmarshal(participantIDsRaw, &bk.ParticipantIDs); err != nil {
			logger.Log.Warn(fmt.Sprintf("[update-booking-receipt-uc] Failed to unmarshal participant IDs: %v", err))
		}
	}
	bk.BookingID = idUUID
	bk.ReceiptImage = receiptBytes

	// Step 7️⃣ Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-booking-receipt-uc] ❌ Commit failed: %v", err))
		return nil, fmt.Errorf("commit failed: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[update-booking-receipt-uc] ✅ DB updated and committed for booking %s", bookingID))

	// Step 8️⃣ Send Admin Notification Email (no approve/reject)
	adminEmail := os.Getenv("SMTP_USER")
	if adminEmail == "" {
		logger.Log.Warn("[update-booking-receipt-uc] ⚠️ ADMIN_EMAIL not set — skipping notification.")
		return &bk, nil
	}

	encodedReceipt := base64.StdEncoding.EncodeToString(bk.ReceiptImage)

	go func() {
		err := auth.SendBookingNotificationEmail(
			adminEmail,
			bk.BookingID.String(),
			bk.BookingEmail,
			bk.SeatType,
			bk.TotalAmount,
			encodedReceipt,
		)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[update-booking-receipt-uc] ❌ Failed to send admin notification: %v", err))
		} else {
			logger.Log.Info(fmt.Sprintf("[update-booking-receipt-uc] ✉️ Admin notified of re-upload for booking %s", bookingID))
		}
	}()

	logger.Log.Info(fmt.Sprintf("[update-booking-receipt-uc] 🎯 Receipt re-upload completed for booking %s", bookingID))
	return &bk, nil
}
