package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	if err := realMain(os.Args, logger); err != nil {
		logger.Error("failed to run server", zap.Error(err))
		os.Exit(1)
	}
}
