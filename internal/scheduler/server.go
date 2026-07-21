package scheduler

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"

	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

type workerClient interface {
	RunBatch(ctx context.Context, in *wv1.RunBatchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[wv1.RunBatchResponse], error)
}

type Server struct {
	sv1.UnimplementedSchedulerServiceServer
	worker workerClient
}

func NewServer(worker workerClient) *Server {
	return &Server{worker: worker}
}

func (s *Server) Enqueue(req *sv1.EnqueueRequest, stream grpc.ServerStreamingServer[sv1.EnqueueResponse]) error {
	batchID := fmt.Sprintf("batch-%d", time.Now().UnixNano())

	workerStream, err := s.worker.RunBatch(stream.Context(), &wv1.RunBatchRequest{
		BatchId: batchID,
		Items: []*wv1.BatchItem{
			{RequestId: req.RequestId, Prompt: req.Prompt},
		},
	})
	if err != nil {
		return fmt.Errorf("dispatch to worker: %w", err)
	}

	for {
		chunk, err := workerStream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("worker stream: %w", err)
		}
		if err := stream.Send(&sv1.EnqueueResponse{
			RequestId:    chunk.RequestId,
			Token:        chunk.Token,
			IsFinal:      chunk.IsFinal,
			FinishReason: "",
		}); err != nil {
			return fmt.Errorf("send to gateway: %w", err)
		}
	}
}
