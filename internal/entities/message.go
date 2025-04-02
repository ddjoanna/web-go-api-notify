package entity

import (
	"time"

	"gorm.io/gorm"
)

type MessageType string

const (
	MessageType_SMS  MessageType = "sms"  // 簡訊
	MessageType_MAIL MessageType = "mail" // 郵件
)

type MessageStatus string

const (
	MessageStatus_PENDING   MessageStatus = "pending"   // 待處理
	MessageStatus_ENQUEUED  MessageStatus = "enqueued"  // 已加入佇列
	MessageStatus_SCHEDULED MessageStatus = "scheduled" // 已排定
	MessageStatus_SENT      MessageStatus = "sent"      // 已發送
	MessageStatus_FAILED    MessageStatus = "failed"    // 發送失敗
	MessageStatus_CANCELED  MessageStatus = "canceled"  // 已取消
)

type Message struct {
	Id            string         `gorm:"primaryKey" json:"id"`
	Type          MessageType    `json:"type"`                             // 消息類型，使用 MessageType 枚舉
	Data          string         `json:"data"`                             // 消息內容
	SenderName    string         `json:"sender_name"`                      // 寄件者名稱
	SenderAddress string         `json:"sender_address"`                   // 寄件者地址
	Subject       string         `json:"subject"`                          // 主旨
	ScheduledAt   *time.Time     `json:"scheduled_at"`                     // 預約時間，使用指標來處理空值
	Status        MessageStatus  `json:"status"`                           // 消息狀態，使用 MessageStatus 枚舉
	CreatedAt     time.Time      `json:"created_at" gorm:"type:timestamp"` // 創建時間
	UpdatedAt     time.Time      `json:"updated_at" gorm:"type:timestamp"` // 更新時間
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamp"` // 刪除時間，使用 GORM 的軟刪除

	// 外鍵關聯
	Queues  []*Queue  `gorm:"foreignKey:MessageId" json:"queues"`
	Targets []*Target `gorm:"foreignKey:MessageId" json:"targets"`
}

func (Message) TableName() string {
	return "notify.messages"
}
