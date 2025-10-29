package concert

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"     // Using the correct module path
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The Concert struct (including PaymentIDs) is assumed here
// NOTE: The GetConcert function is assumed here

// PartialUpdateConcertParams defines fields that can be optionally updated.
type PartialUpdateConcertParams struct {
	Title   string   `json:"title,omitempty"`
	Venue   string   `json:"venue,omitempty"`
	Timing  string   `json:"timing,omitempty"`
	SeatIDs []string `json:"seatIDs,omitempty"`
	// ✨ ADDED: Field to receive payment IDs from the frontend ✨
	PaymentDetailsIDs []string `json:"paymentDetailsIDs,omitempty"`
	Description       string   `json:"description,omitempty"`
}

// UpdateConcert performs a general update of concert details.
func UpdateConcert(concertID string, payload []byte) (*Concert, error) {
	logger.Log.Info(fmt.Sprintf("[update-concert-uc] Starting update for concert ID: %s", concertID))

	var p PartialUpdateConcertParams

	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-concert-uc] Failed to unmarshal update payload for %s: %v", concertID, err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// 1. Validate ID
	id, err := uuid.Parse(concertID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-concert-uc] Update failed for %s: Invalid UUID format.", concertID))
		return nil, fmt.Errorf("invalid concert ID format: %w", err)
	}

	// 2. Prepare SeatIDs (if provided)
	var seatIDsJSON []byte
	if p.SeatIDs != nil {
		logger.Log.Info(fmt.Sprintf("[update-concert-uc] New SeatIDs provided (%d). Marshaling.", len(p.SeatIDs)))
		seatIDsJSON, err = json.Marshal(p.SeatIDs)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[update-concert-uc] Failed to marshal new SeatIDs for %s: %v", concertID, err))
			return nil, fmt.Errorf("failed to marshal SeatIDs to JSON: %w", err)
		}
	}

	// ✨ 3. Prepare PaymentIDs (if provided) ✨
	var paymentIDsJSON []byte
	if p.PaymentDetailsIDs != nil {
		logger.Log.Info(fmt.Sprintf("[update-concert-uc] New PaymentIDs provided (%d). Marshaling.", len(p.PaymentDetailsIDs)))
		paymentIDsJSON, err = json.Marshal(p.PaymentDetailsIDs)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("[update-concert-uc] Failed to marshal new PaymentIDs for %s: %v", concertID, err))
			return nil, fmt.Errorf("failed to marshal PaymentIDs to JSON: %w", err)
		}
	}

	// 4. Build the dynamic SQL query
	sets := []string{}
	args := []interface{}{id} // Start with concert_id as $1
	argCounter := 2

	if p.Title != "" {
		sets = append(sets, fmt.Sprintf("title = $%d", argCounter))
		args = append(args, p.Title)
		argCounter++
	}
	if p.Venue != "" {
		sets = append(sets, fmt.Sprintf("venue = $%d", argCounter))
		args = append(args, p.Venue)
		argCounter++
	}
	if p.Timing != "" {
		sets = append(sets, fmt.Sprintf("timing = $%d", argCounter))
		args = append(args, p.Timing)
		argCounter++
	}
	if p.SeatIDs != nil {
		sets = append(sets, fmt.Sprintf("seat_ids = $%d", argCounter))
		args = append(args, seatIDsJSON)
		argCounter++
	}
	// ✨ Append PaymentIDs if provided ✨
	if p.PaymentDetailsIDs != nil {
		sets = append(sets, fmt.Sprintf("payment_ids = $%d", argCounter))
		args = append(args, paymentIDsJSON)
		argCounter++
	}
	if p.Description != "" {
		sets = append(sets, fmt.Sprintf("description = $%d", argCounter))
		args = append(args, p.Description)
		argCounter++
	}

	if len(sets) == 0 {
		logger.Log.Warn(fmt.Sprintf("[update-concert-uc] Update skipped for %s: No updatable fields.", concertID))
		return GetConcert(concertID) // Return current details
	}

	// ✨ 5. Update Construct final SQL (include payment_ids in RETURNING) ✨
	updateSQL := fmt.Sprintf(`
		UPDATE concert
		SET %s
		WHERE concert_id = $1
		RETURNING concert_id, title, venue, timing, seat_ids, payment_ids, description`,
		strings.Join(sets, ", "))

	logger.Log.Info(fmt.Sprintf("[update-concert-uc] Executing UPDATE for %s with %d fields.", concertID, len(sets)))

	// 6. Execute and scan the returned row
	c := &Concert{}
	row := db.DB.QueryRow(updateSQL, args...)
	var returnedSeatIDsJSON []byte
	var returnedPaymentIDsJSON []byte // ✨ Variable for returned payment IDs ✨
	var concertIDUUID uuid.UUID

	// ✨ 7. Update Scan arguments ✨
	if err := row.Scan(
		&concertIDUUID, &c.Title, &c.Venue, &c.Timing,
		&returnedSeatIDsJSON,
		&returnedPaymentIDsJSON, // ✨ Scan the returned column ✨
		&c.Description,
	); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[update-concert-uc] Update failed for %s: Not found.", concertID))
			return nil, fmt.Errorf("concert with ID %s not found", concertID)
		}
		logger.Log.Error(fmt.Sprintf("[update-concert-uc] Database update error for %s: %v", concertID, err))
		return nil, fmt.Errorf("database update error: %w", err)
	}
	c.ConcertID = concertIDUUID.String() // Assign if needed

	// Unmarshal Seat IDs
	if len(returnedSeatIDsJSON) > 0 && string(returnedSeatIDsJSON) != "null" {
		if err := json.Unmarshal(returnedSeatIDsJSON, &c.SeatIDs); err != nil {
			logger.Log.Error(fmt.Sprintf("[update-concert-uc] Unmarshal SeatIDs failed for %s: %v", concertID, err))
			return nil, fmt.Errorf("failed to unmarshal returned seat IDs: %w", err)
		}
	}

	// ✨ 8. Unmarshal Payment IDs ✨
	if len(returnedPaymentIDsJSON) > 0 && string(returnedPaymentIDsJSON) != "null" {
		if err := json.Unmarshal(returnedPaymentIDsJSON, &c.PaymentIDs); err != nil {
			logger.Log.Error(fmt.Sprintf("[update-concert-uc] Unmarshal PaymentIDs failed for %s: %v", concertID, err))
			return nil, fmt.Errorf("failed to unmarshal returned payment IDs: %w", err)
		}
	}

	logger.Log.Info(fmt.Sprintf("[update-concert-uc] Concert %s updated successfully.", concertID))
	return c, nil
}
