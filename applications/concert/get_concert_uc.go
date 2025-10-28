package concert

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bk-concerts/db" // Using the correct module path

	"github.com/google/uuid"
)

// GetConcert retrieves a single concert's details from the database by its ID.
func GetConcert(concertID string) (*Concert, error) {
	// 1. Validate and convert the ID string to uuid.UUID
	id, err := uuid.Parse(concertID)
	if err != nil {
		return nil, fmt.Errorf("invalid concert ID format: %w", err)
	}

	// 2. Define the SQL query
	const selectSQL = `
		SELECT concert_id, title, venue, timing, seat_ids, description
		FROM concert
		WHERE concert_id = $1`

	// 3. Execute the query
	row := db.DB.QueryRow(selectSQL, id)

	c := &Concert{}
	var seatIDsJSON []byte // Variable to temporarily hold the JSONB data

	// 4. Scan the row data into the struct fields
	err = row.Scan(
		&c.ConcertID,
		&c.Title,
		&c.Venue,
		&c.Timing,
		&seatIDsJSON, // Scan the JSONB data into a byte slice
		&c.Description,
	)

	// 5. Check the result of the scan
	if err != nil {
		if err == sql.ErrNoRows {
			// No concert was found
			return nil, fmt.Errorf("concert with ID %s not found", concertID)
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 6. Unmarshal the JSONB byte slice back into the []string field
	if len(seatIDsJSON) > 0 && string(seatIDsJSON) != "null" {
		if err := json.Unmarshal(seatIDsJSON, &c.SeatIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal seat IDs from database: %w", err)
		}
	}

	// 7. Return the retrieved concert details
	return c, nil
}
