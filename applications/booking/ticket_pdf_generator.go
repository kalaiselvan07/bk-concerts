package booking

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"

	"bk-concerts/applications/concert"
	"bk-concerts/applications/participant"
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

type concertInfo struct {
	Title string
	Time  string
	Venue string
}

type paymentInfo struct {
	Type    string
	Details string
}

// GenerateTicketPDF creates a single-page eTicket PDF with footer at bottom.
func GenerateTicketPDF(bookingID string) (*Booking, []byte, error) {
	// --- Fetch Booking ---
	bk, err := GetBooking(bookingID)
	if err != nil || bk == nil {
		return nil, nil, fmt.Errorf("booking not found: %w", err)
	}

	// --- Fetch dependencies ---
	cn, pd, participants := fetchBookingDependencies(bk)

	var ci concertInfo
	if cn != nil {
		ci = concertInfo{Title: cn.Title, Time: cn.Timing, Venue: cn.Venue}
	}

	var pay paymentInfo
	if pd != nil {
		pay = paymentInfo{Type: pd.PaymentType, Details: pd.Details}
	}

	// --- Initialize PDF ---
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Disable auto page breaks (we’ll control layout)
	pdf.SetAutoPageBreak(false, 0)

	// --- Header ---
	pdf.SetFont("Helvetica", "B", 22)
	pdf.Cell(0, 15, "BLACKTICKETS OFFICIAL eTICKET")

	// --- Logo ---
	logoPath := "resources/logo_resized.png"
	if _, err := os.Stat(logoPath); err == nil {
		info := pdf.RegisterImageOptions(logoPath, gofpdf.ImageOptions{})
		w, h := info.Extent()
		maxW, maxH := 35.0, 28.0
		scale := 1.0
		if w > maxW {
			scale = maxW / w
		}
		if h*scale > maxH {
			scale = maxH / h
		}
		newW := w * scale
		newH := h * scale
		xPos := 195 - newW
		yPos := 14.0
		pdf.ImageOptions(logoPath, xPos, yPos, newW, newH, false, gofpdf.ImageOptions{}, 0, "")
	}

	pdf.Ln(30)

	// --- Divider ---
	pdf.SetDrawColor(220, 220, 220)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(8)

	// --- Booking Summary + QR ---
	yStart := pdf.GetY()
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(15, yStart, 120, 55, "F")

	// Booking Summary
	pdf.SetXY(20, yStart+7)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.Cell(0, 8, "BOOKING SUMMARY")
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(0, 8, fmt.Sprintf("Booking ID: %s", bk.BookingID))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Email: %s", bk.BookingEmail))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Seat Type: %s", bk.SeatType))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Quantity: %d", bk.SeatQuantity))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("Total Paid: %.2f", bk.TotalAmount))

	// QR
	qrURL := fmt.Sprintf("https://bkentertainments.vercel.app/api/v1/admin/mark-attended/%s", bk.BookingID)
	qrBytes, _ := qrcode.Encode(qrURL, qrcode.Medium, 256)
	pdf.RegisterImageOptionsReader("qr", gofpdf.ImageOptions{ImageType: "png"}, bytes.NewReader(qrBytes))
	pdf.ImageOptions("qr", 145, yStart+5, 45, 0, false, gofpdf.ImageOptions{ImageType: "png"}, 0, "")

	pdf.SetY(yStart + 63)
	pdf.SetFont("Helvetica", "I", 10)
	pdf.Cell(0, 6, "Scan this QR code for entry verification.")
	pdf.Ln(8)

	// --- Concert Details ---
	drawSectionTitle(pdf, "CONCERT DETAILS")
	pdf.SetFont("Helvetica", "", 12)
	if ci.Title != "" {
		pdf.Cell(0, 8, fmt.Sprintf("Title: %s", ci.Title))
		pdf.Ln(6)
	}
	if ci.Time != "" {
		pdf.Cell(0, 8, fmt.Sprintf("Date & Time: %s", ci.Time))
		pdf.Ln(6)
	}
	if ci.Venue != "" {
		pdf.Cell(0, 8, fmt.Sprintf("Venue: %s", ci.Venue))
		pdf.Ln(8)
	}

	// --- Payment Info ---
	drawSectionTitle(pdf, "PAYMENT INFORMATION")
	pdf.SetFont("Helvetica", "", 12)
	if pay.Type != "" {
		pdf.Cell(0, 8, fmt.Sprintf("Method: %s", pay.Type))
		pdf.Ln(6)
	}
	if pay.Details != "" {
		pdf.MultiCell(0, 8, fmt.Sprintf("Details: %s", pay.Details), "", "", false)
		pdf.Ln(6)
	}

	// --- Participants ---
	drawSectionTitle(pdf, "PARTICIPANTS")
	pdf.SetFont("Helvetica", "", 12)
	maxParticipants := 6 // limit to keep single-page fit
	for i, p := range participants {
		if i >= maxParticipants {
			pdf.Cell(0, 8, fmt.Sprintf("... and %d more participants", len(participants)-maxParticipants))
			pdf.Ln(6)
			break
		}
		pdf.Cell(0, 8, fmt.Sprintf("%d. %s | %s", i+1, p.Name, p.WaNum))
		pdf.Ln(6)
		if p.Email != "" {
			pdf.Cell(0, 8, fmt.Sprintf("   Email: %s", p.Email))
			pdf.Ln(6)
		}
	}
	pdf.Ln(6)

	// --- Footer at bottom of page ---
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, 285, 195, 285)
	pdf.SetY(288)
	pdf.SetFont("Helvetica", "I", 10)
	pdf.CellFormat(0, 8, "© 2025 BlackTickets. All Rights Reserved.", "", 0, "C", false, 0, "")

	// --- Output ---
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, nil, err
	}
	return bk, buf.Bytes(), nil
}

// drawSectionTitle adds consistent section headers
func drawSectionTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(0, 9, title, "", 1, "L", true, 0, "")
	pdf.Ln(3)
}

// --- Dependency fetcher (unchanged) ---
func fetchBookingDependencies(bk *Booking) (
	concertDetails *concert.Concert,
	paymentInfo *paymentdetails.PaymentDetails,
	participants []ticketParticipant) {

	cn, err := concert.GetConcert(bk.ConcertID)
	if err != nil && err != sql.ErrNoRows {
		logger.Log.Warn(fmt.Sprintf("concert fetch error: %v", err))
	}
	concertDetails = cn

	pd, pErr := paymentdetails.GetPayment(bk.PaymentDetailsID)
	if pErr != nil && pErr != sql.ErrNoRows {
		logger.Log.Warn(fmt.Sprintf("payment fetch error: %v", pErr))
	}
	paymentInfo = pd

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
