package seat

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"

	"github.com/google/uuid"
)

type PartialUpdateSeatParams struct {
	SeatType  string  `json:"seatType,omitempty"`
	PriceGel  float64 `json:"priceGel,omitempty"`
	PriceInr  float64 `json:"priceInr,omitempty"`
	Available *int    `json:"available,omitempty"` // Use a pointer to distinguish 0 from 'not provided'
	Notes     string  `json:"notes,omitempty"`
}

// UpdateSeat performs a general update of seat details based on the payload.
// It assumes the full payload (minus SeatID) is the data to be potentially updated.
func UpdateSeat(seatID string, payload []byte) (*Seat, error) {
	// We use the simpler CreateSeatParams struct to bind (assuming all fields are sent,
	// or we use a separate struct for partial updates if fields can be omitted.)
	// Let's use the dynamic method for true partial updates.
	var p PartialUpdateSeatParams

	// Unmarshal into the PartialUpdate struct
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 1. Validate ID
	id, err := uuid.Parse(seatID)
	if err != nil {
		return nil, fmt.Errorf("invalid seat ID format: %w", err)
	}

	// 2. Build the dynamic SQL query
	sets := []string{}
	args := []interface{}{id} // Start with seat_id as the first argument ($1)
	argCounter := 2           // SQL placeholders start at $2 for the first update field

	// Check each optional field and build the SET clause dynamically
	if p.SeatType != "" {
		sets = append(sets, fmt.Sprintf("seat_type = $%d", argCounter))
		args = append(args, p.SeatType)
		argCounter++
	}
	if p.PriceGel != 0.0 {
		sets = append(sets, fmt.Sprintf("price_gel = $%d", argCounter))
		args = append(args, p.PriceGel)
		argCounter++
	}
	if p.PriceInr != 0.0 {
		sets = append(sets, fmt.Sprintf("price_inr = $%d", argCounter))
		args = append(args, p.PriceInr)
		argCounter++
	}
	// Check if 'Available' pointer is non-nil (meaning user provided the field)
	if p.Available != nil {
		sets = append(sets, fmt.Sprintf("available = $%d", argCounter))
		args = append(args, *p.Available) // Dereference the pointer
		argCounter++
	}
	// Check 'Notes' field
	if p.Notes != "" {
		sets = append(sets, fmt.Sprintf("notes = $%d", argCounter))
		args = append(args, p.Notes)
		argCounter++
	}

	// If no fields were provided, nothing to update.
	if len(sets) == 0 {
		// Fetch the current seat to return the existing details
		return GetSeat(seatID)
	}

	// 3. Construct the final SQL
	updateSQL := fmt.Sprintf(`
        UPDATE seat
        SET %s
        WHERE seat_id = $1
        RETURNING seat_id, seat_type, price_gel, price_inr, available, notes`,
		strings.Join(sets, ", "))

	// 4. Execute the update and scan the returned row
	st := &Seat{}
	row := db.DB.QueryRow(updateSQL, args...)

	if err := row.Scan(
		&st.SeatID, &st.SeatType, &st.PriceGel, &st.PriceInr, &st.Available, &st.Notes,
	); err != nil {
		// If it's a "no rows" error, the seat was not found
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("seat with ID %s not found", seatID)
		}
		return nil, fmt.Errorf("database update error: %w", err)
	}

	return st, nil
}

// UpdateAvailableTx updates the available seat count within a transaction.
func UpdateAvailableTx(tx *sql.Tx, seatID string, newAvailable int) (*Seat, error) {
	id, err := uuid.Parse(seatID)
	if err != nil {
		return nil, fmt.Errorf("invalid seat ID format: %w", err)
	}

	const updateSQL = `
		UPDATE seat
		SET available = $2
		WHERE seat_id = $1
		RETURNING seat_id, seat_type, price_gel, price_inr, available, notes`

	st := &Seat{}
	var seatIDUUID string

	row := tx.QueryRow(updateSQL, id, newAvailable) // ⬅️ Uses the transaction object (tx)

	if err := row.Scan(
		&seatIDUUID,
		&st.SeatType,
		&st.PriceGel,
		&st.PriceInr,
		&st.Available,
		&st.Notes,
	); err != nil {
		// No need to check for sql.ErrNoRows here as the prior locked row query handled existence
		return nil, fmt.Errorf("failed to scan updated seat row: %w", err)
	}
	st.SeatID = seatIDUUID // Assuming SeatID in the struct is uuid.UUID

	return st, nil
}
