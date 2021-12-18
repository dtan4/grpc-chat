package chat

import (
	"context"
	"io"
	"sync"

	chatv1 "github.com/dtan4/grpc-chat/backend/api/chat/v1"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Server struct {
	chatv1.UnimplementedChatServiceServer

	logger   *zap.Logger
	chPool   map[string](chan *chatv1.StreamResponse)
	chPoolMu sync.Mutex
}

var _ chatv1.ChatServiceServer = (*Server)(nil)

func New(logger *zap.Logger) *Server {
	return &Server{
		logger: logger,
		chPool: make(map[string]chan *chatv1.StreamResponse),
	}
}

func generateRandomID() string {
	return uuid.NewString()
}

func (s *Server) broadcast(resp *chatv1.StreamResponse) {
	s.chPoolMu.Lock()
	defer s.chPoolMu.Unlock()

	var wg sync.WaitGroup

	for _, ch := range s.chPool {
		wg.Add(1)
		ch := ch

		go func() {
			defer wg.Done()
			ch <- resp
		}()
	}

	wg.Wait()
}

func (s *Server) addChan(id string, ch chan *chatv1.StreamResponse) {
	s.chPoolMu.Lock()
	defer s.chPoolMu.Unlock()

	s.logger.Info("adding new channel", zap.String("id", id))
	s.chPool[id] = ch
}

func (s *Server) deleteChan(id string) {
	s.chPoolMu.Lock()
	defer s.chPoolMu.Unlock()

	s.logger.Info("removing a channel", zap.String("id", id))
	delete(s.chPool, id)
}

func (s *Server) StreamReceive(ctx context.Context, stream chatv1.ChatService_StreamServer, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			in, err := stream.Recv()
			if err == io.EOF {
				s.logger.Info("streaming finished")
				close(errCh)

				return
			}

			if err != nil {
				s.logger.Error("failed to receive request", zap.Error(err))
				errCh <- err

				return
			}

			s.logger.Info("received message", zap.Any("message", in))

			s.broadcast(&chatv1.StreamResponse{
				Username:  in.Username,
				Message:   in.Message,
				Timestamp: in.Timestamp,
			})
		}
	}
}

func (s *Server) StreamSend(ctx context.Context, stream chatv1.ChatService_StreamServer, ch <-chan *chatv1.StreamResponse, errCh chan<- error) {
	for {
		select {
		case resp := <-ch:
			if err := stream.Send(resp); err != nil {
				s.logger.Error("failed to send message", zap.Error(err))
				errCh <- err

				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) Stream(stream chatv1.ChatService_StreamServer) error {
	chID := generateRandomID()
	ch := make(chan *chatv1.StreamResponse, 1)
	s.addChan(chID, ch)
	defer s.deleteChan(chID)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	errCh := make(chan error, 1)

	go s.StreamReceive(ctx, stream, errCh)
	go s.StreamSend(ctx, stream, ch, errCh)

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}
