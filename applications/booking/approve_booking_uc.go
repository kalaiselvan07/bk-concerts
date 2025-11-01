package booking

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bk-concerts/applications/auth"
	"bk-concerts/applications/concert"
	"bk-concerts/applications/participant"
	"bk-concerts/applications/paymentdetails"
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

	// --- Fetch Booking ---
	bk, err := GetBooking(bookingID)
	if err != nil || bk == nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// --- Fetch dependent data ---
	concertDetails, paymentInfo, participants := fetchBookingDependencies(bk)

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
	pdfBytes, err := GenerateTicketPDF(bk, concertDetails, paymentInfo, participants)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] ⚠️ PDF generation failed: %v", err))
	}

	// --- Send Email with PDF ---
	if emailErr := auth.SendBookingApprovalMail(
		bk.BookingEmail,
		bk.BookingID,
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

// fetchBookingDependencies gathers concert, payment, and participants data.
func fetchBookingDependencies(bk *Booking) (
	concertDetails *concert.Concert,
	paymentInfo *paymentdetails.PaymentDetails,
	participants []ticketParticipant,
) {
	// --- Concert ---
	cn, err := concert.GetConcert(bk.ConcertID)
	if err != nil && err != sql.ErrNoRows {
		logger.Log.Warn(fmt.Sprintf("concert fetch error: %v", err))
	}
	concertDetails = cn

	// --- Payment ---
	pd, pErr := paymentdetails.GetPayment(bk.PaymentDetailsID)
	if pErr != nil && pErr != sql.ErrNoRows {
		logger.Log.Warn(fmt.Sprintf("payment fetch error: %v", pErr))
	}
	paymentInfo = pd

	// --- Participants ---
	for _, pid := range bk.ParticipantIDs {
		pt, err := participant.GetParticipant(pid)
		if err != nil || pt == nil {
			logger.Log.Warn(fmt.Sprintf("participant fetch failed for ID: %s, %v", pid, err))
			continue
		}
		participants = append(participants, ticketParticipant{
			Name:  pt.Name,
			WaNum: pt.WaNum,
			Email: pt.Email,
		})
	}
	return
}
