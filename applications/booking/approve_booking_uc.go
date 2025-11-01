package booking

import (
	"context"
	"fmt"
	"time"

	"bk-concerts/applications/auth"
	"bk-concerts/db"
	"bk-concerts/logger"

	"github.com/google/uuid"
)

// ApproveBookingUC approves a booking, updates DB, and emails a ticket PDF.
func ApproveBookingUC(bookingID string) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[approve-booking-uc] Processing booking approval for: %s", bookingID))

	id, err := uuid.Parse(bookingID)
	if err != nil {
		return nil, fmt.Errorf("invalid booking ID: %w", err)
	}

	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// --- Update booking status ---
	if _, err = tx.Exec(`UPDATE booking SET booking_status='APPROVED', updated_at=$2 WHERE booking_id=$1`,
		id, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[approve-booking-uc] ✅ Booking %s marked APPROVED.", bookingID))

	// --- Generate eTicket PDF ---
	bk, pdfBytes, err := GenerateTicketPDF(bookingID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] ⚠️ PDF generation failed: %v", err))
	}

	// --- Send Email with PDF ---
	if emailErr := auth.SendBookingApprovalMail(
		bk.BookingEmail,
		bookingID,
		bk.SeatType,
		bk.SeatQuantity,
		bk.TotalAmount,
		pdfBytes,
	); emailErr != nil {
		logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] ⚠️ Email sending failed: %v", emailErr))
	} else {
		logger.Log.Info(fmt.Sprintf("[approve-booking-uc] ✉️ Approval email sent to %s", bk.BookingEmail))
	}

	return bk, nil
}
