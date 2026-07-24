package scheduler_test

import (
	"context"
	"io"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fadyphil/llm-batching-gateway/internal/scheduler"
	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

type mockWorkerStream struct {
	grpc.ClientStream
	chunks []*wv1.RunBatchResponse
	pos    int
}

func (m *mockWorkerStream) Recv() (*wv1.RunBatchResponse, error) {
	if m.pos >= len(m.chunks) {
		return nil, io.EOF
	}
	m.pos++
	return m.chunks[m.pos-1], nil
}

type errWorkerStream struct {
	grpc.ClientStream
	err error
}

func (e *errWorkerStream) Recv() (*wv1.RunBatchResponse, error) {
	return nil, e.err
}

type mockWorkerClient struct {
	runBatchFunc func(context.Context, *wv1.RunBatchRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[wv1.RunBatchResponse], error)
}

func (m *mockWorkerClient) RunBatch(ctx context.Context, req *wv1.RunBatchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[wv1.RunBatchResponse], error) {
	return m.runBatchFunc(ctx, req)
}

type mockSchedulerStream struct {
	grpc.ServerStream
	ctx  context.Context
	sent []*sv1.EnqueueResponse
}

func (m *mockSchedulerStream) Context() context.Context { return m.ctx }
func (m *mockSchedulerStream) Send(resp *sv1.EnqueueResponse) error {
	m.sent = append(m.sent, resp)
	return nil
}

func TestEnqueue_ValidRequest_StreamsTokens(t *testing.T) {
	worker := &mockWorkerClient{
		runBatchFunc: func(ctx context.Context, req *wv1.RunBatchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[wv1.RunBatchResponse], error) {
			return &mockWorkerStream{
				chunks: []*wv1.RunBatchResponse{
					{RequestId: "r1", Token: "hello"},
					{RequestId: "r1", Token: " world"},
					{RequestId: "r1", IsFinal: true},
				},
			}, nil
		},
	}

	srv := scheduler.NewServer(worker)
	stream := &mockSchedulerStream{ctx: context.Background()}

	err := srv.Enqueue(&sv1.EnqueueRequest{RequestId: "r1", Prompt: "hi"}, stream)
	if err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}
	if len(stream.sent) != 3 {
		t.Fatalf("got %d responses; want 3", len(stream.sent))
	}
	if !stream.sent[2].IsFinal {
		t.Error("last response should be final")
	}
}

func TestEnqueue_WorkerError_Propagates(t *testing.T) {
	worker := &mockWorkerClient{
		runBatchFunc: func(ctx context.Context, req *wv1.RunBatchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[wv1.RunBatchResponse], error) {
			return nil, status.Error(codes.Unavailable, "worker offline")
		},
	}

	srv := scheduler.NewServer(worker)
	err := srv.Enqueue(&sv1.EnqueueRequest{RequestId: "r1", Prompt: "hi"}, &mockSchedulerStream{ctx: context.Background()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEnqueue_WorkerStreamRecvError_Propagates(t *testing.T) {
	worker := &mockWorkerClient{
		runBatchFunc: func(ctx context.Context, req *wv1.RunBatchRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[wv1.RunBatchResponse], error) {
			return &errWorkerStream{err: io.ErrUnexpectedEOF}, nil
		},
	}

	srv := scheduler.NewServer(worker)
	err := srv.Enqueue(&sv1.EnqueueRequest{RequestId: "r1", Prompt: "hi"}, &mockSchedulerStream{ctx: context.Background()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
