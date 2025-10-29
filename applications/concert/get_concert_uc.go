package concert

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bk-concerts/db"
	"bk-concerts/logger"

	"github.com/google/uuid"
)

// NOTE: Concert struct definition (including PaymentIDs) is assumed here

// GetConcert retrieves a single concert's details from the database by its ID.
func GetConcert(concertID string) (*Concert, error) {
	logger.Log.Info(fmt.Sprintf("[get-concert-uc] Starting retrieval for concert ID: %s", concertID))

	id, err := uuid.Parse(concertID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-concert-uc] Retrieval failed for %s: Invalid UUID format.", concertID))
		return nil, fmt.Errorf("invalid concert ID format: %w", err)
	}

	// 1. SELECT query includes payment_ids
	const selectSQL = `
		SELECT concert_id, title, venue, timing, seat_ids, payment_ids, description
		FROM concert
		WHERE concert_id = $1`

	logger.Log.Info(fmt.Sprintf("[get-concert-uc] Executing SELECT query for ID: %s", concertID))
	row := db.DB.QueryRow(selectSQL, id)

	c := &Concert{}
	var seatIDsJSON []byte
	var paymentIDsJSON []byte   // Variable for payment IDs JSONB
	var concertIDUUID uuid.UUID // Use UUID type for scanning

	// 2. Scan arguments include paymentIDsJSON
	err = row.Scan(
		&concertIDUUID,
		&c.Title,
		&c.Venue,
		&c.Timing,
		&seatIDsJSON,
		&paymentIDsJSON, // Scan the new column
		&c.Description,
	)

	if err != nil {
		// 3. Complete Error Handling
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[get-concert-uc] Retrieval failed for %s: Not found in database.", concertID))
			return nil, fmt.Errorf("concert with ID %s not found", concertID)
		}
		logger.Log.Error(fmt.Sprintf("[get-concert-uc] Database query error for %s: %v", concertID, err))
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// Assign the scanned UUID to the struct field
	c.ConcertID = concertIDUUID.String() // Or keep as UUID if struct uses uuid.UUID

	// 4. Unmarshal Seat IDs
	if len(seatIDsJSON) > 0 && string(seatIDsJSON) != "null" {
		if err := json.Unmarshal(seatIDsJSON, &c.SeatIDs); err != nil {
			logger.Log.Error(fmt.Sprintf("[get-concert-uc] Failed to unmarshal seat IDs for %s: %v", concertID, err))
			return nil, fmt.Errorf("failed to unmarshal seat IDs from database: %w", err)
		}
	}

	// 5. Unmarshal Payment IDs
	if len(paymentIDsJSON) > 0 && string(paymentIDsJSON) != "null" {
		if err := json.Unmarshal(paymentIDsJSON, &c.PaymentIDs); err != nil {
			logger.Log.Error(fmt.Sprintf("[get-concert-uc] Failed to unmarshal payment IDs for %s: %v", concertID, err))
			return nil, fmt.Errorf("failed to unmarshal payment IDs from database: %w", err)
		}
	}

	logger.Log.Info(fmt.Sprintf("[get-concert-uc] Successfully retrieved concert: %s", concertID))
	return c, nil
}
