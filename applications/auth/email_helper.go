package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"mime/quotedprintable"
	"net/smtp"
	"os"
	"strings"

	"bk-concerts/logger"
)

const WARNINGMSG = "SMTP credentials missing from environment. Using mock sender."

// 🔹 SendOTP sends login OTP email
func SendOTP(toEmail, code string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	logger.Log.Info(fmt.Sprintf("[auth] Attempting to send OTP email to: %s", toEmail))

	if host == "" || user == "" || pass == "" {
		logger.Log.Warn(fmt.Sprintf("[auth] %s", WARNINGMSG))
		sendEmailMock(toEmail, code)
		return fmt.Errorf(WARNINGMSG)
	}

	subject := "Subject: BlackTickets Login OTP\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\r\n"
	body := fmt.Sprintf("Your one-time login code is: %s\n\nThis code is valid for 5 minutes.", code)

	msg := []byte(subject + mime + "\r\n" + body)
	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	logger.Log.Info(fmt.Sprintf("[auth] Sending via SMTP server: %s", addr))
	err := smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to send OTP to %s: %v", toEmail, err))
		return fmt.Errorf("failed to send OTP to %s: %w", toEmail, err)
	}

	logger.Log.Info(fmt.Sprintf("[auth] OTP email successfully delivered to %s.", toEmail))
	return nil
}

// SendBookingNotificationEmail sends a simple HTML email with booking details and receipt image (no approve/reject links)
func SendBookingNotificationEmail(toEmail, bookingID, userEmail, seatType string, total float64, receiptBase64 string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	if host == "" || user == "" || pass == "" {
		logger.Log.Warn("[auth] SMTP credentials missing; skipping admin email.")
		return fmt.Errorf("SMTP not configured")
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)
	boundary := "BOUNDARY-NOTIFY-12345"

	htmlBody := fmt.Sprintf(`
--%s
Content-Type: text/html; charset="UTF-8"

<html>
  <body style="font-family:Arial,sans-serif; color:#333;">
    <h2>🎟️ New Booking Notification</h2>
    <p><b>Booking ID:</b> %s</p>
    <p><b>User Email:</b> %s</p>
    <p><b>Seat Type:</b> %s</p>
    <p><b>Total Amount:</b> ₹%.2f</p>
    <p>Status: <b style="color:#007bff;">Pending Verification</b></p>
    <p style="margin-top:20px;">Attached below is the receipt uploaded by the user.</p>
    <img src="cid:receiptImage" style="max-width:500px; border-radius:8px; margin-top:10px;" alt="Receipt"/>
    <p style="font-size:12px; color:#777; margin-top:30px;">This is an automated system notification from BlackTickets.</p>
  </body>
</html>

--%s
Content-Type: image/png
Content-Transfer-Encoding: base64
Content-ID: <receiptImage>

%s
--%s--
`, boundary, bookingID, userEmail, seatType, total, boundary, receiptBase64, boundary)

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: 🆕 New Booking Created [%s]\r\nMIME-Version: 1.0\r\nContent-Type: multipart/related; boundary=%s\r\n\r\n",
		user, toEmail, bookingID, boundary,
	)

	fullMessage := []byte(headers + htmlBody)
	if err := smtp.SendMail(addr, auth, user, []string{toEmail}, fullMessage); err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to send booking notification: %v", err))
		return err
	}

	logger.Log.Info(fmt.Sprintf("[auth] 📧 Booking notification sent to admin %s for booking %s", toEmail, bookingID))
	return nil
}

// 🔸 Developer mock email sender
func sendEmailMock(email, code string) {
	fmt.Printf("\n--- MOCK EMAIL SENT ---\n")
	fmt.Printf("To: %s\n", email)
	fmt.Printf("Your One-Time Code is: %s\n", code)
	fmt.Printf("-----------------------\n")
}

// SendReceiptReuploadNotification notifies the admin when a user re-uploads their payment receipt.
func SendReceiptReuploadNotification(toEmail, bookingID, userEmail, seatType string, amount float64, base64Receipt, approveURL, rejectURL string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	logger.Log.Info(fmt.Sprintf("[auth] Sending receipt re-upload notification for booking %s to admin %s", bookingID, toEmail))

	if host == "" || user == "" || pass == "" {
		logger.Log.Warn("[auth] SMTP credentials missing; skipping actual send.")
		return fmt.Errorf("SMTP credentials missing")
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)
	boundary := "BOUNDARY-REUPLOAD-12345"

	// 🧾 Email HTML body
	htmlBody := fmt.Sprintf(`
--%s
Content-Type: text/html; charset="UTF-8"

<html>
  <body style="font-family: Arial, sans-serif; color: #333;">
    <h2>🔄 Receipt Re-upload Alert</h2>
    <p>User <b>%s</b> has re-uploaded their payment receipt for verification.</p>

    <table style="border-collapse: collapse; margin-top: 10px;">
      <tr><td><b>Booking ID:</b></td><td>%s</td></tr>
      <tr><td><b>Seat Type:</b></td><td>%s</td></tr>
      <tr><td><b>Total:</b></td><td>₹%.2f</td></tr>
      <tr><td><b>Status:</b></td><td><span style="color:#007bff;">Pending Verification</span></td></tr>
    </table>

    <p style="margin-top:15px;">Receipt Image:</p>
    <img src="cid:receiptImage" alt="Receipt" style="max-width: 450px; border: 1px solid #ccc; border-radius: 8px;">

    <p style="margin-top: 25px;">
      <a href="%s" style="background-color: #28a745; color: white; padding: 10px 18px; border-radius: 6px; text-decoration: none; font-weight: bold;">✅ Approve</a>
      <a href="%s" style="background-color: #dc3545; color: white; padding: 10px 18px; border-radius: 6px; text-decoration: none; font-weight: bold; margin-left: 10px;">❌ Reject</a>
    </p>

    <p style="margin-top:20px; font-size:12px; color:#777;">
      This notification was automatically generated by <b>BlackTickets</b> after a user re-uploaded their receipt.
    </p>
  </body>
</html>

--%s
Content-Type: image/png
Content-Transfer-Encoding: base64
Content-ID: <receiptImage>

%s
--%s--
`, boundary, userEmail, bookingID, seatType, amount, approveURL, rejectURL, boundary, base64Receipt, boundary)

	// 📬 Headers
	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: 🔄 Receipt Re-uploaded for Booking %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/related; boundary=%s\r\n\r\n",
		user, toEmail, bookingID, boundary,
	)

	fullMessage := []byte(headers + htmlBody)

	// ✉️ Send email
	if err := smtp.SendMail(addr, auth, user, []string{toEmail}, fullMessage); err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to send re-upload notification: %v", err))
		return err
	}

	logger.Log.Info(fmt.Sprintf("[auth] ✅ Receipt re-upload notification sent to %s for booking %s", toEmail, bookingID))
	return nil
}

// SendBookingVerificationMail sends HTML-based booking verification or notification emails.
func SendBookingVerificationMail(toEmail, status, bookingID, base64Receipt, note string) error {
	logger.Log.Info(fmt.Sprintf("[auth] Preparing booking verification email for %s (Status: %s)", toEmail, status))

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	adminMail := os.Getenv("SMTP_USER") // Optional fallback recipient

	if host == "" || user == "" || pass == "" {
		logger.Log.Warn("[auth] Missing SMTP credentials. Using mock email sender.")
		sendBookingMailMock(toEmail, status, bookingID, note)
		return fmt.Errorf("missing SMTP credentials")
	}

	if strings.TrimSpace(toEmail) == "" && adminMail != "" {
		toEmail = adminMail
	}

	// --- Build HTML body ---
	subject := fmt.Sprintf("Subject: BlackTickets Booking %s Notification\r\n", strings.Title(strings.ToLower(status)))
	mime := "MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=BOUNDARY\r\n\r\n"

	var htmlBody string
	switch strings.ToUpper(status) {
	case "PENDING_VERIFICATION":
		htmlBody = fmt.Sprintf(`
			<html>
			<body style="font-family:Arial, sans-serif; line-height:1.6;">
				<h2 style="color:#0070f3;">🎟️ New Booking Awaiting Verification</h2>
				<p>A new booking (<strong>%s</strong>) requires your attention.</p>
				<p>Please verify the payment receipt and approve or reject:</p>
				<p>
					<a href="https://your-domain.com/dashboard/bookings/%s?action=approve" style="background-color:#28a745;color:white;padding:10px 16px;text-decoration:none;border-radius:6px;">Approve</a>
					<a href="https://your-domain.com/dashboard/bookings/%s?action=reject" style="background-color:#dc3545;color:white;padding:10px 16px;text-decoration:none;border-radius:6px;margin-left:8px;">Reject</a>
				</p>
				<p><i>This booking will remain pending until verified by admin.</i></p>
			</body>
			</html>
		`, bookingID, bookingID, bookingID)

	case "APPROVED":
		htmlBody = fmt.Sprintf(`
			<html>
			<body style="font-family:Arial, sans-serif; line-height:1.6;">
				<h2 style="color:#28a745;">✅ Booking Approved</h2>
				<p>Dear Customer, your booking <strong>%s</strong> has been successfully approved.</p>
				<p>Your e-ticket is attached below. Please bring it to the event.</p>
				<p>Thank you for booking with <strong>BlackTickets</strong>!</p>
			</body>
			</html>
		`, bookingID)

	case "REJECTED":
		htmlBody = fmt.Sprintf(`
			<html>
			<body style="font-family:Arial, sans-serif; line-height:1.6;">
				<h2 style="color:#dc3545;">❌ Booking Rejected</h2>
				<p>Dear Customer, your booking <strong>%s</strong> could not be verified.</p>
				<p><strong>Reason:</strong> %s</p>
				<p>If you believe this was a mistake, please contact support or re-upload your payment receipt.</p>
			</body>
			</html>
		`, bookingID, note)
	default:
		htmlBody = fmt.Sprintf(`<html><body><p>Booking %s updated. Status: %s</p></body></html>`, bookingID, status)
	}

	// --- Create the MIME multipart message ---
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.SetBoundary("BOUNDARY")

	// 1️⃣ HTML body part
	htmlPart, _ := writer.CreatePart(map[string][]string{
		"Content-Type":              {"text/html; charset=UTF-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	qp := quotedprintable.NewWriter(htmlPart)
	_, _ = qp.Write([]byte(htmlBody))
	_ = qp.Close()

	// 2️⃣ Optional receipt image attachment
	if base64Receipt != "" {
		decoded, err := base64.StdEncoding.DecodeString(base64Receipt)
		if err == nil {
			partHeader := map[string][]string{
				"Content-Type":              {"image/png"},
				"Content-Transfer-Encoding": {"base64"},
				"Content-Disposition":       {`attachment; filename="receipt.png"`},
			}
			part, _ := writer.CreatePart(partHeader)
			part.Write(decoded)
		} else {
			logger.Log.Warn(fmt.Sprintf("[auth] Failed to decode base64 receipt for booking %s: %v", bookingID, err))
		}
	}

	_ = writer.Close()

	msg := []byte(subject + mime + body.String())
	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	// --- Send email ---
	err := smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to send booking email to %s: %v", toEmail, err))
		return fmt.Errorf("failed to send booking email: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[auth] Booking email sent to %s successfully.", toEmail))
	return nil
}

// Mock fallback (when SMTP disabled)
func sendBookingMailMock(email, status, bookingID, note string) {
	fmt.Printf("\n--- MOCK BOOKING EMAIL ---\n")
	fmt.Printf("TO: %s\nSTATUS: %s\nBOOKING ID: %s\nNOTE: %s\n", email, status, bookingID, note)
	fmt.Println("--------------------------")
}

// SendBookingApprovalMail sends booking approval email with attached PDF ticket.
func SendBookingApprovalMail(toEmail, bookingID, seatType string, qty int, total float64, pdfBytes []byte) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	if host == "" || user == "" || pass == "" {
		logger.Log.Warn("[auth] Missing SMTP credentials, cannot send ticket email.")
		return fmt.Errorf("missing SMTP credentials")
	}

	boundary := "TICKET-BOUNDARY-123"
	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	htmlBody := fmt.Sprintf(`
--%s
Content-Type: text/html; charset="UTF-8"

<html>
  <body style="font-family:Arial, sans-serif;">
    <h2 style="color:#28a745;">✅ Booking Approved!</h2>
    <p>Your booking <b>%s</b> has been approved.</p>
    <p>Seat Type: <b>%s</b><br>
    Quantity: <b>%d</b><br>
    Total Amount: <b>₹%.2f</b></p>
    <p>Your e-ticket is attached to this email.</p>
    <p style="margin-top:20px;">Thank you for choosing <b>BlackTickets</b>!</p>
  </body>
</html>
`, boundary, bookingID, seatType, qty, total)

	pdfEncoded := base64.StdEncoding.EncodeToString(pdfBytes)

	htmlBody += fmt.Sprintf(`
--%s
Content-Type: application/pdf
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="BlackTickets_%s.pdf"

%s
--%s--
`, boundary, bookingID, pdfEncoded, boundary)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: 🎟️ Your BlackTickets e-Ticket\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n", user, toEmail, boundary)

	msg := []byte(headers + htmlBody)
	err := smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[auth] Failed to send ticket PDF email: %v", err))
		return err
	}

	logger.Log.Info(fmt.Sprintf("[auth] Ticket email with PDF sent to %s", toEmail))
	return nil
}
