package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/dtan4/grpc-chat/backend/api/chat/v1"
)

const (
	serverAddr = "localhost:50051"
)

func realMain(args []string, logger *zap.Logger) error {
	opts := []grpc.DialOption{}

	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	logger.Info("client is dialing", zap.String("addr", serverAddr))

	client := chatv1.NewChatServiceClient(conn)

	ctx := context.Background()

	stream, err := client.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new stream: %w", err)
	}
	defer stream.CloseSend()

	waitc := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		close(waitc)
	}()

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				logger.Error("failed to receive response", zap.Error(err))
				close(waitc)
				return
			}

			logger.Info("received message", zap.Any("message", in))
		}
	}()

	username := strconv.FormatInt(time.Now().Unix(), 10)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			msg := scanner.Text()

			if err := stream.Send(&chatv1.StreamRequest{
				Username:  username,
				Message:   msg,
				Timestamp: timestamppb.Now(),
			}); err != nil {
				logger.Error("failed to send message", zap.Error(err))
			}
		}
	}()

	<-waitc

	return nil
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	if err := realMain(os.Args, logger); err != nil {
		logger.Error("failed to run client", zap.Error(err))
		os.Exit(1)
	}
}
