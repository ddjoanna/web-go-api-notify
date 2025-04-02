package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	shared "notify-service/internal"
	consumer "notify-service/internal/consumer"
	entity "notify-service/internal/entities"
	mailer "notify-service/internal/mailer"
	service "notify-service/internal/services"
	util "notify-service/internal/utils"

	"github.com/IBM/sarama"
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MailHandler struct {
	db        *gorm.DB
	config    *shared.Config
	snowflake *snowflake.Node
	service   *service.MailService
}

func NewMailHandler(
	db *gorm.DB,
	config *shared.Config,
	snowflake *snowflake.Node,
	service *service.MailService,
) consumer.TopicHandler {
	return &MailHandler{
		db:        db,
		config:    config,
		snowflake: snowflake,
		service:   service,
	}
}

func (h MailHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage) error {
	logCtx := h.prepareLogContext(ctx, msg)

	var req entity.Queue
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		logCtx.WithError(err).Error("Invalid message format")
		return err
	}

	queue, err := h.fetchQueueData(ctx, req)
	if err != nil {
		return h.createFailureEventAndUpdateStatus(ctx, queue, nil)
	}

	if err := h.updateProcessingStatus(ctx, queue); err != nil {
		return h.createFailureEventAndUpdateStatus(ctx, queue, nil)
	}

	response := h.service.SendEmail(ctx, queue.Targets, queue.Message)

	return h.handleSendResult(ctx, queue, &response)
}

func (h MailHandler) prepareLogContext(ctx context.Context, msg *sarama.ConsumerMessage) *log.Entry {
	return log.WithContext(ctx).WithFields(log.Fields{
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"timestamp": msg.Timestamp,
	})
}

func (h MailHandler) fetchQueueData(ctx context.Context, req entity.Queue) (*entity.Queue, error) {
	startAt, err := util.ConvertSnowflakeToTime(req.Id)
	if err != nil {
		return nil, fmt.Errorf("error convert snowflake to time: %w", err)
	}

	var queue entity.Queue
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err = tx.
			Preload("Targets").
			Preload("Message").
			Where("created_at >= ?", startAt).
			Find(&queue, "id = ?", req.Id).
			Error
		if err != nil {
			return fmt.Errorf("error fetching queue: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

func (h MailHandler) updateProcessingStatus(ctx context.Context, queue *entity.Queue) error {
	startAt, err := util.ConvertSnowflakeToTime(queue.Id)
	if err != nil {
		return fmt.Errorf("error convert snowflake to time: %w", err)
	}

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err = tx.Model(entity.Target{}).
			Where("queue_id = ?", queue.Id).
			Where("created_at >= ?", startAt).
			Updates(map[string]interface{}{
				"status": entity.TargetStatus_SENDING,
			}).
			Error
		if err != nil {
			return fmt.Errorf("error updating target status to SENDING: %w", err)
		}

		err = tx.Model(entity.Queue{}).
			Where("id = ?", queue.Id).
			Where("created_at >= ?", startAt).
			Updates(map[string]interface{}{
				"status": entity.QueueStatus_SENDING,
			}).
			Error
		if err != nil {
			return fmt.Errorf("error updating queue status to SENDING: %w", err)
		}
		return nil
	})
}

func (h MailHandler) handleSendResult(ctx context.Context, queue *entity.Queue, response *mailer.MailResponse) error {
	if response.Status != "sent" {
		return h.createFailureEventAndUpdateStatus(ctx, queue, response)
	}

	return h.createSuccessEventAndUpdateStatus(ctx, queue, response)
}

func (h MailHandler) createFailureEventAndUpdateStatus(ctx context.Context, queue *entity.Queue, response *mailer.MailResponse) error {
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Queue{}).
			Where("id = ?", queue.Id).
			Update("status", entity.QueueStatus_FAILED).
			Error; err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to update queue status")
			return err
		}

		if err := tx.Model(&entity.Message{}).
			Where("id = ?", queue.MessageId).
			Update("status", entity.MessageStatus_FAILED).
			Error; err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to update queue status")
			return err
		}

		if err := tx.Model(&entity.Target{}).
			Where("queue_id = ?", queue.Id).
			Update("status", entity.TargetStatus_FAILED).
			Error; err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to update targets status")
			return err
		}

		jsonResponse, err := json.Marshal(response.ProviderResponse)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to marshal provider response")
			return err
		}

		event := entity.Event{
			Id:              h.snowflake.Generate().String(),
			Provider:        h.config.MailProvider,
			Status:          entity.EventStatus_FAILED,
			ProviderTraceId: response.TraceId,
			QueueId:         queue.Id,
			Data:            jsonResponse,
			CreatedAt:       time.Now(),
		}
		err = h.db.Create(&event).Error
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to create event")
			return err
		}

		return nil
	})

	return err
}

func (h MailHandler) createSuccessEventAndUpdateStatus(ctx context.Context, queue *entity.Queue, response *mailer.MailResponse) error {
	err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Queue{}).
			Where("id = ?", queue.Id).
			Update("status", entity.QueueStatus_SUCCESS).
			Error; err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to update queue status")
			return err
		}

		if err := tx.Model(&entity.Message{}).
			Where("id = ?", queue.MessageId).
			Update("status", entity.MessageStatus_SENT).
			Error; err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to update queue status")
			return err
		}

		eventStatus := entity.ConvertEventStatusWithString[response.Status]
		jsonResponse, err := json.Marshal(response.ProviderResponse)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to marshal provider response")
			return err
		}
		event := entity.Event{
			Id:              h.snowflake.Generate().String(),
			Provider:        h.config.MailProvider,
			Status:          eventStatus,
			ProviderTraceId: response.TraceId,
			QueueId:         queue.Id,
			Data:            jsonResponse,
			CreatedAt:       time.Now(),
		}
		err = tx.Create(&event).Error
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Error creating event")
		}

		for _, target := range queue.Targets {
			if err := tx.Model(entity.Target{}).
				Where("id = ?", target.Id).
				Updates(map[string]interface{}{
					"provider_trace_id": response.TraceId,
					"status":            string(response.Status),
				}).
				Error; err != nil {
				log.WithContext(ctx).WithError(err).WithFields(log.Fields{
					"target_id": target.Id,
				}).Error("Error updating target status")
				continue
			}
		}

		return nil
	})

	return err
}
