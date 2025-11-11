package booking

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	BookingID        uuid.UUID `json:"bookingID"`
	BookingEmail     string    `json:"bookingEmail"`
	BookingStatus    string    `json:"bookingStatus"`
	PaymentDetailsID string    `json:"paymentDetailsID"` // Should be UUID in production
	ReceiptImage     []byte    `json:"receiptImage"`     // Stored as BYTEA/BLOB
	SeatQuantity     int       `json:"seatQuantity"`
	ConcertID        string    `json:"concertId"`
	SeatID           string    `json:"seatID"`
	SeatType         string    `json:"seatType"`
	TotalAmount      float64   `json:"totalAmount"`
	ParticipantIDs   []string  `json:"participantIDs"` // Stored as JSONB
	CreatedAt        time.Time `json:"createdAt"`
	UserNotes        string    `json:"userNotes"`
}

const (
	VERIFYING = "VERIFYING"
	CONFIRMED = "CONFIRMED"
	CANCELLED = "CANCELLED"
)
