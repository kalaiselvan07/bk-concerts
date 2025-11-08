package seat

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"supra/db"     // Using the correct module path
	"supra/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

type PartialUpdateSeatParams struct {
	SeatType  string  `json:"seatType,omitempty"`
	PriceGel  float64 `json:"priceGel,omitempty"`
	PriceInr  float64 `json:"priceInr,omitempty"`
	Available *int    `json:"available,omitempty"` // Use a pointer to distinguish 0 from 'not provided'
	Notes     string  `json:"notes,omitempty"`
}

// NOTE: The Seat struct and GetSeat function are assumed to be defined elsewhere in this package.

// UpdateSeat performs a general update of seat details based on the payload.
func UpdateSeat(seatID string, payload []byte) (*Seat, error) {
	logger.Log.Info(fmt.Sprintf("[update-seat-uc] Starting general update for SeatID: %s", seatID))

	var p PartialUpdateSeatParams

	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-seat-uc] Failed to unmarshal update payload for %s: %v", seatID, err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 1. Validate ID
	id, err := uuid.Parse(seatID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-seat-uc] Update failed for %s: Invalid UUID format.", seatID))
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
	if p.Available != nil {
		sets = append(sets, fmt.Sprintf("available = $%d", argCounter))
		args = append(args, *p.Available) // Dereference the pointer
		argCounter++
	}
	if p.Notes != "" {
		sets = append(sets, fmt.Sprintf("notes = $%d", argCounter))
		args = append(args, p.Notes)
		argCounter++
	}

	// If no fields were provided, nothing to update.
	if len(sets) == 0 {
		logger.Log.Warn(fmt.Sprintf("[update-seat-uc] Update skipped for %s: No fields provided in payload.", seatID))
		return GetSeat(seatID)
	}

	// 3. Construct the final SQL
	updateSQL := fmt.Sprintf(`
		UPDATE seat
		SET %s
		WHERE seat_id = $1
		RETURNING seat_id, seat_type, price_gel, price_inr, available, notes`,
		strings.Join(sets, ", "))

	logger.Log.Info(fmt.Sprintf("[update-seat-uc] Executing general UPDATE for %s with %d fields modified.", seatID, len(sets)))

	// 4. Execute the update and scan the returned row
	st := &Seat{}
	row := db.DB.QueryRow(updateSQL, args...)

	if err := row.Scan(
		&st.SeatID, &st.SeatType, &st.PriceGel, &st.PriceInr, &st.Available, &st.Notes,
	); err != nil {
		// If it's a "no rows" error, the seat was not found
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[update-seat-uc] General update failed for %s: Seat not found.", seatID))
			return nil, fmt.Errorf("seat with ID %s not found", seatID)
		}
		logger.Log.Error(fmt.Sprintf("[update-seat-uc] Database update error for %s: %v", seatID, err))
		return nil, fmt.Errorf("database update error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[update-seat-uc] Seat %s updated successfully.", seatID))
	return st, nil
}

// UpdateAvailableTx updates the available seat count within a transaction.
func UpdateAvailableTx(tx *sql.Tx, seatID string, newAvailable int) (*Seat, error) {
	logger.Log.Info(fmt.Sprintf("[update-seat-uc] Starting transactional availability update for ID: %s. New count: %d", seatID, newAvailable))

	id, err := uuid.Parse(seatID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-seat-uc] Transactional update failed for %s: Invalid UUID format.", seatID))
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
		logger.Log.Error(fmt.Sprintf("[update-seat-uc] Failed to scan updated row during transaction for %s: %v", seatID, err))
		return nil, fmt.Errorf("failed to scan updated seat row: %w", err)
	}
	st.SeatID = seatIDUUID // Assuming SeatID in the struct is uuid.UUID

	logger.Log.Info(fmt.Sprintf("[update-seat-uc] Seat %s availability updated to %d within transaction.", seatID, newAvailable))
	return st, nil
}
