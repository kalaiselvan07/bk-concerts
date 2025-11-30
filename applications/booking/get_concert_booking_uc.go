package booking

import (
	"database/sql"
	"fmt"

	"supra/concert/domain"
	"supra/db"
	"supra/logger"

	"github.com/google/uuid"
)

// GetConcert retrieves a single concert's details from the database by its ID.
func GetConcertBooking(concertID string) (bool, error) {
	logger.Log.Info(fmt.Sprintf("[get-concert-booking-uc] Starting retrieval for concert ID: %s", concertID))

	id, err := uuid.Parse(concertID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[get-concert-booking-uc] Retrieval failed for %s: Invalid UUID format.", concertID))
		return false, fmt.Errorf("invalid concert ID format: %w", err)
	}

	// 1. SELECT query includes payment_ids
	const selectSQL = `
		SELECT booking
		FROM concert
		WHERE concert_id = $1`

	logger.Log.Info(fmt.Sprintf("[get-concert-booking-uc] Executing SELECT query for ID: %s", concertID))
	row := db.DB.QueryRow(selectSQL, id)

	c := &domain.Concert{}

	// 2. Scan arguments include paymentIDsJSON
	err = row.Scan(
		&c.Booking,
	)

	if err != nil {
		// 3. Complete Error Handling
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[get-concert-booking-uc] Retrieval failed for %s: Not found in database.", concertID))
			return false, fmt.Errorf("concert with ID %s not found", concertID)
		}
		logger.Log.Error(fmt.Sprintf("[get-concert-booking-uc] Database query error for %s: %v", concertID, err))
		return false, fmt.Errorf("database query error: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[get-concert-booking-uc] Successfully retrieved concert: %s", concertID))
	return c.Booking, nil
}
