package seat

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db"     // Using the correct module path
	"bk-concerts/logger" // ⬅️ Assuming this import path

	"github.com/google/uuid"
)

// NOTE: The Seat struct is assumed to be defined elsewhere in this package.

type CreateSeatParams struct {
	SeatType  string  `json:"seatType" validate:"requried"`
	PriceGel  float64 `json:"priceGel" validate:"requried"`
	PriceInr  float64 `json:"priceInr" validate:"requried"`
	Available int     `json:"available"`
	Notes     string  `json:"notes,omitempty"`
}

func AddSeat(payload []byte) (*Seat, error) {
	var p *CreateSeatParams

	logger.Log.Info("[create-seat-uc] Starting seat creation process.")

	err := json.Unmarshal(payload, &p)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[create-seat-uc] Failed to unmarshal payload: %v", err))
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	newID := uuid.New().String()
	logger.Log.Info(fmt.Sprintf("[create-seat-uc] Generated new SeatID: %s for type: %s", newID, p.SeatType))

	st := &Seat{
		SeatID:    newID,
		SeatType:  p.SeatType,
		PriceGel:  p.PriceGel,
		PriceInr:  p.PriceInr,
		Available: p.Available,
		Notes:     p.Notes,
	}

	const insertSQL = `
		INSERT INTO seat (seat_id, seat_type, price_gel, price_inr, available, notes) 
		VALUES ($1, $2, $3, $4, $5, $6)`

	logger.Log.Info(fmt.Sprintf("[create-seat-uc] Executing INSERT for SeatID: %s, Price: %.2f INR", newID, p.PriceInr))

	_, err = db.DB.Exec(
		insertSQL,
		st.SeatID,
		st.SeatType,
		st.PriceGel,
		st.PriceInr,
		st.Available,
		st.Notes,
	)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("[create-seat-uc] Failed to insert seat %s into database: %v", newID, err))
		return nil, fmt.Errorf("failed to insert seat into database: %w", err)
	}

	logger.Log.Info(fmt.Sprintf("[create-seat-uc] Seat %s created successfully. Available count: %d", newID, st.Available))
	return st, nil
}
