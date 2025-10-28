package concert

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db" // Using the correct module path

	"github.com/google/uuid"
)

type CreateConcertParams struct {
	Title       string   `json:"title" validate:"required"`
	Venue       string   `json:"venue" validate:"required"`
	Timing      string   `json:"timing" validate:"required"`
	SeatIDs     []string `json:"seatIDs"` // Incoming list of associated Seat IDs
	Description string   `json:"description,omitempty"`
}

// The structs defined in concert/concert.go are implicitly imported here.

// CreateConcert handles the creation of a new concert record in the database.
// It accepts a raw JSON payload and returns the created Concert object.
func CreateConcert(payload []byte) (*Concert, error) {
	var p CreateConcertParams
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	newID := uuid.New().String()

	// 1. Prepare the SeatIDs for DB storage (JSON encoding)
	seatIDsJSON, err := json.Marshal(p.SeatIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SeatIDs to JSON: %w", err)
	}

	concert := &Concert{
		ConcertID:   newID,
		Title:       p.Title,
		Venue:       p.Venue,
		Timing:      p.Timing,
		SeatIDs:     p.SeatIDs,
		Description: p.Description,
	}

	// 2. Define the SQL INSERT statement
	// We assume you will create a 'concert' table that supports UUID and TEXT fields.
	// The 'seat_ids' column should be of type TEXT or JSONB in Postgres.
	const insertSQL = `
		INSERT INTO concert (concert_id, title, venue, timing, seat_ids, description) 
		VALUES ($1, $2, $3, $4, $5, $6)`

	// 3. Execute the SQL command
	_, err = db.DB.Exec(
		insertSQL,
		concert.ConcertID,
		concert.Title,
		concert.Venue,
		concert.Timing,
		seatIDsJSON, // Storing the JSON byte array
		concert.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert concert into database: %w", err)
	}

	// 4. Return the created Concert object
	return concert, nil
}
