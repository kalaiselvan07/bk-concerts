package paymentdetails

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"

	"github.com/google/uuid"
)

// UpdatePaymentParams defines fields that can be optionally updated for a payment record.
// All fields use the omitempty tag to enable partial updates.
type UpdatePaymentParams struct {
	PaymentType string `json:"paymentType,omitempty"`
	Details     string `json:"details,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// UpdatePayment performs a general update of payment details based on the payload.
func UpdatePayment(paymentID string, payload []byte) (*PaymentDetails, error) {
	var p UpdatePaymentParams

	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 1. Validate ID
	id, err := uuid.Parse(paymentID)
	if err != nil {
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

	// 4. Execute and scan the returned row
	pd := &PaymentDetails{}
	row := db.DB.QueryRow(updateSQL, args...)

	if err := row.Scan(
		&pd.PaymentID, &pd.PaymentType, &pd.Details, &pd.Notes,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment with ID %s not found", paymentID)
		}
		return nil, fmt.Errorf("database update error: %w", err)
	}

	return pd, nil
}
