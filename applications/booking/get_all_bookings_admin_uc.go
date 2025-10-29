package booking

import (
	"context"
	"encoding/json"
	"fmt"

	"bk-concerts/db"
	"bk-concerts/logger"
)

// GetAllBookingsAdminUC retrieves all bookings (regardless of status) for admin dashboard.
func GetAllBookingsAdminUC() ([]*Booking, error) {
	logger.Log.Info("[get-all-bookings-admin-uc] Fetching all bookings for admin view.")

	query := `
		SELECT 
			booking_id, booking_email, booking_status, payment_details_id,
			receipt_image, seat_quantity, seat_id, total_amount, seat_type,
			participant_ids, created_at
		FROM booking
		ORDER BY created_at DESC
	`

	ctx := context.Background()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-bookings-admin-uc] Database query failed: %v", err))
		return nil, fmt.Errorf("failed to fetch bookings: %w", err)
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
			&receiptBytes, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount,
			&bk.SeatType, &participantIDsRaw, &bk.CreatedAt,
		); err != nil {
			logger.Log.Warn(fmt.Sprintf("[get-all-bookings-admin-uc] Row scan failed: %v", err))
			continue
		}

		if len(participantIDsRaw) > 0 {
			_ = json.Unmarshal(participantIDsRaw, &bk.ParticipantIDs)
		}

		bk.ReceiptImage = receiptBytes
		bookings = append(bookings, &bk)
	}

	if err := rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-bookings-admin-uc] Row iteration error: %v", err))
		return nil, fmt.Errorf("failed during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-bookings-admin-uc] Retrieved %d total bookings for admin.", len(bookings)))
	return bookings, nil
}
