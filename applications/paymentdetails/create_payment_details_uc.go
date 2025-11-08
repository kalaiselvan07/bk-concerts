package paymentdetails

import (
	"encoding/json"
	"fmt"

	"supra/db"     // Using the correct module path
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The PaymentDetails struct is assumed to be defined elsewhere in this package.

// CreatePaymentParams is used for the creation payload.
type CreatePaymentParams struct {
	PaymentType string `json:"paymentType" validate:"required"`
	Details     string `json:"details" validate:"required"`
	Notes       string `json:"notes,omitempty"`
}

// AddPayment handles the creation of a new payment record in the database.
func AddPayment(payload []byte) (*PaymentDetails, error) {
	logger.Log.Info("[create-paymentdetails-uc] Starting new payment record creation.")

	var p CreatePaymentParams
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[create-paymentdetails-uc] Unmarshal failed: %v", err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	newID := uuid.New().String()
	logger.Log.Info(fmt.Sprintf("[create-paymentdetails-uc] Generated PaymentID: %s for type: %s", newID, p.PaymentType))

	pt := &PaymentDetails{
		PaymentID:   newID,
		PaymentType: p.PaymentType,
		Details:     p.Details,
		Notes:       p.Notes,
	}

	const insertSQL = `
		INSERT INTO payment (payment_id, payment_type, details, notes) 
		VALUES ($1, $2, $3, $4)`

	logger.Log.Info(fmt.Sprintf("[create-paymentdetails-uc] Executing INSERT for PaymentID: %s", newID))

	_, err := db.DB.Exec(
		insertSQL,
		pt.PaymentID,
		pt.PaymentType,
		pt.Details,
		pt.Notes,
	)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[create-paymentdetails-uc] DB INSERT failed for %s: %v", newID, err))
		return nil, fmt.Errorf("failed to insert payment into database: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[create-paymentdetails-uc] Payment record %s created successfully.", newID))
	// Return the created PaymentDetails object
	return pt, nil
}
