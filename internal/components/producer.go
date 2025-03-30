package component

import (
	"context"
	shared "notify-service/internal"
	"strings"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	log "github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func NewProducer(
	lc fx.Lifecycle,
	config *shared.Config,
) sarama.SyncProducer {
	kafkaVersion, err := sarama.ParseKafkaVersion(config.KafkaVersion)
	if err != nil {
		log.WithError(err).Fatalf("Error parsing Kafka version: %s", kafkaVersion)
	}

	// Setup Kafka producer
	log.Infof("Producer connecting to Kafka broker at %s", config.KafkaBrokers)
	producerConfig := sarama.NewConfig()
	producerConfig.Version = kafkaVersion
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(strings.Split(config.KafkaBrokers, ","), producerConfig)
	if err != nil {
		log.Panicf("Error creating producer: %v", err)
	}
	producer = otelsarama.WrapSyncProducer(producerConfig, producer)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return producer.Close()
		},
	})
	return producer
}
