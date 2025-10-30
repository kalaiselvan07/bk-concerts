package booking

import (
	"context"
	"encoding/json"
	"fmt"

	"bk-concerts/db"
	"bk-concerts/logger"
)

// GetPendingBookingsUC retrieves all bookings awaiting verification by admin.
// These are bookings with status VERIFYING or PENDING_VERIFICATION.
func GetPendingBookingsUC() ([]*Booking, error) {
	logger.Log.Info("[get-pending-bookings-uc] Fetching all pending bookings for admin review.")

	query := `
		SELECT 
			booking_id, booking_email, booking_status, payment_details_id,
			receipt_image, seat_quantity, seat_id, concert_id, total_amount, seat_type,
			participant_ids, created_at
		FROM booking
		WHERE booking_status IN ('VERIFYING', 'PENDING_VERIFICATION')
		ORDER BY created_at DESC
	`

	ctx := context.Background()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-pending-bookings-uc] Database query failed: %v", err))
		return nil, fmt.Errorf("failed to fetch pending bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		var (
			bk                Booking
			participantIDsRaw []byte
			receiptBytes      []byte
		)

		if err := rows.Scan(
			&bk.BookingID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
			&receiptBytes, &bk.SeatQuantity, &bk.SeatID, &bk.ConcertID, &bk.TotalAmount,
			&bk.SeatType, &participantIDsRaw, &bk.CreatedAt,
		); err != nil {
			logger.Log.Warn(fmt.Sprintf("[get-pending-bookings-uc] Row scan failed: %v", err))
			continue
		}

		if len(participantIDsRaw) > 0 {
			_ = json.Unmarshal(participantIDsRaw, &bk.ParticipantIDs)
		}

		bk.ReceiptImage = receiptBytes
		bookings = append(bookings, &bk)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-pending-bookings-uc] Row iteration error: %v", err))
		return nil, fmt.Errorf("failed during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-pending-bookings-uc] Retrieved %d pending bookings.", len(bookings)))
	return bookings, nil
}
