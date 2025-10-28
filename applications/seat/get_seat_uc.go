package seat

import (
	"database/sql"
	"fmt"

	"bk-concerts/db" // ⬅️ Import your database package

	"github.com/google/uuid"
)

// GetSeat retrieves a single seat's details from the database by its ID.
func GetSeat(seatID string) (*Seat, error) {
	// 1. Validate and convert the seatID string to uuid.UUID
	id, err := uuid.Parse(seatID)
	if err != nil {
		return nil, fmt.Errorf("invalid seat ID format: %w", err)
	}

	// 2. Define the SQL query
	const selectSQL = `
		SELECT seat_id, seat_type, price_gel, price_inr, available, notes
		FROM seat
		WHERE seat_id = $1`

	// 3. Execute the query using QueryRow for a single result
	row := db.DB.QueryRow(selectSQL, id)

	// 4. Initialize the Seat struct to hold the result
	st := &Seat{}

	// 5. Scan the row data into the struct fields
	err = row.Scan(
		&st.SeatID,
		&st.SeatType,
		&st.PriceGel,
		&st.PriceInr,
		&st.Available,
		&st.Notes,
	)

	// 6. Check the result of the scan
	if err != nil {
		if err == sql.ErrNoRows {
			// This is a common and important check: no seat was found
			return nil, fmt.Errorf("seat with ID %s not found", seatID)
		}
		// Handle other potential database or scanning errors
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 7. Return the retrieved seat details
	return st, nil
}

// GetSeatForUpdateTx retrieves a seat and locks the row for the duration of the transaction.
func GetSeatForUpdateTx(tx *sql.Tx, seatID string) (*Seat, error) {
	id, err := uuid.Parse(seatID)
	if err != nil {
		return nil, fmt.Errorf("invalid seat ID format: %w", err)
	}

	// NOTE: Appending "FOR UPDATE" is crucial for preventing concurrent bookings
	// from reading the same 'Available' value before it's updated.
	const selectSQL = `
		SELECT seat_id, seat_type, price_gel, price_inr, available, notes
		FROM seat
		WHERE seat_id = $1 FOR UPDATE` // ⬅️ LOCKS THE ROW

	row := tx.QueryRow(selectSQL, id) // ⬅️ Uses the transaction object (tx)
	st := &Seat{}
	var seatIDUUID string

	err = row.Scan(
		&seatIDUUID,
		&st.SeatType,
		&st.PriceGel,
		&st.PriceInr,
		&st.Available,
		&st.Notes,
	)
	st.SeatID = seatIDUUID // Assuming SeatID in the struct is uuid.UUID

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("seat with ID %s not found for booking", seatID)
		}
		return nil, fmt.Errorf("transactional query error: %w", err)
	}
	return st, nil
}
