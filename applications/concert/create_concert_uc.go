package concert

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db"     // Using the correct module path
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The Concert struct (with PaymentIDs) is assumed to be defined elsewhere in this package.

type CreateConcertParams struct {
	Title             string   `json:"title" validate:"required"`
	Venue             string   `json:"venue" validate:"required"`
	Timing            string   `json:"timing" validate:"required"`
	SeatIDs           []string `json:"seatIDs"` // Incoming list of associated Seat IDs
	PaymentDetailsIDs []string `json:"paymentDetailsIDs,omitempty"`
	Description       string   `json:"description,omitempty"`
}

// CreateConcert handles the creation of a new concert record in the database.
// It accepts a raw JSON payload and returns the created Concert object.
func CreateConcert(payload []byte) (*Concert, error) {
	var p CreateConcertParams

	logger.Log.Info("[create-concert-uc] Starting concert creation process.")

	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[create-concert-uc] Failed to unmarshal payload: %v", err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	newID := uuid.New().String() // Consider using uuid.UUID type internally
	logger.Log.Info(fmt.Sprintf("[create-concert-uc] Generated new ConcertID: %s for title: %s", newID, p.Title))

	// 1. Prepare SeatIDs
	logger.Log.Info(fmt.Sprintf("[create-concert-uc] Marshaling %d SeatIDs for storage.", len(p.SeatIDs)))
	seatIDsJSON, err := json.Marshal(p.SeatIDs)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[create-concert-uc] Failed to marshal SeatIDs: %v", err))
		return nil, fmt.Errorf("failed to marshal SeatIDs to JSON: %w", err)
	}

	// ✨ 2. Prepare PaymentIDs ✨
	logger.Log.Info(fmt.Sprintf("[create-concert-uc] Marshaling %d PaymentIDs for storage.", len(p.PaymentDetailsIDs)))
	paymentIDsJSON, err := json.Marshal(p.PaymentDetailsIDs) // Marshal the new field
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[create-concert-uc] Failed to marshal PaymentIDs: %v", err))
		return nil, fmt.Errorf("failed to marshal PaymentIDs to JSON: %w", err)
	}

	concert := &Concert{
		ConcertID:   newID,
		Title:       p.Title,
		Venue:       p.Venue,
		Timing:      p.Timing,
		SeatIDs:     p.SeatIDs,
		PaymentIDs:  p.PaymentDetailsIDs, // ✨ Store the Go slice in the struct ✨
		Description: p.Description,
	}

	// ✨ 3. Update SQL INSERT statement ✨
	const insertSQL = `
		INSERT INTO concert (concert_id, title, venue, timing, seat_ids, payment_ids, description) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)` // Now 7 placeholders

	// ✨ 4. Update Execute arguments ✨
	logger.Log.Info(fmt.Sprintf("[create-concert-uc] Inserting new concert record into database for ID: %s", newID))
	_, err = db.DB.Exec(
		insertSQL,
		concert.ConcertID,
		concert.Title,
		concert.Venue,
		concert.Timing,
		seatIDsJSON,    // Storing seat IDs JSON
		paymentIDsJSON, // ✨ Storing payment IDs JSON ✨
		concert.Description,
	)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[create-concert-uc] Failed to insert concert %s into database: %v", newID, err))
		return nil, fmt.Errorf("failed to insert concert into database: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[create-concert-uc] Concert %s created successfully. Venue: %s", newID, p.Venue))
	// 5. Return the created Concert object
	return concert, nil
}
