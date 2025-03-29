package model

import (
	notifypb "proto/pkg/notify/v1/notify"
	"time"
)

type SendSmsRequest struct {
	Sms
	Receivers   []string   `json:"receivers" validate:"required,min=1,max=1000,dive,regexp=09[0-9]{8}"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

type SendMailRequest struct {
	Mail
	Receivers   []string   `json:"receivers" validate:"required,min=1,max=1000,dive,email"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

type CancelScheduledByMessageIdRequest struct {
	MessageId string `json:"message_id" validate:"required"`
}

type ListStatusWithPagingRequest struct {
	MessageType string                `json:"message_type" validate:"required,oneof=sms mail"`
	MessageId   string                `json:"message_id" validate:"omitempty,required_without=receiver"`
	Receiver    string                `json:"receiver" validate:"omitempty,required_without=message_id"`
	Page        *notifypb.PageRequest `json:"page" validate:"required"`
	StartAt     *time.Time            `json:"start_at" validate:"required"`
	EndAt       *time.Time            `json:"end_at" validate:"required"`
}

type PageRequest struct {
	Index     int     `json:"index" validate:"required,gte=1"`                // 頁碼，從 1 開始
	Size      int     `json:"size" validate:"required,gte=1"`                 // 頁面大小
	SortField *string `json:"sort_field" validate:"omitempty"`                // 排序字段
	SortOrder *string `json:"sort_order" validate:"omitempty,oneof=asc desc"` // 排序方式 asc 或 desc
}
