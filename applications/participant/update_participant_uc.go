package participant

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bk-concerts/db"

	"github.com/google/uuid"
)

// UpdateParticipantParams defines fields that can be optionally updated.
// Attended is included as a pointer to allow setting it to false specifically.
type UpdateParticipantParams struct {
	Name     string `json:"name,omitempty"`
	WaNum    string `json:"wpNum,omitempty"`
	Email    string `json:"email,omitempty"`
	Attended *bool  `json:"attended,omitempty"` // Use pointer to differentiate false from omitted
}

// UpdateParticipant performs a general update of participant details.
func UpdateParticipant(userID string, payload []byte) (*Participant, error) {
	var p UpdateParticipantParams
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	id, err := uuid.Parse(userID)
	if err != nil {
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
		return GetParticipant(userID)
	}

	updateSQL := fmt.Sprintf(`
		UPDATE participant
		SET %s
		WHERE user_id = $1
		RETURNING user_id, name, wa_num, email, attended`,
		strings.Join(sets, ", "))

	pt := &Participant{}
	row := db.DB.QueryRow(updateSQL, args...)
	var userIDUUID uuid.UUID

	if err := row.Scan(
		&userIDUUID, &pt.Name, &pt.WaNum, &pt.Email, &pt.Attended,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("participant with ID %s not found", userID)
		}
		return nil, fmt.Errorf("database update error: %w", err)
	}
	pt.UserID = userIDUUID.String()

	return pt, nil
}
