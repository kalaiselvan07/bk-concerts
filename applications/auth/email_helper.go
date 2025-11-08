package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"supra/logger"
)

const resendAPI = "https://api.resend.com/emails"
const defaultFrom = "BlackticketEntertainments <noreply@bookingx.live>"

// ---- Resend payloads ----

type Attachment struct {
	Filename string `json:"filename"`
	// Resend expects base64-encoded content
	Content string `json:"content"`
}

type ResendEmail struct {
	From        string       `json:"from"`
	To          string       `json:"to"`
	Subject     string       `json:"subject"`
	Html        string       `json:"html"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// single helper for sending (optionally with attachments)
func sendEmailResend(to, subject, htmlBody, textBody string, atts ...Attachment) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		logger.Log.Warn("[auth] Missing RESEND_API_KEY, mock email triggered.")
		fmt.Printf("\n--- MOCK EMAIL ---\nTo: %s\nSubject: %s\nBody:\n%s\nAttachments: %d\n-------------------\n",
			to, subject, htmlBody, len(atts))
		return nil
	}

	payload := ResendEmail{
		From:        defaultFrom,
		To:          to,
		Subject:     subject,
		Html:        htmlBody,
		Text:        textBody,
		Attachments: atts,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", resendAPI, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Resend API error: %s", resp.Status)
	}

	logger.Log.Info(fmt.Sprintf("[auth] ✅ Email sent to %s via Resend.", to))
	return nil
}

// ---------- Public API ----------

// OTP
func SendOTP(toEmail, code string) error {
	logger.Log.Info(fmt.Sprintf("[auth] Sending OTP email to %s", toEmail))
	html := fmt.Sprintf(`
		<h2>🔐 Your BlackTickets OTP</h2>
		<p>Your one-time login code is:</p>
		<h3 style="color:#0070f3;">%s</h3>
		<p>This code is valid for 5 minutes.</p>
	`, code)
	return sendEmailResend(toEmail, "Your BlackTickets Login OTP", html, "Your OTP code is "+code)
}

// Admin notification (includes inline preview + attachment)
func SendBookingNotificationEmail(toEmail, bookingID, userEmail, seatType string, total float64, receiptBase64 string) error {
	html := fmt.Sprintf(`
		<h2>🎟️ New Booking Notification</h2>
		<p><b>Booking ID:</b> %s</p>
		<p><b>User:</b> %s</p>
		<p><b>Seat Type:</b> %s</p>
		<p><b>Total:</b> ₹%.2f</p>
		<p>Status: <b style="color:#007bff;">Pending Verification</b></p>
		<p>Receipt (preview):</p>
		<img src="data:image/png;base64,%s" style="max-width:500px;border-radius:8px;" />
	`, bookingID, userEmail, seatType, total, receiptBase64)

	att := Attachment{Filename: "receipt.png", Content: receiptBase64}
	return sendEmailResend(toEmail, fmt.Sprintf("🆕 New Booking Created [%s]", bookingID), html, "", att)
}

// Re-upload notification (with Approve/Reject + attachment)
func SendReceiptReuploadNotification(toEmail, bookingID, userEmail, seatType string, amount float64, base64Receipt string) error {
	html := fmt.Sprintf(`
		<h2>🔄 Receipt Re-upload Alert</h2>
		<p>User <b>%s</b> re-uploaded payment receipt for:</p>
		<ul>
			<li><b>Booking ID:</b> %s</li>
			<li><b>Seat Type:</b> %s</li>
			<li><b>Amount:</b> ₹%.2f</li>
		</ul>

		<p>Receipt (preview):</p>
		<img src="data:image/png;base64,%s" style="max-width:450px;margin-top:15px;border-radius:6px;" />
	`, userEmail, bookingID, seatType, amount, base64Receipt)

	att := Attachment{Filename: "receipt.png", Content: base64Receipt}
	return sendEmailResend(toEmail, fmt.Sprintf("🔄 Receipt Re-uploaded [%s]", bookingID), html, "", att)
}

// Status updates
func SendBookingVerificationMail(toEmail, status, bookingID, base64Receipt, note string) error {
	logger.Log.Info(fmt.Sprintf("[auth] Sending booking verification email (Status: %s) to %s", status, toEmail))
	statusUpper := strings.ToUpper(status)

	var subject, html string
	switch statusUpper {
	case "PENDING_VERIFICATION":
		subject = "🎟️ Booking Pending Verification"
		html = fmt.Sprintf(`
			<h2>🎟️ New Booking Awaiting Verification</h2>
			<p>Booking ID: <b>%s</b></p>
			<p>Status: Pending Verification</p>
			<img src="data:image/png;base64,%s" style="max-width:400px;margin-top:10px;border-radius:8px;" />
			<p><i>Approve or reject from your admin dashboard.</i></p>
		`, bookingID, base64Receipt)
	case "APPROVED":
		subject = "✅ Booking Approved"
		html = fmt.Sprintf(`
			<h2>✅ Booking Approved</h2>
			<p>Your booking <b>%s</b> has been approved!</p>
			<p>Thank you for booking with <b>BlackTickets</b>.</p>
		`, bookingID)
	case "REJECTED":
		subject = "❌ Booking Rejected"
		html = fmt.Sprintf(`
			<h2>❌ Booking Rejected</h2>
			<p>Your booking <b>%s</b> was rejected.</p>
			<p>Reason: %s</p>
			<p>If this was a mistake, please contact support.</p>
		`, bookingID, note)
	default:
		subject = "ℹ️ Booking Updated"
		html = fmt.Sprintf("<p>Booking %s updated. Status: %s</p>", bookingID, status)
	}

	return sendEmailResend(toEmail, subject, html, "")
}

// Approval mail — attach the PDF e-ticket
func SendBookingApprovalMail(toEmail, bookingID, seatType string, qty int, total float64, pdfBytes []byte) error {
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)

	html := fmt.Sprintf(`
		<h2>✅ Booking Approved!</h2>
		<p>Your booking <b>%s</b> has been approved.</p>
		<p>Seat Type: %s<br>Quantity: %d<br>Total: ₹%.2f</p>
		<p>Your e-ticket PDF is attached to this email.</p>
	`, bookingID, seatType, qty, total)

	att := Attachment{
		Filename: fmt.Sprintf("e-ticket-%s.pdf", bookingID),
		Content:  pdfBase64,
	}
	return sendEmailResend(toEmail, fmt.Sprintf("🎟️ Your BlackTickets e-Ticket [%s]", bookingID), html, "", att)
}
