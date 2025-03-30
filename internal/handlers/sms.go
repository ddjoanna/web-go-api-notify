package handler

import (
	"context"
	consumer "notify-service/internal/consumer"

	"github.com/IBM/sarama"
	log "github.com/sirupsen/logrus"
)

type SmsHandler struct{}

func NewSmsHandler() consumer.TopicHandler {
	return &SmsHandler{}
}

// TODO: 實作 Handle 方法
func (h SmsHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage) error {
	log.WithContext(ctx).Infof("Processing SMS message: %s", string(msg.Value))
	return nil
}
