package entity

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatus_SENT      EventStatus = "sent"      // 已發送
	EventStatus_DELIVERED EventStatus = "delivered" // 已送達
	EventStatus_OPENED    EventStatus = "opened"    // 已開啟
	EventStatus_CLICKED   EventStatus = "clicked"   // 已點擊
	EventStatus_FAILED    EventStatus = "failed"    // 發送失敗
)

var (
	ConvertEventStatusWithString = map[string]EventStatus{
		string(EventStatus_SENT):      EventStatus_SENT,
		string(EventStatus_DELIVERED): EventStatus_DELIVERED,
		string(EventStatus_OPENED):    EventStatus_OPENED,
		string(EventStatus_CLICKED):   EventStatus_CLICKED,
		string(EventStatus_FAILED):    EventStatus_FAILED,
	}
)

type Event struct {
	Id              string          `gorm:"primaryKey" json:"id"`
	Provider        string          `json:"provider"`                         // 事件來源
	Status          EventStatus     `json:"status"`                           // 事件狀態，使用枚舉來規範
	ProviderTraceId string          `json:"provider_trace_id"`                // 事件來源的追蹤 ID
	QueueId         string          `json:"queue_id"`                         // queues.id
	Data            json.RawMessage `json:"data"`                             // 事件資料
	CreatedAt       time.Time       `json:"created_at" gorm:"type:timestamp"` // 創建時間
	UpdatedAt       time.Time       `json:"updated_at" gorm:"type:timestamp"` // 更新時間
	DeletedAt       gorm.DeletedAt  `json:"deleted_at" gorm:"type:timestamp"` // 刪除時間，使用 gorm 的軟刪除
}

func (Event) TableName() string {
	return "notify.events"
}
