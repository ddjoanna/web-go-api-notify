package entity

import (
	"time"

	"gorm.io/gorm"
)

type Provider string

const (
	Provider_SENDGRID Provider = "sendgrid" // SendGrid 發送商
	Provider_MITAKE   Provider = "mitake"   // Mitake 發送商
)

type TargetStatus string

const (
	TargetStatus_PENDING  TargetStatus = "pending"  // 待處理
	TargetStatus_ENQUEUED TargetStatus = "enqueued" // 已加入佇列
	TargetStatus_SENDING  TargetStatus = "sending"  // 發送中
	TargetStatus_SENT     TargetStatus = "sent"     // 已發送
	TargetStatus_SUCCESS  TargetStatus = "success"  // 發送成功
	TargetStatus_FAILED   TargetStatus = "failed"   // 發送失敗
	TargetStatus_CANCELED TargetStatus = "canceled" // 已取消
)

type Target struct {
	Id              string         `gorm:"primaryKey" json:"id"`
	MessageId       string         `json:"message_id"`                       // 關聯的 message.id
	Receiver        string         `json:"receiver"`                         // 接收者 (AES-128 加密的 email 或 receiver_phone)
	ReceiverHash    string         `json:"receiver_hash"`                    // 接收者的 MD5 哈希值 (email 或 phone)
	Status          TargetStatus   `json:"status"`                           // 目標發送狀態，使用 TargetStatus 枚舉
	QueueId         string         `json:"queue_id"`                         // 關聯的 queue.id
	DriverTraceId   string         `json:"driver_trace_id"`                  // 驅動器的 trace_id
	Provider        Provider       `json:"provider"`                         // 發送商，使用 Provider 枚舉
	ProviderTraceId string         `json:"provider_trace_id"`                // 發送商的 trace_id
	CreatedAt       time.Time      `json:"created_at" gorm:"type:timestamp"` // 創建時間
	UpdatedAt       time.Time      `json:"updated_at" gorm:"type:timestamp"` // 更新時間
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamp"` // 刪除時間，使用 GORM 的軟刪除

	// 外鍵關聯
	Message Message `gorm:"foreignKey:MessageId" json:"message"`
	Queue   Queue   `gorm:"foreignKey:QueueId" json:"queue"`
}

func (Target) TableName() string {
	return "notify.targets"
}
