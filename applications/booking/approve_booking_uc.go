package booking

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"bk-concerts/applications/auth"
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
		return nil, fmt.Errorf("invalid booking ID: %w", err)
	}

	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback()

	// --- Fetch Booking Info
	var bk Booking
	var receiptBytes, participantIDsRaw []byte
	query := `
		SELECT booking_id, booking_email, booking_status, payment_details_id,
		       receipt_image, seat_quantity, seat_id, total_amount, seat_type,
		       participant_ids, created_at
		FROM booking WHERE booking_id = $1
	`
	err = tx.QueryRow(query, id).Scan(
		&bk.BookingID, &bk.BookingEmail, &bk.BookingStatus, &bk.PaymentDetailsID,
		&receiptBytes, &bk.SeatQuantity, &bk.SeatID, &bk.TotalAmount,
		&bk.SeatType, &participantIDsRaw, &bk.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, err
	}

	_ = json.Unmarshal(participantIDsRaw, &bk.ParticipantIDs)
	bk.ReceiptImage = receiptBytes

	// --- Update status
	_, err = tx.Exec(`UPDATE booking SET booking_status='APPROVED', updated_at=$2 WHERE booking_id=$1`, id, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}
	logger.Log.Info(fmt.Sprintf("[approve-booking-uc] Booking %s marked APPROVED.", bookingID))

	// --- Generate Ticket PDF
	pdfBytes, err := generateTicketPDF(&bk)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[approve-booking-uc] Failed to generate ticket PDF: %v", err))
	} else {
		logger.Log.Info(fmt.Sprintf("[approve-booking-uc] PDF generated successfully for %s", bookingID))
	}

	emailErr := auth.SendBookingApprovalMail(
		bk.BookingEmail,
		bk.BookingID.String(),
		bk.SeatType,
		bk.SeatQuantity,
		bk.TotalAmount,
		pdfBytes,
	)
	if emailErr != nil {
		logger.Log.Warn(fmt.Sprintf("[approve-booking-uc] Booking approved, but email failed: %v", emailErr))
	} else {
		logger.Log.Info(fmt.Sprintf("[approve-booking-uc] Approval email with PDF sent to %s", bk.BookingEmail))
	}

	return &bk, nil
}

// --- Helper to create the PDF ticket ---
func generateTicketPDF(bk *Booking) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
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
	pdf.Cell(0, 10, fmt.Sprintf("Total: ₹%.2f", bk.TotalAmount))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Issued At: %s", time.Now().Format("02 Jan 2006 15:04")))
	pdf.Ln(14)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Thank you for booking with BlackTickets!")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 8, "Please present this e-ticket at entry.")
	pdf.Ln(20)

	// Add QR-like text (BookingID)
	pdf.SetFont("Courier", "B", 12)
	pdf.MultiCell(0, 8, fmt.Sprintf("Booking Ref:\n%s", bk.BookingID.String()), "1", "L", false)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
