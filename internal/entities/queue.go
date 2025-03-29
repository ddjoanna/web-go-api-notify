package entity

import (
	"time"

	"gorm.io/gorm"
)

type QueueStatus string

const (
	QueueStatus_PENDING  QueueStatus = "pending"  // 待處理
	QueueStatus_ENQUEUED QueueStatus = "enqueued" // 已加入佇列
	QueueStatus_SENDING  QueueStatus = "sending"  // 發送中
	QueueStatus_SUCCESS  QueueStatus = "success"  // 發送成功
	QueueStatus_FAILED   QueueStatus = "failed"   // 發送失敗
	QueueStatus_CANCELED QueueStatus = "canceled" // 已取消
)

type QueueDriver string

const (
	QueueDriver_KAFKA QueueDriver = "kafka" // Kafka 驅動器
)

type Queue struct {
	Id        string         `gorm:"primaryKey" json:"id"`
	Status    QueueStatus    `json:"status"`                           // 佇列狀態，使用 QueueStatus 枚舉
	MessageId string         `json:"message_id"`                       // 關聯的消息 ID
	Driver    QueueDriver    `json:"driver"`                           // 佇列驅動器，使用 QueueDriver 枚舉
	CreatedAt time.Time      `json:"created_at" gorm:"type:timestamp"` // 創建時間
	UpdatedAt time.Time      `json:"updated_at" gorm:"type:timestamp"` // 更新時間
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"type:timestamp"` // 刪除時間，使用 GORM 的軟刪除

	// 外鍵關聯
	Targets []Target `gorm:"foreignKey:QueueId" json:"targets"`
}

func (Queue) TableName() string {
	return "notify.queues"
}
