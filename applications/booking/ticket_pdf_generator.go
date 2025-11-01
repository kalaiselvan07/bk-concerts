package booking

import (
	"bytes"
	"fmt"

	"bk-concerts/applications/concert"
	"bk-concerts/applications/paymentdetails"

	"github.com/jung-kurt/gofpdf"
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

// GenerateTicketPDF creates the e-ticket PDF bytes for email attachment.
func GenerateTicketPDF(
	bk *Booking,
	cn *concert.Concert,
	pd *paymentdetails.PaymentDetails,
	participants []ticketParticipant,
) ([]byte, error) {

	var ci concertInfo
	if cn != nil {
		ci = concertInfo{Title: cn.Title, Time: cn.Timing, Venue: cn.Venue}
	}

	var pay paymentInfo
	if pd != nil {
		pay = paymentInfo{Type: pd.PaymentType, Details: pd.Details}
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 12, "BlackTickets - eTicket")
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
	pdf.Ln(12)

	// Concert Info
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "🎵 Concert Details")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	if ci.Title != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Concert: %s", ci.Title))
		pdf.Ln(8)
	}
	if ci.Time != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Date & Time: %s", ci.Time))
		pdf.Ln(8)
	}
	if ci.Venue != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Venue: %s", ci.Venue))
		pdf.Ln(10)
	}

	// Payment Info
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "💳 Payment Information")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	if pay.Type != "" {
		pdf.Cell(0, 10, fmt.Sprintf("Method: %s", pay.Type))
		pdf.Ln(8)
	}
	if pay.Details != "" {
		pdf.MultiCell(0, 8, fmt.Sprintf("Details: %s", pay.Details), "", "", false)
		pdf.Ln(4)
	}

	// Participants
	if len(participants) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "👥 Participants")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 12)
		for i, p := range participants {
			pdf.Cell(0, 8, fmt.Sprintf("#%d  Name: %s", i+1, p.Name))
			pdf.Ln(6)
			pdf.Cell(0, 8, fmt.Sprintf("    WhatsApp: %s", p.WaNum))
			pdf.Ln(6)
			if p.Email != "" {
				pdf.Cell(0, 8, fmt.Sprintf("    Email: %s", p.Email))
				pdf.Ln(6)
			}
			pdf.Ln(3)
		}
		pdf.Ln(6)
	}

	// Footer
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Thank you for booking with BlackTickets!")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 8, "Please present this e-ticket at entry. Have a great show!")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
