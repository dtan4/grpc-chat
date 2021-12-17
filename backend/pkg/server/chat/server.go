package chat

import (
	"io"

	chatv1 "github.com/dtan4/grpc-chat/backend/api/chat/v1"
	"go.uber.org/zap"
)

type Server struct {
	chatv1.UnimplementedChatServiceServer

	logger *zap.Logger
}

var _ chatv1.ChatServiceServer = (*Server)(nil)

func New(logger *zap.Logger) *Server {
	return &Server{
		logger: logger,
	}
}

func (s *Server) Stream(stream chatv1.ChatService_StreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			s.logger.Info("streaming finished")

			return nil
		}

		if err != nil {
			s.logger.Error("failed to receive request", zap.Error(err))

			return err
		}

		s.logger.Info("received message", zap.Any("message", in))

		if err := stream.Send(&chatv1.StreamResponse{
			Username:  in.Username,
			Message:   in.Message,
			Timestamp: in.Timestamp,
		}); err != nil {
			s.logger.Error("failed to send message", zap.Error(err))

			return err
		}
	}
}
