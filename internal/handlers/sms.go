package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	shared "notify-service/internal"
	consumer "notify-service/internal/consumer"
	entity "notify-service/internal/entities"
	service "notify-service/internal/services"
	smser "notify-service/internal/smser"
	util "notify-service/internal/utils"

	"github.com/IBM/sarama"
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SmsHandler struct {
	db        *gorm.DB
	config    *shared.Config
	snowflake *snowflake.Node
	service   *service.SmsService
}

func NewSmsHandler(
	db *gorm.DB,
	config *shared.Config,
	snowflake *snowflake.Node,
	service *service.SmsService,
) consumer.TopicHandler {
	return &SmsHandler{
		db:        db,
		config:    config,
		snowflake: snowflake,
		service:   service,
	}
}

func (h SmsHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage) error {
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

	response := h.service.SendSms(ctx, queue.Targets, queue.Message)

	return h.handleSendResult(ctx, queue, &response)
}

func (h SmsHandler) prepareLogContext(ctx context.Context, msg *sarama.ConsumerMessage) *log.Entry {
	return log.WithContext(ctx).WithFields(log.Fields{
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"timestamp": msg.Timestamp,
	})
}

func (h SmsHandler) fetchQueueData(ctx context.Context, req entity.Queue) (*entity.Queue, error) {
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

func (h SmsHandler) updateProcessingStatus(ctx context.Context, queue *entity.Queue) error {
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

func (h SmsHandler) handleSendResult(ctx context.Context, queue *entity.Queue, response *smser.SmsBatchResponse) error {
	if response.Status != "sent" {
		return h.createFailureEventAndUpdateStatus(ctx, queue, response)
	}

	return h.createSuccessEventAndUpdateStatus(ctx, queue, response)
}

func (h SmsHandler) createFailureEventAndUpdateStatus(ctx context.Context, queue *entity.Queue, response *smser.SmsBatchResponse) error {
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

		for _, row := range response.SmsResponse {
			jsonResponse, err := json.Marshal(row.ProviderResponse)
			if err != nil {
				log.WithContext(ctx).WithError(err).Error("Failed to marshal provider response")
				continue
			}

			event := entity.Event{
				Id:              h.snowflake.Generate().String(),
				Provider:        h.config.SmsProvider,
				Status:          entity.EventStatus_FAILED,
				ProviderTraceId: row.TraceId,
				QueueId:         queue.Id,
				Data:            jsonResponse,
				CreatedAt:       time.Now(),
			}
			if err := tx.Model(&entity.Event{}).
				Create(&event).
				Error; err != nil {
				log.WithContext(ctx).WithError(err).Error("Error creating event")
				continue
			}

			continue
		}

		return nil
	})

	return err
}

func (h SmsHandler) createSuccessEventAndUpdateStatus(ctx context.Context, queue *entity.Queue, response *smser.SmsBatchResponse) error {
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

		for _, row := range response.SmsResponse {
			jsonResponse, err := json.Marshal(row.ProviderResponse)
			if err != nil {
				log.WithContext(ctx).WithError(err).Error("Failed to marshal provider response")
				continue
			}

			event := entity.Event{
				Id:              h.snowflake.Generate().String(),
				Provider:        h.config.SmsProvider,
				Status:          entity.EventStatus_SENT,
				ProviderTraceId: row.TraceId,
				QueueId:         queue.Id,
				Data:            jsonResponse,
				CreatedAt:       time.Now(),
			}
			if err := tx.Model(&entity.Event{}).
				Create(&event).
				Error; err != nil {
				log.WithContext(ctx).WithError(err).Error("Error creating event")
			}

			if err := tx.Model(entity.Target{}).
				Where("id = ?", row.TraceId).
				Updates(map[string]interface{}{
					"provider_trace_id": row.TraceId,
					"status":            string(row.Status),
				}).
				Error; err != nil {
				log.WithContext(ctx).WithError(err).WithFields(log.Fields{
					"target_id": row.TraceId,
				}).Error("Error updating target status")
			}
			continue
		}

		return nil
	})

	return err
}
