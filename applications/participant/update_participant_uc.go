package participant

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"     // Using the correct module path
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// UpdateParticipantParams defines fields that can be optionally updated.
type UpdateParticipantParams struct {
	Name     string `json:"name,omitempty"`
	WaNum    string `json:"wpNum,omitempty"`
	Email    string `json:"email,omitempty"`
	Attended *bool  `json:"attended,omitempty"` // Use pointer to differentiate false from omitted
}

// NOTE: The Participant struct and GetParticipant function are assumed to be defined elsewhere in this package.

// UpdateParticipant performs a general update of participant details.
func UpdateParticipant(userID string, payload []byte) (*Participant, error) {
	logger.Log.Info(fmt.Sprintf("[update-participant-uc] Starting update for UserID: %s", userID))

	var p UpdateParticipantParams
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[update-participant-uc] Unmarshal failed for %s: %v", userID, err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("[update-participant-uc] Update failed for %s: Invalid UUID format.", userID))
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	sets := []string{}
	args := []interface{}{id}
	argCounter := 2

	if p.Name != "" {
		sets = append(sets, fmt.Sprintf("name = $%d", argCounter))
		args = append(args, p.Name)
		argCounter++
	}
	if p.WaNum != "" {
		sets = append(sets, fmt.Sprintf("wa_num = $%d", argCounter))
		args = append(args, p.WaNum)
		argCounter++
	}
	if p.Email != "" {
		sets = append(sets, fmt.Sprintf("email = $%d", argCounter))
		args = append(args, p.Email)
		argCounter++
	}
	if p.Attended != nil {
		sets = append(sets, fmt.Sprintf("attended = $%d", argCounter))
		args = append(args, *p.Attended)
		argCounter++
	}

	if len(sets) == 0 {
		logger.Log.Warn(fmt.Sprintf("[update-participant-uc] Update skipped for %s: No fields provided.", userID))
		return GetParticipant(userID)
	}

	updateSQL := fmt.Sprintf(`
		UPDATE participant
		SET %s
		WHERE user_id = $1
		RETURNING user_id, name, wa_num, email, attended`,
		strings.Join(sets, ", "))

	logger.Log.Info(fmt.Sprintf("[update-participant-uc] Executing UPDATE for %s with %d fields modified.", userID, len(sets)))

	pt := &Participant{}
	row := db.DB.QueryRow(updateSQL, args...)
	var userIDUUID uuid.UUID

	if err := row.Scan(
		&userIDUUID, &pt.Name, &pt.WaNum, &pt.Email, &pt.Attended,
	); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Warn(fmt.Sprintf("[update-participant-uc] Update failed for %s: Participant not found.", userID))
			return nil, fmt.Errorf("participant with ID %s not found", userID)
		}
		logger.Log.Error(fmt.Sprintf("[update-participant-uc] Database update error for %s: %v", userID, err))
		return nil, fmt.Errorf("database update error: %w", err)
	}
	pt.UserID = userIDUUID.String()

	logger.Log.Info(fmt.Sprintf("[update-participant-uc] Participant %s updated successfully. Name: %s", userID, pt.Name))
	return pt, nil
}
