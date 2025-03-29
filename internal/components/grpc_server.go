package component

import (
	"context"
	"fmt"
	"net"
	shared "notify-service/internal"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func NewGrpcServer(
	lc fx.Lifecycle,
	config *shared.Config,
	grpcServices []GrpcService,
) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(traceIDInterceptor),
	)
	for _, svc := range grpcServices {
		svc.Register(grpcServer)
	}
	reflection.Register(grpcServer)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
				if err != nil {
					log.WithError(err).Fatal("gRPC server failed to listen")
				}
				log.Infof("grpc server listening at %v", lis.Addr())

				err = grpcServer.Serve(lis)
				if err != nil {
					log.WithError(err).Fatal("Error starting grpc server")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
	})
	return grpcServer
}

func traceIDInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	span := trace.SpanFromContext(ctx)

	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()

		md := metadata.Pairs(
			"Trace-ID", traceID,
			"Span-ID", spanID,
		)
		grpc.SendHeader(ctx, md)
	}

	return handler(ctx, req)
}

type GrpcService interface {
	Register(server *grpc.Server)
}
