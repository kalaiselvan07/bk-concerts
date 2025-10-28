package concert

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db" // Using the correct module path
)

// GetAllConcerts retrieves a slice of all concert records from the database.
func GetAllConcerts() ([]*Concert, error) {
	// 1. Define the SQL query
	const selectAllSQL = `
		SELECT concert_id, title, venue, timing, seat_ids, description
		FROM concert
		ORDER BY timing DESC` // Order by timing descending to show latest first

	// 2. Execute the query
	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close() // Ensure the result set is closed

	// 3. Initialize the slice
	concerts := make([]*Concert, 0)

	// 4. Iterate through the results
	for rows.Next() {
		c := &Concert{}
		var seatIDsJSON []byte // Variable to temporarily hold the JSONB data

		// Scan the row data
		err := rows.Scan(
			&c.ConcertID,
			&c.Title,
			&c.Venue,
			&c.Timing,
			&seatIDsJSON, // Scan the JSONB data
			&c.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning concert row: %w", err)
		}

		// Unmarshal the JSONB byte slice back into the []string field
		if len(seatIDsJSON) > 0 && string(seatIDsJSON) != "null" {
			if err := json.Unmarshal(seatIDsJSON, &c.SeatIDs); err != nil {
				return nil, fmt.Errorf("failed to unmarshal seat IDs from database: %w", err)
			}
		}

		// Add the concert to the slice
		concerts = append(concerts, c)
	}

	// 5. Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// 6. Return the slice of concerts
	return concerts, nil
}
