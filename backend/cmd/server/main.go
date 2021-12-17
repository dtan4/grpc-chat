package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	chatv1 "github.com/dtan4/grpc-chat/backend/api/chat/v1"
	"github.com/dtan4/grpc-chat/backend/pkg/server/chat"
	"google.golang.org/grpc/reflection"
)

const (
	gRPCPort = 50051
)

func realMain(args []string, logger *zap.Logger) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", gRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen :%d", gRPCPort)
	}
	defer l.Close()

	s := grpc.NewServer()

	chatv1.RegisterChatServiceServer(s, chat.New(logger))

	reflection.Register(s)

	logger.Info("server is listening", zap.Any("addr", l.Addr()))

	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		logger.Info("shutting down gracefully")
		s.GracefulStop()

		close(idleConnsClosed)
	}()

	if err := s.Serve(l); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	<-idleConnsClosed

	return nil
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create zap logger: %s", err)
	}
	defer logger.Sync()

	if err := realMain(os.Args, logger); err != nil {
		logger.Error("failed to run server", zap.Error(err))
		os.Exit(1)
	}
}
