package mailer

import "context"

type MailProvider interface {
	SendEmail(ctx context.Context, request MailRequest) MailResponse
}

type SendgidStatus string

const (
	SendgidStatus_SENT   SendgidStatus = "sent"   // 已發送
	SendgidStatus_FAILED SendgidStatus = "failed" // 發送失敗
)

type MailReceiver struct {
	Email string `json:"email"`
}

type MailMessage struct {
	SenderName    string `json:"sender_name"`
	SenderAddress string `json:"sender_address"`
	Subject       string `json:"subject"`
	Body          string `json:"body"`
}

type MailRequest struct {
	Receivers []MailReceiver `json:"receivers"`
	Message   MailMessage    `json:"message"`
}

type MailResponse struct {
	Status           string `json:"status"`
	TraceId          string `json:"trace_id"`
	ProviderResponse string `json:"provider_response"`
}
