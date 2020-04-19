package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	grouppb "github.com/benkim0414/groupservice/pb"
	"github.com/benkim0414/groupservice/pkg/groupservice"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to build zap logger: %v", err)
	}
	defer logger.Sync()

	lis, err := net.Listen("tcp", fmt.Sprintf(":50051"))
	if err != nil {
		logger.Fatal("failed to listen: %v", zap.Error(err))
	}

	ctx := context.Background()
	serviceAccountFilePath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	svc, err := groupservice.New(ctx, serviceAccountFilePath)
	if err != nil {
		logger.Fatal("failed to initialize the service: %v", zap.Error(err))
	}

	grpc_zap.ReplaceGrpcLoggerV2(logger)
	s := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(logger),
		),
	)
	healthpb.RegisterHealthServer(s, health.NewServer())
	grouppb.RegisterGroupServiceServer(s, svc)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		logger.Fatal("failed to serve: %v", zap.Error(err))
	}
}
