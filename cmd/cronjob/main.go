package main

import (
	"context"
	"os"

	shared "notify-service/internal"
	component "notify-service/internal/components"
	job "notify-service/internal/jobs"
	service "notify-service/internal/services"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"github.com/urfave/cli/v2"
	metricssdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"
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
			&cli.IntFlag{
				Name:        "schedule-limit-days",
				Usage:       "Schedule limit days",
				EnvVars:     []string{"SCHEDULE_LIMIT_DAYS"},
				Value:       30,
				Destination: &config.ScheduleLimitDays,
			},
			&cli.StringFlag{
				Name:        "aes-key",
				Usage:       "AES key",
				EnvVars:     []string{"AES_KEY"},
				Destination: &config.AESKey,
			},
			&cli.IntFlag{
				Name:        "sms-provider-batch-limit",
				Usage:       "SMS provider batch limit",
				EnvVars:     []string{"SMS_PROVIDER_API_BATCH_LIMIT"},
				Value:       1000,
				Destination: &config.SmsProviderBatchLimit,
			},
			&cli.StringFlag{
				Name:        "sms-provider-api-token",
				Usage:       "SMS provider API Token",
				EnvVars:     []string{"SMS_PROVIDER_API_TOKEN"},
				Destination: &config.SmsProviderToken,
			},
			&cli.IntFlag{
				Name:        "mail-provider-api-batch-limit",
				Usage:       "Mail provider API batch limit",
				EnvVars:     []string{"MAIL_PROVIDER_API_BATCH_LIMIT"},
				Value:       1000,
				Destination: &config.MailProviderBatchLimit,
			},
			&cli.StringFlag{
				Name:        "mail-provider-api-token",
				Usage:       "Mail provider API Token",
				EnvVars:     []string{"MAIL_PROVIDER_API_TOKEN"},
				Destination: &config.MailProviderToken,
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
			&cli.StringFlag{
				Name:        "job-name",
				Usage:       "job name",
				EnvVars:     []string{"JOB_NAME"},
				Destination: &config.JobName,
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})

	go func() {
		fx.New(
			fx.Supply(&config),
			fx.Provide(
				component.NewOtlpConn,
				component.NewTracerProvider,
				component.NewMeterProvider,
				component.NewSnowflake,
				component.NewAesGcm,
				component.NewDb,
				component.NewValidator,
				component.NewProducer,
				fx.Annotate(
					component.NewGrpcServer,
					fx.ParamTags("", "", `group:"grpcServices"`),
				),
				fx.Annotate(
					service.NewNotifyService,
				),
				fx.Annotate(
					job.NewRunner,
					fx.ParamTags(`group:"jobs"`),
				),
				AsJob(job.NewDispatchScheduledMessagesJob),
			),
			fx.Invoke(
				func(*tracesdk.TracerProvider) {},
				func(*metricssdk.MeterProvider) {},
				func(*grpc.Server) {},
				func(*gorm.DB) {},
				func(factory *job.Runner) {
					jobName := config.JobName
					job, err := factory.GetJob(jobName)
					if err != nil {
						log.Fatalf("%v error: %v", jobName, err)
					}
					if err := job.Execute(ctx); err != nil {
						log.Fatalf("%v failed: %v", jobName, err)
					}
					log.Infof("%v completed successfully.", jobName)
					close(done) // Signal completion
				},
			),
		).Run()
	}()

	select {
	case <-ctx.Done():
		log.Info("Context canceled, shutting down gracefully...")
	case <-done:
		log.Info("Job completed successfully, shutting down...")
	}
	return nil
}

func AsGrpcService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(component.GrpcService)),
		fx.ResultTags(`group:"grpcServices"`),
	)
}

func AsJob(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(job.Job)),
		fx.ResultTags(`group:"jobs"`),
	)
}
