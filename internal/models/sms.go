package model

import notifypb "proto/pkg/notify/v1/notify"

type Sms struct {
	Body string `json:"body" validate:"required"`
}

type MessageType string

const (
	MessageType_SMS  MessageType = "sms"
	MessageType_MAIL MessageType = "mail"
)

var (
	ConvertMessageTypeWithProto = map[notifypb.MessageType]MessageType{
		notifypb.MessageType_SMS:  MessageType_SMS,
		notifypb.MessageType_MAIL: MessageType_MAIL,
	}
)
