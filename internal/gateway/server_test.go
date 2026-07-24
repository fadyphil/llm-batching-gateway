package gateway_test

import (
	"context"
	"io"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/fadyphil/llm-batching-gateway/internal/gateway"
	commonv1 "github.com/fadyphil/llm-batching-gateway/proto/go/common/v1"
	gv1 "github.com/fadyphil/llm-batching-gateway/proto/go/gateway/v1"
	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
)

type mockSchedulerStream struct {
	grpc.ClientStream
	chunks []*sv1.EnqueueResponse
	pos    int
}

func (m *mockSchedulerStream) Recv() (*sv1.EnqueueResponse, error) {
	if m.pos >= len(m.chunks) {
		return nil, io.EOF
	}
	m.pos++
	return m.chunks[m.pos-1], nil
}

type mockSchedulerClient struct {
	schedulerClientFunc func(context.Context, *sv1.EnqueueRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[sv1.EnqueueResponse], error)
}

func (m *mockSchedulerClient) Enqueue(ctx context.Context, req *sv1.EnqueueRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[sv1.EnqueueResponse], error) {
	return m.schedulerClientFunc(ctx, req)
}

type mockGatewayStream struct {
	grpc.ServerStream
	ctx  context.Context
	sent []*gv1.CompleteResponse
}

func newMockGatewayStream(ctx context.Context) *mockGatewayStream {
	return &mockGatewayStream{ctx: ctx}
}

func (m *mockGatewayStream) Context() context.Context { return m.ctx }

func (m *mockGatewayStream) Send(resp *gv1.CompleteResponse) error {
	m.sent = append(m.sent, resp)
	return nil
}

func TestComplete_NoMetadata_ReturnsUnauthenticated(t *testing.T) {
	ctx := context.Background()
	srv := gateway.NewServer(nil, "secret")

	err := srv.Complete(&gv1.CompleteRequest{}, newMockGatewayStream(ctx))

	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("got code %v; want %v", status.Code(err), codes.Unauthenticated)
	}
}

func TestComplete_WrongToken_ReturnsUnauthenticated(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "wrong"))
	srv := gateway.NewServer(nil, "secret")

	err := srv.Complete(&gv1.CompleteRequest{}, newMockGatewayStream(ctx))

	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("got code %v; want %v", status.Code(err), codes.Unauthenticated)
	}
}

func TestComplete_ValidToken_StreamsTokensFromScheduler(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "secret"))

	scheduler := &mockSchedulerClient{
		schedulerClientFunc: func(ctx context.Context, req *sv1.EnqueueRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[sv1.EnqueueResponse], error) {
			if req.Model != "test-model" {
				t.Errorf("got model %q; want %q", req.Model, "test-model")
			}
			return &mockSchedulerStream{
				chunks: []*sv1.EnqueueResponse{
					{RequestId: "r1", Token: "hello", IsFinal: false},
					{RequestId: "r1", Token: " world", IsFinal: false},
					{RequestId: "r1", Token: "", IsFinal: true, FinishReason: "stop"},
				},
			}, nil
		},
	}

	srv := gateway.NewServer(scheduler, "secret")
	stream := newMockGatewayStream(ctx)

	err := srv.Complete(&gv1.CompleteRequest{
		SessionId: "s1",
		Prompt:    "test",
		Model:     "test-model",
		Priority:  commonv1.Priority_PRIORITY_INTERACTIVE,
	}, stream)

	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if len(stream.sent) != 3 {
		t.Errorf("got %d responses; want 3", len(stream.sent))
	}
	if stream.sent[len(stream.sent)-1].IsFinal != true {
		t.Error("last response should be final")
	}
}

func TestComplete_SchedulerError_PropagatesError(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "secret"))

	scheduler := &mockSchedulerClient{
		schedulerClientFunc: func(ctx context.Context, req *sv1.EnqueueRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[sv1.EnqueueResponse], error) {
			return nil, status.Error(codes.ResourceExhausted, "scheduler full")
		},
	}

	srv := gateway.NewServer(scheduler, "secret")

	err := srv.Complete(&gv1.CompleteRequest{}, newMockGatewayStream(ctx))

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
