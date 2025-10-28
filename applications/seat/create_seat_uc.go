package seat

import (
	"encoding/json"
	"fmt"

	"bk-concerts/db"

	"github.com/google/uuid"
)

type CreateSeatParams struct {
	SeatType  string  `json:"seatType" validate:"requried"`
	PriceGel  float32 `json:"priceGel" validate:"requried"`
	PriceInr  float32 `json:"priceInr" validate:"requried"`
	Available int     `json:"available"`
	Notes     string  `json:"notes,omitempty"`
}

func AddSeat(payload []byte) (*Seat, error) {
	var p *CreateSeatParams
	err := json.Unmarshal(payload, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	st := &Seat{
		SeatID:    uuid.New().String(),
		SeatType:  p.SeatType,
		PriceGel:  p.PriceGel,
		PriceInr:  p.PriceInr,
		Available: p.Available,
		Notes:     p.Notes,
	}

	const insertSQL = `
		INSERT INTO seat (seat_id, seat_type, price_gel, price_inr, available, notes) 
		VALUES ($1, $2, $3, $4, $5, $6)`

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
		return nil, fmt.Errorf("failed to insert seat into database: %w", err)
	}

	return st, nil
}
