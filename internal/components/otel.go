package component

import (
	"context"

	shared "notify-service/internal"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	metricssdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewOtlpConn(
	lc fx.Lifecycle,
	config *shared.Config,
) *grpc.ClientConn {
	conn, err := grpc.NewClient(config.OtlpEndpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.WithError(err).Fatalf("failed to create gRPC client connection to %s", config.OtlpEndpoint)
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})
	return conn
}

func NewTracerProvider(
	lc fx.Lifecycle,
	conn *grpc.ClientConn,
	config *shared.Config,
) *tracesdk.TracerProvider {
	// Traces
	traceExporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.WithError(err).Error("failed to initialize trace exporter")
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(traceExporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.OtlpServiceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	})
	return tp
}

func NewMeterProvider(
	lc fx.Lifecycle,
	conn *grpc.ClientConn,
	config *shared.Config,
) *metricssdk.MeterProvider {
	metricExporter, err := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		log.WithError(err).Error("failed to initialize metric exporter")
	}
	mp := metricssdk.NewMeterProvider(
		metricssdk.WithReader(metricssdk.NewPeriodicReader(metricExporter)),
		metricssdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.OtlpServiceName),
		)),
	)
	otel.SetMeterProvider(mp)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return mp.Shutdown(ctx)
		},
	})
	return nil
}
