package participant

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bk-concerts/db" // Using the correct module path

	"github.com/google/uuid"
)

// CreateParticipantParams is used for the creation payload.
type CreateParticipantParams struct {
	Name  string `json:"name" validate:"required"`
	WaNum string `json:"wpNum" validate:"required"`
	Email string `json:"email,omitempty"`
}

// AddParticipant handles the creation of a new participant record in the database.
func AddParticipant(payload []byte) (*Participant, error) {
	var p CreateParticipantParams
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	pt := &Participant{
		UserID:   uuid.New().String(),
		Name:     p.Name,
		WaNum:    p.WaNum,
		Email:    p.Email,
		Attended: false, // Default to false upon creation
	}

	const insertSQL = `
		INSERT INTO participant (user_id, name, wa_num, email, attended) 
		VALUES ($1, $2, $3, $4, $5)`

	_, err := db.DB.Exec(
		insertSQL,
		pt.UserID,
		pt.Name,
		pt.WaNum,
		pt.Email,
		pt.Attended,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert participant into database: %w", err)
	}

	// Return the created Participant object
	return pt, nil
}

// AddParticipantTx creates a new participant record within a transaction.
// It is used by the booking process.
func AddParticipantTx(tx *sql.Tx, payload []byte) (*Participant, error) {
	var p CreateParticipantParams
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal participant payload: %w", err)
	}

	pt := &Participant{
		UserID:   uuid.New().String(), // Assuming UserID is uuid.UUID here
		Name:     p.Name,
		WaNum:    p.WaNum,
		Email:    p.Email,
		Attended: false,
	}

	const insertSQL = `
		INSERT INTO participant (user_id, name, wa_num, email, attended) 
		VALUES ($1, $2, $3, $4, $5)`

	// Use tx.Exec() to run the command within the ongoing transaction
	_, err := tx.Exec(
		insertSQL,
		pt.UserID,
		pt.Name,
		pt.WaNum,
		pt.Email,
		pt.Attended,
	)
	if err != nil {
		return nil, fmt.Errorf("transactional insert failed: %w", err)
	}

	// Note: We return the Participant object with the new UUID (pt.UserID)
	// but the ID in the struct is still the UUID type.

	return pt, nil
}
