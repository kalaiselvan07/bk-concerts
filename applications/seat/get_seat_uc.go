package seat

import (
	"database/sql"
	"fmt"

	"bk-concerts/db"     // ⬅️ Import your database package
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The Seat struct is assumed to be defined elsewhere in this package.

// GetSeat retrieves a single seat's details from the database by its ID.
func GetSeat(seatID string) (*Seat, error) {
	logger.Log.Info(fmt.Sprintf("[get-seat-uc] Starting GET read operation for ID: %s", seatID))

	// 1. Validate and convert the seatID string to uuid.UUID
	id, err := uuid.Parse(seatID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-seat-uc] Read failed for %s: Invalid UUID format.", seatID))
		return nil, fmt.Errorf("invalid seat ID format: %w", err)
	}

	// 2. Define the SQL query
	const selectSQL = `
		SELECT seat_id, seat_type, price_gel, price_inr, available, notes
		FROM seat
		WHERE seat_id = $1`

	// 3. Execute the query using QueryRow for a single result
	logger.Log.Info(fmt.Sprintf("[get-seat-uc] Executing standard read query for ID: %s", seatID))
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
			logger.Log.Warn(fmt.Sprintf("[get-seat-uc] Read failed for %s: Seat not found.", seatID))
			return nil, fmt.Errorf("seat with ID %s not found", seatID)
		}
		logger.Log.Error(fmt.Sprintf("[get-seat-uc] Database query error for %s: %v", seatID, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-seat-uc] Seat %s retrieved successfully. Available: %d", seatID, st.Available))
	// 7. Return the retrieved seat details
	return st, nil
}

// GetSeatForUpdateTx retrieves a seat and locks the row for the duration of the transaction.
func GetSeatForUpdateTx(tx *sql.Tx, seatID string) (*Seat, error) {
	logger.Log.Info(fmt.Sprintf("[get-seat-uc] Starting transactional lock (FOR UPDATE) for SeatID: %s", seatID))

	id, err := uuid.Parse(seatID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-seat-uc] Transactional read failed for %s: Invalid UUID format.", seatID))
		return nil, fmt.Errorf("invalid seat ID format: %w", err)
	}

	// NOTE: Appending "FOR UPDATE" is crucial for preventing concurrent bookings
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
			logger.Log.Warn(fmt.Sprintf("[get-seat-uc] Transactional lock failed for %s: Seat not found.", seatID))
			return nil, fmt.Errorf("seat with ID %s not found for booking", seatID)
		}
		logger.Log.Error(fmt.Sprintf("[get-seat-uc] Transactional query error for %s: %v", seatID, err))
		return nil, fmt.Errorf("transactional query error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-seat-uc] Seat %s successfully locked for update. Current Available: %d", seatID, st.Available))
	return st, nil
}
