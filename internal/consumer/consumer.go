package consumer

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/propagation"
)

type TopicHandler interface {
	Handle(ctx context.Context, msg *sarama.ConsumerMessage) error
}

type Consumer struct {
	handlers map[string]TopicHandler
}

func NewConsumer() *Consumer {
	return &Consumer{
		handlers: make(map[string]TopicHandler),
	}
}

// RegisterHandler 註冊一個 topic 的處理器
func (c *Consumer) RegisterHandler(topic string, handler TopicHandler) {
	c.handlers[topic] = handler
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	log.Info("Consumer group setup")
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Info("Consumer group cleanup")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		propagators := propagation.TraceContext{}
		ctx := propagators.Extract(context.Background(), otelsarama.NewConsumerMessageCarrier(msg))

		handler, ok := c.handlers[msg.Topic]
		// 根據錯誤類型決定是否重試或跳過
		if !ok {
			log.WithContext(ctx).Errorf("No handler found for topic %s", msg.Topic)
			session.MarkMessage(msg, "")
			continue
		}

		if err := handler.Handle(ctx, msg); err != nil {
			log.WithContext(ctx).WithError(err).Errorf("Failed to handle message")
			continue
		}

		// 成功處理消息後，標記偏移量
		session.MarkMessage(msg, "")
		session.Commit()
		log.WithContext(ctx).Infof("Successfully handled message: %s", string(msg.Value))
	}
	return nil
}
