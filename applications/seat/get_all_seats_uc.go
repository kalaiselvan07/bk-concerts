package seat

import (
	"fmt"

	"bk-concerts/db" // ⬅️ Import your database package
)

// GetAllSeats retrieves a slice of all seat records from the database.
func GetAllSeats() ([]*Seat, error) {
	// 1. Define the SQL query
	// ORDER BY SeatType for consistent sorting is a good practice.
	const selectAllSQL = `
		SELECT seat_id, seat_type, price_gel, price_inr, available, notes
		FROM seat
		ORDER BY seat_type, price_inr`

	// 2. Execute the query
	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close() // Ensure the result set is closed when the function exits

	// 3. Initialize a slice to hold the retrieved seats
	seats := make([]*Seat, 0)

	// 4. Iterate through the result set
	for rows.Next() {
		// Initialize a new Seat struct for each row
		st := &Seat{}

		// Scan the column values into the struct fields
		err := rows.Scan(
			&st.SeatID,
			&st.SeatType,
			&st.PriceGel,
			&st.PriceInr,
			&st.Available,
			&st.Notes,
		)
		if err != nil {
			// Log and return the error if scanning fails
			return nil, fmt.Errorf("error scanning seat row: %w", err)
		}

		// Add the successfully scanned seat to the slice
		seats = append(seats, st)
	}

	// 5. Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// 6. Return the slice of seats
	return seats, nil
}
