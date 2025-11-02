package booking

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"bk-concerts/applications/participant"
	"bk-concerts/applications/paymentdetails"
	"bk-concerts/applications/seat"
	"bk-concerts/logger"

	"github.com/jung-kurt/gofpdf"
	qrcode "github.com/skip2/go-qrcode"
)

type paymentInfo struct {
	Type    string
	Details string
}

// SafeAddImage safely adds an image with automatic format detection.
func SafeAddImage(pdf *gofpdf.Fpdf, path string, x, y, width float64, readDpi bool) {
	if _, err := os.Stat(path); err != nil {
		logger.Log.Warn(fmt.Sprintf("[generate-ticket-uc] Image not found: %s", path))
		return
	}

	ext := filepath.Ext(path)
	imgType := ""
	switch ext {
	case ".png", ".PNG":
		imgType = "PNG"
	case ".jpg", ".jpeg", ".JPG", ".JPEG":
		imgType = "JPG"
	default:
		logger.Log.Warn(fmt.Sprintf("[generate-ticket-uc] Unknown image format: %s", path))
		return
	}

	pdf.ImageOptions(path, x, y, width, 0, false,
		gofpdf.ImageOptions{ImageType: imgType, ReadDpi: readDpi}, 0, "")
}

// GenerateTicketPDF creates an elite e-ticket design with structured layout.
func GenerateTicketPDF(bookingID string) (*Booking, []byte, error) {
	logger.Log.Info(fmt.Sprintf("[generate-ticket-uc] Generating ticket for bookingID: %s", bookingID))

	// --- Fetch Booking ---
	bk, err := GetBooking(bookingID)
	if err != nil || bk == nil {
		return nil, nil, fmt.Errorf("booking not found: %w", err)
	}

	st, err := seat.GetSeat(bk.SeatID)
	if err != nil || st == nil {
		return nil, nil, fmt.Errorf("seat not found: %w", err)
	}
	gelAmount := float64(bk.SeatQuantity) * st.PriceGel

	// --- Fetch Payment Info ---
	pd, _ := paymentdetails.GetPayment(bk.PaymentDetailsID)
	var pay paymentInfo
	if pd != nil {
		pay = paymentInfo{Type: pd.PaymentType, Details: pd.Details}
	}

	// --- Fetch Participants ---
	var participants []*participant.Participant
	for _, pid := range bk.ParticipantIDs {
		p, err := participant.GetParticipant(pid)
		if err == nil && p != nil {
			participants = append(participants, p)
		}
	}

	posterPath := "resources/asal.jpg"
	logoPath := "resources/whitelogo.png"

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetMargins(0, 0, 0)

	// --- Poster Banner ---
	SafeAddImage(pdf, posterPath, 0, 0, 210, true)

	// --- QR Code ---
	qrURL := fmt.Sprintf("https://bkentertainments.vercel.app/concerts/participants/?bookingID=%s", bk.BookingID)
	qrBytes, _ := qrcode.Encode(qrURL, qrcode.Medium, 512)
	qrX, qrY, qrSize := 170.0, 17.0, 35.0
	pdf.RegisterImageOptionsReader("qr",
		gofpdf.ImageOptions{ImageType: "png"}, bytes.NewReader(qrBytes))
	pdf.ImageOptions("qr", qrX, qrY, qrSize, qrSize, false,
		gofpdf.ImageOptions{ImageType: "png"}, 0, "")

	// --- QR Caption ---
	pdf.SetY(qrY + qrSize + 5)
	pdf.SetX(qrX - 6)
	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(210, 210, 210)
	pdf.CellFormat(45, 5, "Scan QR for Participant Details", "", 0, "C", false, 0, "")

	// --- Black Ticket Section ---
	startY := 68.0
	pdf.SetFillColor(0, 0, 0)
	pdf.Rect(0, startY, 210, 212, "F")

	// --- Header ---
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetXY(20, startY+8)
	pdf.Cell(0, 10, "BLACKTICKETS OFFICIAL e-TICKET")

	// Accent line
	pdf.SetDrawColor(216, 27, 96)
	pdf.SetLineWidth(0.8)
	pdf.Line(20, startY+20, 190, startY+20)

	// --- Booking Details (Enhanced Card Layout) ---
	leftX := 20.0
	pdf.SetFont("Helvetica", "B", 15)
	pdf.SetTextColor(216, 27, 96)
	pdf.SetXY(leftX, startY+30)
	pdf.Cell(0, 8, "BOOKING DETAILS")

	// Card Background
	pdf.SetFillColor(15, 15, 15)
	pdf.RoundedRect(leftX-2, startY+40, 172, 47, 3, "1234", "F")

	// Details
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(235, 235, 235)
	pdf.SetXY(leftX+5, startY+46)
	pdf.Cell(60, 8, "Booking ID")
	pdf.Cell(0, 8, fmt.Sprintf(": %s", bk.BookingID))
	pdf.Ln(7)

	pdf.SetX(leftX + 5)
	pdf.Cell(60, 8, "Email")
	pdf.Cell(0, 8, fmt.Sprintf(": %s", bk.BookingEmail))
	pdf.Ln(7)

	pdf.SetX(leftX + 5)
	pdf.Cell(60, 8, "Seat Type")
	pdf.Cell(0, 8, fmt.Sprintf(": %s", bk.SeatType))
	pdf.Ln(7)

	pdf.SetX(leftX + 5)
	pdf.Cell(60, 8, "Quantity")
	pdf.Cell(0, 8, fmt.Sprintf(": %d", bk.SeatQuantity))
	pdf.Ln(7)

	pdf.SetX(leftX + 5)
	pdf.Cell(60, 8, "Total Paid")
	pdf.Cell(0, 8, fmt.Sprintf(": %.2f INR / %.2f GEL", bk.TotalAmount, gelAmount))

	// --- Logo beside Booking Details ---
	if _, err := os.Stat(logoPath); err == nil {
		SafeAddImage(pdf, logoPath, 160, startY+54, 35, false)
	}

	// --- Payment Info ---
	pdf.Ln(18)
	pdf.SetX(leftX)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(216, 27, 96)
	pdf.Cell(0, 8, "PAYMENT INFORMATION")

	pdf.Ln(9)
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(230, 230, 230)
	if pay.Type != "" {
		pdf.SetX(leftX + 5)
		pdf.Cell(0, 8, fmt.Sprintf("Method: %s", pay.Type))
		pdf.Ln(7)
	}
	if pay.Details != "" {
		pdf.SetX(leftX + 5)
		pdf.MultiCell(0, 8, fmt.Sprintf("Details: %s", pay.Details), "", "", false)
	}

	// --- Participant Details ---
	pdf.Ln(10)
	pdf.SetX(leftX)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(216, 27, 96)
	pdf.Cell(0, 8, "PARTICIPANT DETAILS")

	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(230, 230, 230)

	if len(participants) == 0 {
		pdf.SetX(leftX + 5)
		pdf.Cell(0, 8, "No participants found.")
	} else {
		// Divide participants into two columns (left/right)
		half := (len(participants) + 1) / 2
		leftCol := participants[:half]
		var rightCol []*participant.Participant
		if len(participants) > half {
			rightCol = participants[half:]
		}

		// Dynamic box sizing
		boxTop := pdf.GetY() - 2
		lineHeight := 5.0
		rowSpacing := 1.5
		colWidth := 85.0

		// Calculate box height dynamically based on row count
		rowCount := len(leftCol)
		if len(rightCol) > rowCount {
			rowCount = len(rightCol)
		}
		boxHeight := float64(rowCount)*(lineHeight*3+rowSpacing) + 8

		pdf.SetDrawColor(60, 60, 60)
		pdf.RoundedRect(leftX-2, boxTop, 172, boxHeight, 2, "1234", "D")

		// Render participants in two columns
		for i := 0; i < rowCount; i++ {
			y := boxTop + 6 + float64(i)*(lineHeight*3+rowSpacing)

			// Left column
			if i < len(leftCol) {
				p := leftCol[i]
				pdf.SetXY(leftX+5, y)
				pdf.Cell(0, 5, fmt.Sprintf("%d. %s", i+1, p.Name))

				pdf.SetXY(leftX+10, y+lineHeight)
				pdf.Cell(0, 5, fmt.Sprintf("WA: %s", p.WaNum))

				if p.Email != "" {
					pdf.SetXY(leftX+10, y+lineHeight*2)
					pdf.Cell(0, 5, fmt.Sprintf("Email: %s", p.Email))
				}
			}

			// Right column
			if i < len(rightCol) {
				p := rightCol[i]
				pdf.SetXY(leftX+colWidth, y)
				pdf.Cell(0, 5, fmt.Sprintf("%d. %s", i+half+1, p.Name))

				pdf.SetXY(leftX+colWidth+5, y+lineHeight)
				pdf.Cell(0, 5, fmt.Sprintf("WA: %s", p.WaNum))

				if p.Email != "" {
					pdf.SetXY(leftX+colWidth+5, y+lineHeight*2)
					pdf.Cell(0, 5, fmt.Sprintf("Email: %s", p.Email))
				}
			}
		}
	}

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

	logger.Log.Info(fmt.Sprintf("[generate-ticket-uc] ✅ PDF generated successfully for bookingID: %s", bk.BookingID))
	return bk, buf.Bytes(), nil
}
