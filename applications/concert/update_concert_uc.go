package concert

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db" // Using the correct module path

	"github.com/google/uuid"
)

// PartialUpdateConcertParams defines fields that can be optionally updated.
// Using pointers or checking for zero/empty values helps with partial updates.
type PartialUpdateConcertParams struct {
	Title  string `json:"title,omitempty"`
	Venue  string `json:"venue,omitempty"`
	Timing string `json:"timing,omitempty"`
	// Using a slice here means if included, it fully replaces the old SeatIDs.
	SeatIDs     []string `json:"seatIDs,omitempty"`
	Description string   `json:"description,omitempty"`
}

// UpdateConcert performs a general update of concert details.
func UpdateConcert(concertID string, payload []byte) (*Concert, error) {
	var p PartialUpdateConcertParams

	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 1. Validate ID
	id, err := uuid.Parse(concertID)
	if err != nil {
		return nil, fmt.Errorf("invalid concert ID format: %w", err)
	}

	// 2. Prepare SeatIDs for DB (only if provided in payload)
	var seatIDsJSON []byte
	if p.SeatIDs != nil { // Check if the slice was included in the JSON payload
		seatIDsJSON, err = json.Marshal(p.SeatIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal SeatIDs to JSON: %w", err)
		}
	}

	// 3. Build the dynamic SQL query
	sets := []string{}
	args := []interface{}{id} // Start with concert_id as the first argument ($1)
	argCounter := 2           // SQL placeholders start at $2 for the first update field

	if p.Title != "" {
		sets = append(sets, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, p.Title)
		argCounter++
	}
	if p.Venue != "" {
		sets = append(sets, fmt.Sprintf("venue = $%d", argCounter))
		args = append(args, p.Venue)
		argCounter++
	}
	if p.Timing != "" {
		sets = append(sets, fmt.Sprintf("timing = $%d", argCounter))
		args = append(args, p.Timing)
		argCounter++
	}
	if p.SeatIDs != nil {
		sets = append(sets, fmt.Sprintf("seat_ids = $%d", argCounter))
		args = append(args, seatIDsJSON)
		argCounter++
	}
	if p.Description != "" {
		sets = append(sets, fmt.Sprintf("description = $%d", argCounter))
		args = append(args, p.Description)
		argCounter++
	}

	if len(sets) == 0 {
		// No fields to update, return the current details
		return GetConcert(concertID)
	}

	// 4. Construct the final SQL
	updateSQL := fmt.Sprintf(`
		UPDATE concert
		SET %s
		WHERE concert_id = $1
		RETURNING concert_id, title, venue, timing, seat_ids, description`,
		strings.Join(sets, ", "))

	// 5. Execute and scan the returned row
	c := &Concert{}
	row := db.DB.QueryRow(updateSQL, args...)
	var returnedSeatIDsJSON []byte // To store the returned JSONB

	if err := row.Scan(
		&c.ConcertID, &c.Title, &c.Venue, &c.Timing, &returnedSeatIDsJSON, &c.Description,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("concert with ID %s not found", concertID)
		}
		return nil, fmt.Errorf("database update error: %w", err)
	}

	// Unmarshal the returned SeatIDs
	if len(returnedSeatIDsJSON) > 0 && string(returnedSeatIDsJSON) != "null" {
		if err := json.Unmarshal(returnedSeatIDsJSON, &c.SeatIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal returned seat IDs: %w", err)
		}
	}

	return c, nil
}
