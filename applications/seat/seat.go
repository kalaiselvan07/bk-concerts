package seat

type Seat struct {
	SeatID    string  `json:"seatID" validate:"required"`
	SeatType  string  `json:"seatType" validate:"requried"`
	PriceGel  float64 `json:"priceGel" validate:"requried"`
	PriceInr  float64 `json:"priceInr" validate:"requried"`
	Available int     `json:"available"`
	Notes     string  `json:"notes,omitempty"`
}
