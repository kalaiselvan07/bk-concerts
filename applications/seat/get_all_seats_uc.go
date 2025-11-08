package seat

import (
	"fmt"
	// Required for rows.Err() and implicit type handling
	"supra/db"     // ⬅️ Import your database package
	"supra/logger" // ⬅️ Assuming this import path
)

// NOTE: The Seat struct is assumed to be defined elsewhere in this package.

// GetAllSeats retrieves a slice of all seat records from the database.
func GetAllSeats() ([]*Seat, error) {
	logger.Log.Info("[get-all-seat-uc] Starting retrieval of all seat records.")

	// 1. Define the SQL query
	const selectAllSQL = `
		SELECT seat_id, seat_type, price_gel, price_inr, available, notes
		FROM seat
		ORDER BY seat_type, price_inr`

	// 2. Execute the query
	logger.Log.Info("[get-all-seat-uc] Executing SELECT all query.")
	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-seat-uc] Database query failed: %v", err))
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close() // Ensure the result set is closed when the function exits

	// 3. Initialize a slice to hold the retrieved seats
	seats := make([]*Seat, 0)
	recordCount := 0

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
			logger.Log.Error(fmt.Sprintf("[get-all-seat-uc] Error scanning individual seat row: %v", err))
			return nil, fmt.Errorf("error scanning seat row: %w", err)
		}

		// Add the successfully scanned seat to the slice
		seats = append(seats, st)
		recordCount++
	}

	// 5. Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-seat-uc] Error encountered during final row iteration: %v", err))
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-seat-uc] Successfully retrieved %d seat records.", recordCount))
	// 6. Return the slice of seats
	return seats, nil
}
