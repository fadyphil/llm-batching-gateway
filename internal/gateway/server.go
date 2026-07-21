package gateway

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	commonv1 "github.com/fadyphil/llm-batching-gateway/proto/go/common/v1"
	gv1 "github.com/fadyphil/llm-batching-gateway/proto/go/gateway/v1"
	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
)

var now = time.Now

func nextID() string {
	return fmt.Sprintf("req-%d", now().UnixNano())
}

type schedulerClient interface {
	Enqueue(ctx context.Context, in *sv1.EnqueueRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[sv1.EnqueueResponse], error)
}

type Server struct {
	gv1.UnimplementedCompletionServiceServer
	scheduler schedulerClient
	authToken string
}

func NewServer(scheduler schedulerClient, authToken string) *Server {
	return &Server{scheduler: scheduler, authToken: authToken}
}

func (s *Server) Complete(req *gv1.CompleteRequest, stream grpc.ServerStreamingServer[gv1.CompleteResponse]) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}
	tokens := md.Get("authorization")
	if len(tokens) == 0 || tokens[0] != s.authToken {
		return status.Error(codes.Unauthenticated, "invalid auth token")
	}

	requestID := nextID()

	schedStream, err := s.scheduler.Enqueue(stream.Context(), &sv1.EnqueueRequest{
		RequestId:  requestID,
		SessionId:  req.SessionId,
		Prompt:     req.Prompt,
		Model:      req.Model,
		Priority:   commonv1.Priority(req.Priority),
		TokenCount: 0,
	})
	if err != nil {
		return fmt.Errorf("enqueue: %w", err)
	}

	for {
		chunk, err := schedStream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scheduler stream recv: %w", err)
		}
		if err := stream.Send(&gv1.CompleteResponse{
			RequestId:    chunk.RequestId,
			Token:        chunk.Token,
			IsFinal:      chunk.IsFinal,
			FinishReason: chunk.FinishReason,
		}); err != nil {
			return fmt.Errorf("client stream send: %w", err)
		}
	}
}
