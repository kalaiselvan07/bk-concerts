package paymentdetails

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"     // Using the correct module path
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// UpdatePaymentParams defines fields that can be optionally updated for a payment record.
// All fields use the omitempty tag to enable partial updates.
type UpdatePaymentParams struct {
	PaymentType string `json:"paymentType,omitempty"`
	Details     string `json:"details,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// NOTE: The PaymentDetails struct and GetPayment function are assumed to be defined elsewhere in this package.

// UpdatePayment performs a general update of payment details based on the payload.
func UpdatePayment(paymentID string, payload []byte) (*PaymentDetails, error) {
	logger.Log.Info(fmt.Sprintf("[update-payment-details-uc] Starting update process for PaymentID: %s", paymentID))

	var p UpdatePaymentParams

	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-payment-details-uc] Unmarshal failed for %s: %v", paymentID, err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 1. Validate ID
	id, err := uuid.Parse(paymentID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-payment-details-uc] Update failed for %s: Invalid UUID format.", paymentID))
		return nil, fmt.Errorf("invalid payment ID format: %w", err)
	}

	// 2. Build the dynamic SQL query
	sets := []string{}
	args := []interface{}{id} // Start with payment_id as the first argument ($1)
	argCounter := 2           // SQL placeholders start at $2 for the first update field

	if p.PaymentType != "" {
		sets = append(sets, fmt.Sprintf("payment_type = $%d", argCounter))
		args = append(args, p.PaymentType)
		argCounter++
	}
	if p.Details != "" {
		sets = append(sets, fmt.Sprintf("details = $%d", argCounter))
		args = append(args, p.Details)
		argCounter++
	}
	if p.Notes != "" {
		sets = append(sets, fmt.Sprintf("notes = $%d", argCounter))
		args = append(args, p.Notes)
		argCounter++
	}

	if len(sets) == 0 {
		logger.Log.Warn(fmt.Sprintf("[update-payment-details-uc] Update skipped for %s: No updatable fields provided.", paymentID))
		// No fields to update, fetch the current details and return them
		return GetPayment(paymentID)
	}

	// 3. Construct the final SQL
	updateSQL := fmt.Sprintf(`
		UPDATE payment
		SET %s
		WHERE payment_id = $1
		RETURNING payment_id, payment_type, details, notes`,
		strings.Join(sets, ", "))

	logger.Log.Info(fmt.Sprintf("[update-payment-details-uc] Executing UPDATE for %s with %d fields modified.", paymentID, len(sets)))

	// 4. Execute and scan the returned row
	pd := &PaymentDetails{}
	row := db.DB.QueryRow(updateSQL, args...)

	if err := row.Scan(
		&pd.PaymentID, &pd.PaymentType, &pd.Details, &pd.Notes,
	); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[update-payment-details-uc] Update failed for %s: Payment not found.", paymentID))
			return nil, fmt.Errorf("payment with ID %s not found", paymentID)
		}
		logger.Log.Error(fmt.Sprintf("[update-payment-details-uc] Database update error for %s: %v", paymentID, err))
		return nil, fmt.Errorf("database update error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[update-payment-details-uc] Payment %s updated successfully. Type: %s.", paymentID, pd.PaymentType))
	return pd, nil
}
