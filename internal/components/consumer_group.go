package component

import (
	"context"
	"errors"
	"strings"

	shared "notify-service/internal"
	"notify-service/internal/consumer"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/fx"
)

func NewConsumerGroup(
	lc fx.Lifecycle,
	config *shared.Config,
	consumer *consumer.Consumer,
) []sarama.ConsumerGroup {
	topics := []string{
		shared.KafkaTopicSms,
		shared.KafkaTopicMail,
	}
	version, err := sarama.ParseKafkaVersion(config.KafkaVersion)
	if err != nil {
		log.WithError(err).Fatalf("error parsing Kafka version: %s", config.KafkaVersion)
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.Version = version
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerConfig.Consumer.Offsets.AutoCommit.Enable = false

	// Setup Kafka consumer
	log.Infof("Consumer connecting to Kafka broker at %s", config.KafkaBrokers)
	var consumerGroups []sarama.ConsumerGroup

	for i := 0; i < config.KafkaConsumerGroupInstanceNum; i++ {
		consumerGroup, err := sarama.NewConsumerGroup(
			strings.Split(config.KafkaBrokers, ","),
			shared.KafkaGroupIdNotify,
			consumerConfig,
		)
		if err != nil {
			log.WithError(err).Fatalf("Error creating consumer group client: %v", err)
		}
		consumerGroups = append(consumerGroups, consumerGroup)

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					propagators := propagation.TraceContext{}
					consumer := otelsarama.WrapConsumerGroupHandler(consumer, otelsarama.WithPropagators(propagators))
					for {
						err := consumerGroup.Consume(context.Background(), topics, consumer)
						if err != nil {
							if errors.Is(err, sarama.ErrClosedConsumerGroup) {
								return
							}
							log.Panicf("Error from consumer: %v", err)
						}
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return consumerGroup.Close()
			},
		})
	}

	return consumerGroups
}
