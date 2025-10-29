package concert

type Concert struct {
	ConcertID   string   `json:"concertID" validate:"required"`
	Title       string   `json:"title" validate:"required"`
	Venue       string   `json:"venue" validate:"required"`
	Timing      string   `json:"timing" validate:"required"`
	SeatIDs     []string `json:"seatIDs" validate:"required"`
	PaymentIDs  []string `json:"paymentDetailsIDs" validate:"required"`
	Description string   `json:"description,omitempty"`
}
