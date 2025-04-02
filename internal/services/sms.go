package service

import (
	"context"
	shared "notify-service/internal"
	component "notify-service/internal/components"
	entity "notify-service/internal/entities"
	smser "notify-service/internal/smser"

	log "github.com/sirupsen/logrus"
)

type SmsService struct {
	config   *shared.Config
	provider smser.SmsProvider
	aesGcm   *component.AesGcm
}

func NewSmsService(
	config *shared.Config,
	provider smser.SmsProvider,
	aesGcm *component.AesGcm,
) *SmsService {
	return &SmsService{
		config:   config,
		provider: provider,
		aesGcm:   aesGcm,
	}
}

func (s *SmsService) SendSms(ctx context.Context, targets []entity.Target, message entity.Message) smser.SmsBatchResponse {
	requestReceviers := make([]smser.SmsReceiver, len(targets))
	for i, target := range targets {
		recevier, err := s.aesGcm.AesDecrypt(target.Receiver)
		if err != nil {
			log.WithContext(ctx).Errorf("failed to decrypt receiver with target id %s", target.Id)
			continue
		}
		requestReceviers[i] = smser.SmsReceiver{
			TargetId: target.Id,
			Receiver: recevier,
		}
	}

	request := smser.SmsBatchRequest{
		Receivers: requestReceviers,
		Message: smser.SmsMessage{
			MessageId: message.Id,
			Message:   message.Data,
		},
	}
	return s.provider.SendBatchSms(ctx, request)
}
