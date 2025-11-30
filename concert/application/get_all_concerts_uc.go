package application

import (
	"encoding/json"
	"fmt"
	"log/slog"

	// Added for rows.Err()
	"supra/concert/domain"
	"supra/db"     // Using the correct module path
	"supra/logger" // ⬅️ Assuming this import path
	// NOTE: Concert struct definition (including PaymentIDs) is assumed here
)

type GetAllConcertsUC struct {
	log *slog.Logger
}

func NewGetAllConcertsUC(log *slog.Logger) *GetAllConcertsUC {
	return &GetAllConcertsUC{
		log: log,
	}
}

// GetAllConcerts retrieves a slice of all concert records from the database.
func (uc *GetAllConcertsUC) Invoke() ([]*domain.Concert, error) {
	logger.Log.Info("[get-all-concert-uc] Starting retrieval of all concerts.")

	// ✨ 1. Update SELECT query to include payment_ids ✨
	const selectAllSQL = `
		SELECT concert_id, title, venue, timing, seat_ids, payment_ids, description, booking 
		FROM concert
		ORDER BY timing DESC`

	logger.Log.Info("[get-all-concert-uc] Executing SELECT all query.")
	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-concert-uc] Database query failed: %v", err))
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close() // Ensure the result set is closed

	concerts := make([]*domain.Concert, 0)
	recordCount := 0

	// 4. Iterate through the results
	for rows.Next() {
		c := &domain.Concert{}
		var seatIDsJSON []byte
		var paymentIDsJSON []byte // ✨ Variable for payment IDs JSONB ✨
		var concertIDStr string   // Helper if ConcertID is string in struct

		// ✨ 2. Update Scan arguments ✨
		err := rows.Scan(
			&concertIDStr, // Scan UUID to string or handle UUID type
			&c.Title,
			&c.Venue,
			&c.Timing,
			&seatIDsJSON,
			&paymentIDsJSON, // ✨ Scan the new column ✨
			&c.Description,
			&c.Booking,
		)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[get-all-concert-uc] Error scanning concert row: %v", err))
			return nil, fmt.Errorf("error scanning concert row: %w", err)
		}
		c.ConcertID = concertIDStr // Assign if ConcertID is string

		// 3. Unmarshal Seat IDs
		if len(seatIDsJSON) > 0 && string(seatIDsJSON) != "null" {
			if err := json.Unmarshal(seatIDsJSON, &c.SeatIDs); err != nil {
				logger.Log.Error(fmt.Sprintf("[get-all-concert-uc] Failed to unmarshal seat IDs for concert %s: %v", c.ConcertID, err))
				return nil, fmt.Errorf("failed to unmarshal seat IDs from database: %w", err)
			}
		}

		// ✨ 4. Unmarshal Payment IDs ✨
		if len(paymentIDsJSON) > 0 && string(paymentIDsJSON) != "null" {
			if err := json.Unmarshal(paymentIDsJSON, &c.PaymentIDs); err != nil {
				logger.Log.Error(fmt.Sprintf("[get-all-concert-uc] Failed to unmarshal payment IDs for concert %s: %v", c.ConcertID, err))
				return nil, fmt.Errorf("failed to unmarshal payment IDs from database: %w", err)
			}
		}

		concerts = append(concerts, c)
		recordCount++
	}

	// 5. Check for errors encountered during iteration
	if err = rows.Err(); err != nil {
		logger.Log.Error(fmt.Sprintf("[get-all-concert-uc] Error during row iteration: %v", err))
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-all-concert-uc] Successfully retrieved %d concerts.", recordCount))
	// 6. Return the slice of concerts
	return concerts, nil
}
