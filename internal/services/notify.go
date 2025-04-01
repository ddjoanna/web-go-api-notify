package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	shared "notify-service/internal"
	component "notify-service/internal/components"
	entity "notify-service/internal/entities"
	model "notify-service/internal/models"
	util "notify-service/internal/utils"
	errorpb "proto/pkg/notify/v1/error"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type NotifyService struct {
	BaseService
	db        *gorm.DB
	config    *shared.Config
	snowflake *snowflake.Node
	aesGcm    *component.AesGcm
	producer  sarama.SyncProducer
}

func NewNotifyService(
	db *gorm.DB,
	config *shared.Config,
	snowflake *snowflake.Node,
	aesGcm *component.AesGcm,
	producer sarama.SyncProducer,
) *NotifyService {
	return &NotifyService{
		db:        db,
		config:    config,
		snowflake: snowflake,
		aesGcm:    aesGcm,
		producer:  producer,
	}
}

func (s NotifyService) PublishSmsMessage(ctx context.Context, in model.SendSmsRequest) (*entity.Message, error) {
	message := &entity.Message{
		Id:        s.snowflake.Generate().String(),
		Type:      entity.MessageType_SMS,
		Data:      in.Sms.Body,
		CreatedAt: time.Now(),
		Status:    entity.MessageStatus_PENDING,
	}

	if in.ScheduledAt != nil {
		message.ScheduledAt = in.ScheduledAt
		message.Status = entity.MessageStatus_SCHEDULED
	}

	queues, targets, err := s.prepareTargetsAndQueues(in.Receivers, message.Id, entity.MessageType_SMS)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("prepare targets and queues failed with db query: %v", err))
		return nil, err
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(message).Create(message).Error
		if err != nil {
			return s.ServerError("create message failed with db create query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}

		err = tx.Create(targets).Error
		if err != nil {
			return s.ServerError("create target failed with db create query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}

		err = tx.Create(queues).Error
		if err != nil {
			return s.ServerError("create queue failed with db create query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if message.ScheduledAt == nil {
		err := s.HandleEnqueue(ctx, entity.MessageType_SMS, queues)
		if err != nil {
			return nil, s.ServerError("handle enqueue failed", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}
	}

	return message, nil
}

func (s NotifyService) PublishMailMessage(ctx context.Context, in model.SendMailRequest) (*entity.Message, error) {
	message := &entity.Message{
		Id:            s.snowflake.Generate().String(),
		Type:          entity.MessageType_MAIL,
		Data:          in.Mail.Body,
		SenderName:    in.Mail.SenderName,
		SenderAddress: in.Mail.SenderAddress,
		Subject:       in.Mail.Subject,
		CreatedAt:     time.Now(),
		Status:        entity.MessageStatus_PENDING,
	}

	if in.ScheduledAt != nil {
		message.ScheduledAt = in.ScheduledAt
		message.Status = entity.MessageStatus_SCHEDULED
	}

	queues, targets, err := s.prepareTargetsAndQueues(in.Receivers, message.Id, entity.MessageType_MAIL)
	if err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("prepare targets and queues failed with db query: %v", err))
		return nil, err
	}
	transaction := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err = tx.Create(message).Error
		if err != nil {
			return s.ServerError("create message failed with db create query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}

		err = tx.Create(targets).Error
		if err != nil {
			return s.ServerError("create target failed with db create query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}

		err = tx.Create(queues).Error
		if err != nil {
			return s.ServerError("create queue failed with db create query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}
		return nil
	})

	if transaction != nil {
		return nil, transaction
	}

	if message.ScheduledAt == nil {
		err := s.HandleEnqueue(ctx, entity.MessageType_MAIL, queues)
		if err != nil {
			return nil, s.ServerError("handle enqueue failed", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}
	}

	return message, nil
}

func (s NotifyService) prepareTargetsAndQueues(
	receivers []string,
	messageId string,
	messageType entity.MessageType,
) (
	[]*entity.Queue,
	[]*entity.Target,
	error,
) {
	var queues []*entity.Queue
	var targets []*entity.Target

	createdAt := time.Now()
	provider, err := getProvider(messageType)
	if err != nil {
		return nil, nil, err
	}
	limit, err := s.getProviderBatchLimit(messageType)
	if err != nil {
		return nil, nil, err
	}

	chunks := util.ChunkArray(receivers, limit)

	for _, chunk := range chunks {
		queue := &entity.Queue{
			Id:        s.snowflake.Generate().String(),
			Status:    entity.QueueStatus_PENDING,
			MessageId: messageId,
			Driver:    entity.QueueDriver_KAFKA,
			CreatedAt: createdAt,
		}
		queues = append(queues, queue)

		for _, receiver := range chunk {
			receiverEncrypted, err := s.aesGcm.AesEncrypt(receiver)
			if err != nil {
				return nil, nil, err
			}

			receiverHash := util.Md5(receiver)

			target := &entity.Target{
				Id:           s.snowflake.Generate().String(),
				MessageId:    messageId,
				Receiver:     receiverEncrypted,
				ReceiverHash: receiverHash,
				Status:       entity.TargetStatus_PENDING,
				QueueId:      queue.Id,
				Provider:     *provider,
				CreatedAt:    createdAt,
			}
			targets = append(targets, target)
		}
	}

	return queues, targets, nil
}

func getProvider(messageType entity.MessageType) (*entity.Provider, error) {
	switch messageType {
	case entity.MessageType_SMS:
		provider := entity.Provider_MITAKE
		return &provider, nil
	case entity.MessageType_MAIL:
		provider := entity.Provider_SENDGRID
		return &provider, nil
	default:
		return nil, fmt.Errorf("unsupported message type: %v", messageType)
	}
}

func (s NotifyService) getProviderBatchLimit(messageType entity.MessageType) (int, error) {
	switch messageType {
	case entity.MessageType_SMS:
		return s.config.SmsProviderBatchLimit, nil
	case entity.MessageType_MAIL:
		return s.config.MailProviderBatchLimit, nil
	default:
		return 0, fmt.Errorf("unsupported message type: %v", messageType)
	}
}

func (s NotifyService) CancelScheduledByMessageId(ctx context.Context, messageId string) error {
	createdAt, err := util.ConvertSnowflakeToTime(messageId)
	if err != nil {
		return s.ServerError("cancel scheduled message failed with snowflake id", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
	}

	// 避免全表搜尋加上時間戳範圍
	startAt := createdAt.Add(-1 * time.Hour)
	endAt := createdAt.Add(1 * time.Hour)
	var message entity.Message
	err = s.db.WithContext(ctx).
		Where("id = ?", messageId).
		Where("created_at between ? and ?", startAt, endAt).
		First(&message).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.ServerError("message not found", errorpb.ErrorReasonCode_ERR_NOTIFY_MESSAGE_NOT_FOUND)
	}

	if err != nil {
		return s.ServerError("cancel scheduled message failed with db query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
	}

	if message.Status == entity.MessageStatus_ENQUEUED {
		return s.ServerError("message is already enqueued", errorpb.ErrorReasonCode_ERR_NOTIFY_MESSAGE_IS_ENQUEUE_CANNOT_CANCEL)
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(entity.Message{}).
			Where("id = ?", messageId).
			Updates(map[string]interface{}{
				"status":     string(entity.MessageStatus_CANCELED),
				"deleted_at": time.Now(),
			}).Error
		if err != nil {
			return s.ServerError("cancel scheduled message failed with db update query", errorpb.ErrorReasonCode_ERR_COMMON_INTERNAL)
		}

		err = tx.Model(entity.Queue{}).
			Where("message_id = ?", message.Id).
			Where("created_at between ? and ?", startAt, endAt).
			Updates(map[string]interface{}{
				"status":     string(entity.QueueStatus_CANCELED),
				"deleted_at": time.Now(),
			}).Error
		if err != nil {
			log.WithContext(ctx).Error(fmt.Printf("update queue status failed with message id %s error: %v", message.Id, err))
			return err
		}

		err = tx.Model(entity.Target{}).
			Where("message_id = ?", message.Id).
			Where("created_at between ? and ?", startAt, endAt).
			Updates(map[string]interface{}{
				"status":     string(entity.TargetStatus_CANCELED),
				"deleted_at": time.Now(),
			}).Error
		if err != nil {
			log.WithContext(ctx).Error(fmt.Printf("update target status failed with message id %s error: %v", message.Id, err))
			return err
		}
		return nil
	})
}

func (s NotifyService) ListStatusWithPaging(ctx context.Context, in model.ListStatusWithPagingRequest) ([]*entity.Target, int64, error) {
	var targets []*entity.Target

	query := s.db.Model(&entity.Target{}).
		Unscoped().
		Preload("Message").
		Joins("JOIN messages ON messages.id = targets.message_id").
		Where("messages.type = ?", in.MessageType).
		Where("messages.created_at BETWEEN ? AND ?", in.StartAt, in.EndAt).
		Where("targets.created_at BETWEEN ? AND ?", in.StartAt, in.EndAt)

	if in.MessageId != "" {
		query = query.Where("message_id = ?", in.MessageId)
	}

	if in.Receiver != "" {
		receiverHash := util.Md5(in.Receiver)
		query = query.Where("targets.receiver_hash = ?", receiverHash)
	}

	var total int64
	query.Count(&total)

	if in.Page != nil && in.Page.SortField != "" && in.Page.SortOrder != "" {
		query = query.Order(fmt.Sprintf("%s %s", in.Page.SortField, in.Page.SortOrder))
	}

	if in.Page != nil && in.Page.Index > 0 && in.Page.Size > 0 {
		query = query.Offset(int((in.Page.Index - 1) * in.Page.Size)).Limit(int(in.Page.Size))
	}
	err := query.Find(&targets).Error
	if err != nil {
		return nil, 0, err
	}

	return targets, total, nil
}

func (s *NotifyService) HandleEnqueue(
	ctx context.Context,
	messageType entity.MessageType,
	queues []*entity.Queue,
) error {
	queueIDs, messageIDs := extractBatchData(queues)

	if err := s.batchUpdateDatabase(
		ctx,
		queueIDs,
		messageIDs,
		entity.QueueStatus_PROCESS,
		nil,
	); err != nil {
		return fmt.Errorf("failed to update process queues: %w", err)
	}

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup
	successes := &sync.Map{}
	traceIDs := &sync.Map{}

	for _, queue := range queues {
		wg.Add(1)
		sem <- struct{}{}

		go func(q *entity.Queue) {
			defer func() {
				<-sem
				wg.Done()
			}()

			msg, err := s.createKafkaMessage(messageType, q)
			if err != nil {
				log.WithContext(ctx).Errorf("Failed to create kafka message: %v", err)
				return
			}

			if err := withRetry(ctx, 3, func() error {
				_, _, err := s.producer.SendMessage(msg)
				return err
			}); err != nil {
				log.WithContext(ctx).Errorf("Failed to send message to kafka: %v", err)
			} else {
				traceId, err := msg.Key.Encode()
				if err != nil {
					log.WithContext(ctx).Errorf("Failed to encode kafka message key: %v", err)
					return
				}

				successes.Store(q.Id, q.MessageId)
				traceIDs.Store(q.Id, traceId)
			}
		}(queue)
	}

	wg.Wait()

	// 提取成功隊列的 ID 和 Message ID
	successQueueIDs := make([]string, 0)
	successMessageIDs := make([]string, 0)
	traceIDMap := make(map[string]string)

	successes.Range(func(key, value any) bool {
		successQueueIDs = append(successQueueIDs, key.(string))
		successMessageIDs = append(successMessageIDs, value.(string))
		return true
	})

	traceIDs.Range(func(key, value any) bool {
		traceIDMap[key.(string)] = string(value.([]byte)) // 轉換 byte 到 string
		return true
	})

	if err := s.batchUpdateDatabase(
		ctx,
		successQueueIDs,
		successMessageIDs,
		entity.QueueStatus_ENQUEUED,
		traceIDMap,
	); err != nil {
		log.WithContext(ctx).Errorf("Failed to update enqueued queues: %v", err)
	}

	return nil
}

func (s *NotifyService) batchUpdateDatabase(
	ctx context.Context,
	queueIDs []string,
	messageIDs []string,
	queueStatus entity.QueueStatus,
	traceIDs map[string]string,
) error {
	transaction := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Queue{}).
			Where("id IN ?", queueIDs).
			Updates(map[string]interface{}{"status": queueStatus}).
			Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.Message{}).
			Where("id IN ?", messageIDs).
			Updates(map[string]interface{}{"status": queueStatus}).
			Error; err != nil {
			return err
		}

		// 更新 Target 狀態和 driver_trace_id
		if len(traceIDs) > 0 {
			for queueId, traceId := range traceIDs {
				if err := tx.Model(&entity.Target{}).
					Where("queue_id = ?", queueId).
					Update("status", queueStatus).
					Update("driver_trace_id", traceId).
					Error; err != nil {
					return err
				}
			}
		} else {
			if err := tx.Model(&entity.Target{}).
				Where("queue_id IN ?", queueIDs).
				Updates(map[string]interface{}{"status": queueStatus}).
				Error; err != nil {
				return err
			}
		}

		return nil
	})

	if transaction != nil {
		return transaction
	}
	return nil
}

func (s *NotifyService) createKafkaMessage(
	messageType entity.MessageType,
	queue *entity.Queue,
) (*sarama.ProducerMessage, error) {
	jsonData, err := json.Marshal(queue)
	if err != nil {
		return nil, fmt.Errorf("marshal queue failed: %w", err)
	}

	driverTraceId := s.snowflake.Generate().String()
	msg := &sarama.ProducerMessage{
		Topic: getKafkaTopicByMessageType(messageType),
		Key:   sarama.StringEncoder(driverTraceId),
		Value: sarama.ByteEncoder(jsonData),
	}

	return msg, nil
}

func extractBatchData(queues []*entity.Queue) ([]string, []string) {
	queueIDs := make([]string, 0, len(queues))
	messageIDs := make([]string, 0, len(queues))

	for _, queue := range queues {
		queueIDs = append(queueIDs, queue.Id)
		messageIDs = append(messageIDs, queue.MessageId)
	}

	return queueIDs, messageIDs
}

func withRetry(ctx context.Context, maxRetries int, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * time.Duration(i+1)):
		}
	}
	return err
}

func getKafkaTopicByMessageType(messageType entity.MessageType) string {
	switch messageType {
	case entity.MessageType_SMS:
		return shared.KafkaTopicSms
	case entity.MessageType_MAIL:
		return shared.KafkaTopicMail
	default:
		return ""
	}
}
