package paymentdetails

type PaymentDetails struct {
	PaymentID   string `json:"paymentID" validate:"required"`
	PaymentType string `json:"paymentType" validate:"requried"`
	Details     string `json:"details" validate:"requried"`
	Notes       string `json:"notes,omitempty"`
}
