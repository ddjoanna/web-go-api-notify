package service

import (
	"context"
	"errors"
	"fmt"
	shared "notify-service/internal"
	component "notify-service/internal/components"
	entity "notify-service/internal/entities"
	model "notify-service/internal/models"
	util "notify-service/internal/utils"
	errorpb "proto/pkg/notify/v1/error"
	"time"

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
}

func NewNotifyService(
	db *gorm.DB,
	config *shared.Config,
	snowflake *snowflake.Node,
	aesGcm *component.AesGcm,
) *NotifyService {
	return &NotifyService{
		db:        db,
		config:    config,
		snowflake: snowflake,
		aesGcm:    aesGcm,
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
