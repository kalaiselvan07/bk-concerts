package participant

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bk-concerts/db"     // Using the correct module path
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The Participant struct is assumed to be defined elsewhere in this package.

// CreateParticipantParams is used for the creation payload.
type CreateParticipantParams struct {
	Name  string `json:"name" validate:"required"`
	WaNum string `json:"waNum" validate:"required"`
	Email string `json:"email,omitempty"`
}

// AddParticipant handles the creation of a new participant record in the database.
func AddParticipant(payload []byte) (*Participant, error) {
	logger.Log.Info("[create-participant-uc] Starting standard participant creation.")

	var p CreateParticipantParams
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[create-participant-uc] Unmarshal failed: %v", err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	newID := uuid.New().String()
	logger.Log.Info(fmt.Sprintf("[create-participant-uc] Generated UserID: %s for Name: %s", newID, p.Name))

	pt := &Participant{
		UserID:   newID,
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
		logger.Log.Error(fmt.Sprintf("[create-participant-uc] DB INSERT failed for %s: %v", newID, err))
		return nil, fmt.Errorf("failed to insert participant into database: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[create-participant-uc] Participant %s created successfully.", newID))
	// Return the created Participant object
	return pt, nil
}

// AddParticipantTx creates a new participant record within a transaction.
// It is used by the booking process.
func AddParticipantTx(tx *sql.Tx, payload []byte) (*Participant, error) {
	logger.Log.Info("[create-participant-uc] Starting transactional participant creation.")

	var p CreateParticipantParams
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Log.Error(fmt.Sprintf("[create-participant-uc] Transactional unmarshal failed: %v", err))
		return nil, fmt.Errorf("failed to unmarshal participant payload: %w", err)
	}

	newID := uuid.New().String()
	logger.Log.Info(fmt.Sprintf("[create-participant-uc] Generated Transactional UserID: %s for Name: %s", newID, p.Name))

	pt := &Participant{
		UserID:   newID,
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
		logger.Log.Error(fmt.Sprintf("[create-participant-uc] Transactional DB INSERT failed for %s: %v", newID, err))
		return nil, fmt.Errorf("transactional insert failed: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[create-participant-uc] Participant %s inserted successfully within transaction.", newID))
	// Note: We return the Participant object with the new UUID (pt.UserID)
	return pt, nil
}
