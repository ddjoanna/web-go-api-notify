package main

import (
	"fmt"
	"os"

	shared "notify-service/internal"
	component "notify-service/internal/components"
	"notify-service/internal/consumer"
	handler "notify-service/internal/handlers"
	mailer "notify-service/internal/mailer"
	service "notify-service/internal/services"
	smser "notify-service/internal/smser"

	"github.com/IBM/sarama"
	"github.com/bwmarrin/snowflake"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"github.com/urfave/cli/v2"
	metricssdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var (
	config shared.Config
)

func main() {
	app := &cli.App{
		Name:  "notify",
		Usage: "notify service server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "grpc-port",
				Usage:       "gRPC server port",
				Value:       50051,
				EnvVars:     []string{"GRPC_PORT"},
				Destination: &config.GrpcPort,
			},
			&cli.StringFlag{
				Name:        "postgres-host",
				Usage:       "PostgresSQL DB host address",
				EnvVars:     []string{"POSTGRES_HOST"},
				Destination: &config.PostgresHost,
			},
			&cli.IntFlag{
				Name:        "postgres-port",
				Usage:       "PostgresSQL DB port number",
				Value:       5432,
				EnvVars:     []string{"POSTGRES_PORT"},
				Destination: &config.PostgresPort,
			},
			&cli.StringFlag{
				Name:        "postgres-user",
				Usage:       "PostgresSQL DB user",
				EnvVars:     []string{"POSTGRES_USER"},
				Destination: &config.PostgresUser,
			},
			&cli.StringFlag{
				Name:        "postgres-password",
				Usage:       "PostgresSQL DB password",
				EnvVars:     []string{"POSTGRES_PASSWORD"},
				Destination: &config.PostgresPassword,
			},
			&cli.StringFlag{
				Name:        "postgres-db",
				Usage:       "PostgresSQL DB name",
				EnvVars:     []string{"POSTGRES_DB"},
				Destination: &config.PostgresDb,
			},
			&cli.StringFlag{
				Name:        "postgres-schema",
				Usage:       "PostgresSQL DB schema",
				EnvVars:     []string{"POSTGRES_SCHEMA"},
				Destination: &config.PostgresSchema,
			},
			&cli.IntFlag{
				Name:        "db-max-idle-conns",
				Usage:       "PostgresSQL DB max idle connections",
				EnvVars:     []string{"DB_MAX_IDLE_CONNS"},
				Value:       2,
				Destination: &config.DbMaxIdleConns,
			},
			&cli.IntFlag{
				Name:        "db-max-open-conns",
				Usage:       "PostgresSQL DB max open connections",
				EnvVars:     []string{"DB_MAX_OPEN_CONNS"},
				Value:       5,
				Destination: &config.DbMaxOpenConns,
			},
			&cli.StringFlag{
				Name:        "otlp-service-name",
				Usage:       "Service name for observability",
				EnvVars:     []string{"OTLP_SERVICE_NAME"},
				Destination: &config.OtlpServiceName,
			},
			&cli.StringFlag{
				Name:        "otlp-endpoint",
				Usage:       "The endpoint of the OTLP collector",
				EnvVars:     []string{"OTLP_ENDPOINT"},
				Destination: &config.OtlpEndpoint,
			},
			&cli.StringFlag{
				Name:        "log-format",
				Usage:       "Log format",
				EnvVars:     []string{"LOG_FORMAT"},
				Destination: &config.LogFormat,
			},
			&cli.StringFlag{
				Name:        "aes-key",
				Usage:       "AES key",
				EnvVars:     []string{"AES_KEY"},
				Destination: &config.AESKey,
			},
			&cli.StringFlag{
				Name:        "sms-provider",
				Usage:       "SMS provider",
				EnvVars:     []string{"SMS_PROVIDER"},
				Destination: &config.SmsProvider,
			},
			&cli.StringFlag{
				Name:        "mail-provider",
				Usage:       "Mail provider",
				EnvVars:     []string{"MAIL_PROVIDER"},
				Destination: &config.MailProvider,
			},
			&cli.StringFlag{
				Name:        "kafka-broker",
				Usage:       "Kafka broker",
				EnvVars:     []string{"KAFKA_BROKERS"},
				Destination: &config.KafkaBrokers,
			},
			&cli.StringFlag{
				Name:        "kafka-version",
				Usage:       "Kafka version",
				EnvVars:     []string{"KAFKA_VERSION"},
				Destination: &config.KafkaVersion,
			},
			&cli.IntFlag{
				Name:        "kafka-consumer-group-instance-num",
				Usage:       "Kafka consumer group instance number",
				EnvVars:     []string{"KAFKA_CONSUMER_GROUP_INSTANCE_NUM"},
				Value:       1,
				Destination: &config.KafkaConsumerGroupInstanceNum,
			},
			&cli.StringFlag{
				Name:        "sendgrid-api-token",
				Usage:       "Sendgrid API Token",
				EnvVars:     []string{"SENDGRID_TOKEN"},
				Destination: &config.SendgridToken,
			},
			&cli.StringFlag{
				Name:        "mitake-user-name",
				Usage:       "Mitake user name",
				EnvVars:     []string{"MITAKE_USER_NAME"},
				Destination: &config.MitakeUserName,
			},
			&cli.StringFlag{
				Name:        "mitake-password",
				Usage:       "Mitake password",
				EnvVars:     []string{"MITAKE_PASSWORD"},
				Destination: &config.MitakePassword,
			},
		},
		Action: execute,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func execute(cCtx *cli.Context) error {
	log.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
	)))
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	switch config.LogFormat {
	case shared.LOG_FORMAT_JSON:
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}

	log.Infof("Starting %s", config.OtlpServiceName)

	fx.New(
		fx.Supply(&config),
		fx.Provide(
			component.NewOtlpConn,
			component.NewTracerProvider,
			component.NewMeterProvider,
			component.NewSnowflake,
			component.NewAesGcm,
			component.NewDb,
			component.NewConsumerGroup,
			component.NewRestyClient,
			consumer.NewConsumer,
			provideSmsProvider,
			provideMailProvider,
			service.NewSmsService,
			service.NewMailService,
		),
		fx.Invoke(
			func(*tracesdk.TracerProvider) {},
			func(*metricssdk.MeterProvider) {},
			func(*gorm.DB) {},
			func([]sarama.ConsumerGroup) {},
			registerSmsHandler,
			registerMailHandler,
		),
	).Run()
	return nil
}

func registerSmsHandler(
	consumer *consumer.Consumer,
	smsService *service.SmsService,
	db *gorm.DB,
	config *shared.Config,
	snowflake *snowflake.Node,
) {
	handler := handler.NewSmsHandler(db, config, snowflake, smsService)
	consumer.RegisterHandler(shared.KafkaTopicSms, handler)
}

func registerMailHandler(
	consumer *consumer.Consumer,
	mailService *service.MailService,
	db *gorm.DB,
	config *shared.Config,
	snowflake *snowflake.Node,
) {
	handler := handler.NewMailHandler(db, config, snowflake, mailService)
	consumer.RegisterHandler(shared.KafkaTopicMail, handler)
}

func provideSmsProvider(config *shared.Config, resty *resty.Client) (smser.SmsProvider, error) {
	switch config.SmsProvider {
	case "mitake":
		return smser.NewMitakeSmser(config, resty), nil
	default:
		return nil, fmt.Errorf("unsupported SMS provider type: %s", config.SmsProvider)
	}
}

func provideMailProvider(config *shared.Config) (mailer.MailProvider, error) {
	switch config.MailProvider {
	case "sendgrid":
		return mailer.NewSendGridMailer(config), nil
	default:
		return nil, fmt.Errorf("unsupported mail provider type: %s", config.MailProvider)
	}
}
