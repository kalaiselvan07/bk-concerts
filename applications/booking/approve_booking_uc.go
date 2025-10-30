package booking

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"time"

	"bk-concerts/applications/auth"
	"bk-concerts/applications/concert"
	"bk-concerts/applications/paymentdetails"
	"bk-concerts/db"
	"bk-concerts/logger"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

// ApproveBookingUC approves a booking, updates DB, and emails a ticket PDF.
func ApproveBookingUC(bookingID string) (*Booking, error) {
	logger.Log.Info(fmt.Sprintf("[approve-booking-uc] Approving booking: %s", bookingID))

	id, err := uuid.Parse(bookingID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Invalid booking UUID: %v", err))
		return nil, fmt.Errorf("invalid booking ID: %w", err)
	}

	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Failed to start transaction: %v", err))
		return nil, fmt.Errorf("transaction start failed: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op if already committed

	// --- Fetch Booking Info ---
	bk, err := GetBooking(bookingID)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Failed to fetch booking details: %v", err))
		return nil, fmt.Errorf("failed to fetch booking: %w", err)
	}
	if bk == nil {
		logger.Log.Error("[approve-booking-uc] ❌ Booking record was nil.")
		return nil, fmt.Errorf("booking not found")
	}

	logger.Log.Info(fmt.Sprintf("[approve-booking-uc] 🎟️ ConcertID raw value (type: %T): %v", bk.ConcertID, bk.ConcertID))

	// --- Fetch Concert Details ---
	var cnTitle, cnTiming, cnVenue string
	cn, cErr := concert.GetConcert(bk.ConcertID)
	if cErr != nil {
		if cErr == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] ⚠️ No concert found for ID: %v", bk.ConcertID))
		} else {
			logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Failed to fetch concert details: %v", cErr))
			return nil, fmt.Errorf("failed to fetch concert details: %w", cErr)
		}
	} else if cn != nil {
		cnTitle, cnTiming, cnVenue = cn.Title, cn.Timing, cn.Venue
		logger.Log.Info(fmt.Sprintf("[approve-booking-uc] 🎵 Concert found: %s @ %s, %s", cnTitle, cnTiming, cnVenue))
	} else {
		logger.Log.Warn("[approve-booking-uc] ⚠️ Concert returned nil, skipping PDF details.")
	}

	// --- Fetch Payment Info ---
	pd, pErr := paymentdetails.GetPayment(bk.PaymentDetailsID)
	if pErr != nil {
		if pErr == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] ⚠️ No payment found for ID: %v", bk.PaymentDetailsID))
		} else {
			logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Failed to fetch payment info: %v", pErr))
			return nil, fmt.Errorf("failed to fetch payment details: %w", pErr)
		}
	}
	if pd == nil {
		logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] ⚠️ Payment record was nil for ID: %v", bk.PaymentDetailsID))
	}

	// --- Update booking status ---
	if _, err = tx.Exec(`UPDATE booking SET booking_status='APPROVED', updated_at=$2 WHERE booking_id=$1`, id, time.Now()); err != nil {
		logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Failed to update booking status: %v", err))
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	// --- Commit transaction ---
	if err := tx.Commit(); err != nil {
		logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Commit failed: %v", err))
		return nil, fmt.Errorf("commit failed: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[approve-booking-uc] ✅ Booking %s marked APPROVED.", bookingID))

	// --- Generate eTicket PDF (with concert & payment info) ---
	paymentType, paymentDetails := "", ""
	if pd != nil {
		paymentType, paymentDetails = pd.PaymentType, pd.Details
	}

	pdfBytes, err := generateTicketPDF(bk, cnTitle, cnTiming, cnVenue, paymentType, paymentDetails)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[approve-booking-uc] ❌ Failed to generate ticket PDF: %v", err))
	} else {
		logger.Log.Info(fmt.Sprintf("[approve-booking-uc] 📄 PDF generated successfully for %s", bookingID))
	}

	// --- Send Email ---
	if emailErr := auth.SendBookingApprovalMail(
		bk.BookingEmail,
		bk.BookingID.String(),
		bk.SeatType,
		bk.SeatQuantity,
		bk.TotalAmount,
		pdfBytes,
	); emailErr != nil {
		logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] ⚠️ Booking approved, but email failed: %v", emailErr))
	} else {
		logger.Log.Info(fmt.Sprintf("[approve-booking-uc] ✉️ Approval email with PDF sent to %s", bk.BookingEmail))
	}

	return bk, nil
}

// --- Generate eTicket PDF (includes concert + payment info) ---
func generateTicketPDF(bk *Booking, concertName, concertDate, concertVenue, paymentType, paymentDetails string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// --- Header ---
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 12, "🎫 BlackTickets - eTicket")
	pdf.Ln(14)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Booking ID: %s", bk.BookingID))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Email: %s", bk.BookingEmail))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Seat Type: %s", bk.SeatType))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Quantity: %d", bk.SeatQuantity))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Total Amount: ₹%.2f", bk.TotalAmount))
	pdf.Ln(10)

	// --- Concert Info ---
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "🎵 Concert Details")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	if concertName != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Concert: %s", concertName))
		pdf.Ln(8)
	}
	if concertDate != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Date & Time: %s", concertDate))
		pdf.Ln(8)
	}
	if concertVenue != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Venue: %s", concertVenue))
		pdf.Ln(10)
	}

	// --- Payment Info ---
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "💳 Payment Information")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	if paymentType != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Payment Method: %s", paymentType))
		pdf.Ln(8)
	}
	if paymentDetails != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Payment Details: %s", paymentDetails))
		pdf.Ln(10)
	}

	// --- Footer ---
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Thank you for booking with BlackTickets!")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 8, "Please present this e-ticket at entry. Have a great show!")
	pdf.Ln(18)

	// --- Booking Ref ---
	pdf.SetFont("Courier", "B", 12)
	pdf.MultiCell(0, 8, fmt.Sprintf("Booking Reference:\n%s", bk.BookingID.String()), "1", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
