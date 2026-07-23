package scheduler

import (
	"context"
	"testing"
	"time"

	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockStream struct {
	t      *testing.T
	chunks []*sv1.EnqueueResponse
}

func (m *mockStream) Send(resp *sv1.EnqueueResponse) error {
	m.chunks = append(m.chunks, resp)
	return nil
}

func TestEnqueue_BasicRequest_StreamsTokens(t *testing.T) {
	tokenCh := make(chan *wv1.RunBatchResponse, 2)
	tokenCh <- &wv1.RunBatchResponse{RequestId: "req-1", Token: "hel", IsFinal: false}
	tokenCh <- &wv1.RunBatchResponse{RequestId: "req-1", Token: "lo", IsFinal: false}
	tokenCh <- &wv1.RunBatchResponse{RequestId: "req-1", Token: "", IsFinal: true}
	close(tokenCh)
	errCh := make(chan error)
	close(errCh)

	d := &recordingDispatcher{
		runBatchFunc: func(_ context.Context, _ string, _ []*sv1.EnqueueRequest) (<-chan *wv1.RunBatchResponse, <-chan error, error) {
			return tokenCh, errCh, nil
		},
	}

	s := NewScheduler(Config{IngressCapacity: 10, MaxBatchSize: 1, BatchWindow: time.Millisecond, SessionTTL: time.Minute, HeartbeatLimit: time.Minute}, d, NewFakeTimeSource(time.Now()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.Run(ctx)

	stream := &mockStream{t: t}
	err := s.Enqueue(&sv1.EnqueueRequest{RequestId: "req-1"}, stream)
	if err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}

	if len(stream.chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(stream.chunks))
	}
	if stream.chunks[0].Token != "hel" {
		t.Fatalf("expected first token 'hel', got %q", stream.chunks[0].Token)
	}
	if stream.chunks[1].Token != "lo" {
		t.Fatalf("expected second token 'lo', got %q", stream.chunks[1].Token)
	}
	if !stream.chunks[2].IsFinal {
		t.Fatal("expected third chunk IsFinal=true")
	}
}

func TestEnqueue_ChannelFull_ReturnsResourceExhausted(t *testing.T) {
	d := &recordingDispatcher{}

	s := NewScheduler(Config{IngressCapacity: 1, MaxBatchSize: 1, BatchWindow: time.Millisecond, SessionTTL: time.Minute, HeartbeatLimit: time.Minute}, d, NewFakeTimeSource(time.Now()))

	enqueued := make(chan struct{})
	go func() {
		_ = s.Enqueue(&sv1.EnqueueRequest{RequestId: "req-1"}, &mockStream{t: t})
		close(enqueued)
	}()

	select {
	case <-enqueued:
		t.Fatal("first Enqueue returned unexpectedly (should block)")
	case <-time.After(10 * time.Millisecond):
	}

	err := s.Enqueue(&sv1.EnqueueRequest{RequestId: "req-2"}, &mockStream{t: t})
	if err == nil {
		t.Fatal("expected RESOURCE_EXHAUSTED error, got nil")
	}
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected code ResourceExhausted, got %v", status.Code(err))
	}
}
