package handler

import (
	"context"

	consumer "notify-service/internal/consumer"

	"github.com/IBM/sarama"
	log "github.com/sirupsen/logrus"
)

type MailHandler struct{}

func NewMailHandler() consumer.TopicHandler {
	return &MailHandler{}
}

// TODO: 實作 Handle 方法
func (h MailHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage) error {
	log.WithContext(ctx).Infof("Processing SMS message: %s", string(msg.Value))
	return nil
}
