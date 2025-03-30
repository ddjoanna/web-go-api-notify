package job

import (
	"context"
	entity "notify-service/internal/entities"
	service "notify-service/internal/services"
	"time"

	shared "notify-service/internal"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DispatchScheduledMessagesJob struct {
	db            *gorm.DB
	notifyService *service.NotifyService
	config        *shared.Config
}

func NewDispatchScheduledMessagesJob(
	db *gorm.DB,
	notifyService *service.NotifyService,
	config *shared.Config,
) *DispatchScheduledMessagesJob {
	return &DispatchScheduledMessagesJob{
		db:            db,
		notifyService: notifyService,
		config:        config,
	}
}

func (j *DispatchScheduledMessagesJob) Execute(ctx context.Context) error {
	// 設定 startAt 為 ScheduleLimitDays 天前的時間
	messageScheduleLimitDays := j.config.ScheduleLimitDays
	now := time.Now()
	startAt := now.Add(-time.Duration(messageScheduleLimitDays) * 24 * time.Hour)

	var messages []entity.Message
	err := j.db.WithContext(ctx).
		Model(&entity.Message{}).
		Preload("Queues").
		Where("messages.status = ?", entity.MessageStatus_SCHEDULED).
		Where("scheduled_at <= ?", now).
		Where("messages.created_at >= ?", startAt).
		Find(&messages).
		Error

	if err != nil {
		log.WithContext(ctx).WithError(err).Error("error fetching scheduled messages")
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	for _, message := range messages {
		transaction := j.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			err = j.notifyService.HandleEnqueue(ctx, message.Type, message.Queues)
			if err != nil {
				log.WithContext(ctx).WithError(err).Error("error handling enqueue")
				return err
			}
			return nil
		})
		if transaction != nil {
			continue
		}
	}

	return nil
}
