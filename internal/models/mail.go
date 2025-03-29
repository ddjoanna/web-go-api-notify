package model

type Mail struct {
	SenderName    string `json:"sender_name" validate:"omitempty"`
	SenderAddress string `json:"sender_address" validate:"omitempty,email"`
	Subject       string `json:"subject" validate:"required"`
	Body          string `json:"body" validate:"required"`
}
