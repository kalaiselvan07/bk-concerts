package booking

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"

	"bk-concerts/applications/paymentdetails"
	"bk-concerts/logger"

	"github.com/jung-kurt/gofpdf"
	qrcode "github.com/skip2/go-qrcode"
)

type ticketParticipant struct {
	Name  string
	WaNum string
	Email string
}

type paymentInfo struct {
	Type    string
	Details string
}

// GenerateTicketPDF creates a single-page eTicket with a predesigned poster and black bottom half.
func GenerateTicketPDF(bookingID string) (*Booking, []byte, error) {
	// --- Fetch Booking ---
	bk, err := GetBooking(bookingID)
	if err != nil || bk == nil {
		return nil, nil, fmt.Errorf("booking not found: %w", err)
	}

	// fetch payment details
	pd, pErr := paymentdetails.GetPayment(bk.PaymentDetailsID)
	if pErr != nil && pErr != sql.ErrNoRows {
		logger.Log.Warn(fmt.Sprintf("payment fetch error: %v", pErr))
	}

	var pay paymentInfo
	if pd != nil {
		pay = paymentInfo{Type: pd.PaymentType, Details: pd.Details}
	}

	posterPath := "resources/poster.jpg"

	// --- Initialize PDF ---
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(0, 0, 0)

	// --- Top Half: Poster ---
	if _, err := os.Stat(posterPath); err == nil {
		pdf.ImageOptions(posterPath, 0, 0, 210, 145, false,
			gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true}, 0, "")
	} else {
		fmt.Println("⚠️ Poster not found, skipping top half.")
	}

	// --- Bottom Half: Ticket Section ---
	startY := 145.0
	pdf.SetFillColor(0, 0, 0)
	pdf.Rect(0, startY, 210, 152, "F")

	// --- Header ---
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetXY(20, startY+10)
	pdf.Cell(0, 10, "BLACKTICKETS OFFICIAL e-TICKET")

	// Accent line
	pdf.SetDrawColor(216, 27, 96)
	pdf.SetLineWidth(0.8)
	pdf.Line(20, startY+22, 190, startY+22)

	// --- Left Column: Booking + Payment Details ---
	leftX := 20.0
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(216, 27, 96)
	pdf.SetXY(leftX, startY+35)
	pdf.Cell(0, 8, "BOOKING DETAILS")

	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(230, 230, 230)
	pdf.SetX(leftX)
	pdf.Cell(0, 8, fmt.Sprintf("Booking ID: %s", bk.BookingID))
	pdf.Ln(7)
	pdf.SetX(leftX)
	pdf.Cell(0, 8, fmt.Sprintf("Email: %s", bk.BookingEmail))
	pdf.Ln(7)
	pdf.SetX(leftX)
	pdf.Cell(0, 8, fmt.Sprintf("Seat Type: %s", bk.SeatType))
	pdf.Ln(7)
	pdf.SetX(leftX)
	pdf.Cell(0, 8, fmt.Sprintf("Quantity: %d", bk.SeatQuantity))
	pdf.Ln(7)
	pdf.SetX(leftX)
	pdf.Cell(0, 8, fmt.Sprintf("Total Paid: ₹%.2f", bk.TotalAmount))
	pdf.Ln(10)

	// Payment Info
	pdf.SetX(leftX)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(216, 27, 96)
	pdf.Cell(0, 8, "PAYMENT INFORMATION")

	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(230, 230, 230)
	if pay.Type != "" {
		pdf.SetX(leftX)
		pdf.Cell(0, 8, fmt.Sprintf("Method: %s", pay.Type))
		pdf.Ln(7)
	}
	if pay.Details != "" {
		pdf.SetX(leftX)
		pdf.MultiCell(0, 8, fmt.Sprintf("Details: %s", pay.Details), "", "", false)
		pdf.Ln(3)
	}

	// --- Right Column: QR Code ---
	qrURL := fmt.Sprintf("https://bkentertainments.vercel.app/api/v1/participants-details/%s", bk.BookingID)
	qrBytes, _ := qrcode.Encode(qrURL, qrcode.Medium, 256)
	pdf.RegisterImageOptionsReader("qr", gofpdf.ImageOptions{ImageType: "png"}, bytes.NewReader(qrBytes))
	qrX, qrY, qrSize := 135.0, startY+45, 55.0

	// QR box with glow border
	pdf.SetDrawColor(216, 27, 96)
	pdf.SetLineWidth(0.6)
	pdf.Rect(qrX-2, qrY-2, qrSize+4, qrSize+4, "")
	pdf.ImageOptions("qr", qrX, qrY, qrSize, 0, false, gofpdf.ImageOptions{ImageType: "png"}, 0, "")

	pdf.SetY(qrY + qrSize + 6)
	pdf.SetX(qrX)
	pdf.SetFont("Helvetica", "I", 10)
	pdf.SetTextColor(200, 200, 200)
	pdf.Cell(0, 6, "Scan QR for Verification")

	// --- Footer ---
	pdf.SetFillColor(216, 27, 96)
	pdf.Rect(0, 280, 210, 17, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetY(284)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, "© 2025 BlackTickets Entertainment", "", 0, "C", false, 0, "")

	// --- Output ---
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, nil, err
	}
	return bk, buf.Bytes(), nil
}
