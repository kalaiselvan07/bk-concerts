package participant

type Participant struct {
	UserID   string `json:"userID" validate:"required"`
	Name     string `json:"name" validate:"required"`
	WaNum    string `json:"waNum" validate:"required"`
	Email    string `json:"email,omitempty"`
	Attended bool   `json:"attended" validate:"required"`
}
