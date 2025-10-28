package participant

import (
	"bk-concerts/db"
	"fmt"

	"github.com/google/uuid"
)

// GetAllParticipants retrieves a slice of all participant records.
func GetAllParticipants() ([]*Participant, error) {
	const selectAllSQL = `
		SELECT user_id, name, wa_num, email, attended
		FROM participant
		ORDER BY name`

	rows, err := db.DB.Query(selectAllSQL)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	participants := make([]*Participant, 0)

	for rows.Next() {
		p := &Participant{}
		var userIDUUID uuid.UUID

		err := rows.Scan(
			&userIDUUID,
			&p.Name,
			&p.WaNum,
			&p.Email,
			&p.Attended,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning participant row: %w", err)
		}
		p.UserID = userIDUUID.String()
		participants = append(participants, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}
	return participants, nil
}
