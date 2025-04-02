package service

import (
	"context"
	shared "notify-service/internal"
	component "notify-service/internal/components"
	entity "notify-service/internal/entities"
	mailer "notify-service/internal/mailer"

	log "github.com/sirupsen/logrus"
)

type MailService struct {
	config   *shared.Config
	provider mailer.MailProvider
	aesGcm   *component.AesGcm
}

func NewMailService(
	config *shared.Config,
	provider mailer.MailProvider,
	aesGcm *component.AesGcm,
) *MailService {
	return &MailService{
		config:   config,
		provider: provider,
		aesGcm:   aesGcm,
	}
}

func (s *MailService) SendEmail(ctx context.Context, targets []entity.Target, message entity.Message) mailer.MailResponse {
	requestReceviers := make([]mailer.MailReceiver, len(targets))
	for i, target := range targets {
		recevier, err := s.aesGcm.AesDecrypt(target.Receiver)
		if err != nil {
			log.WithContext(ctx).Errorf("failed to decrypt receiver with target id %s", target.Id)
			continue
		}
		requestReceviers[i] = mailer.MailReceiver{
			Email: recevier,
		}
	}

	request := mailer.MailRequest{
		Receivers: requestReceviers,
		Message: mailer.MailMessage{
			SenderName:    message.SenderName,
			SenderAddress: message.SenderAddress,
			Subject:       message.Subject,
			Body:          message.Data,
		},
	}
	return s.provider.SendEmail(ctx, request)
}
