package smser

import "context"

type SmsProvider interface {
	SendBatchSms(ctx context.Context, request SmsBatchRequest) SmsBatchResponse
}

type MitakeStatus string

const (
	MitakeStatus_SENT   MitakeStatus = "sent"   // 已發送
	MitakeStatus_FAILED MitakeStatus = "failed" // 失敗
)

type SmsReceiver struct {
	TargetId string `json:"target_id"`
	Receiver string `json:"receiver"`
}

type SmsMessage struct {
	MessageId string `json:"target_id"`
	Message   string `json:"message"`
}

type SmsBatchRequest struct {
	Receivers []SmsReceiver `json:"receivers"`
	Message   SmsMessage    `json:"message"`
}

type SmsResponse struct {
	Status           string `json:"status"`
	TraceId          string `json:"trace_id"`
	ProviderResponse string `json:"provider_response"`
}
type SmsBatchResponse struct {
	Status      string        `json:"status"`
	MessageId   string        `json:"message_id"`
	SmsResponse []SmsResponse `json:"sms_response"`
}
