package scheduler

import (
	"context"
	"testing"
	"time"

	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
	"pgregory.net/rapid"
)

func TestProperty_Backpressure_NoLostRequestsUnderLoad(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ingressCap := rapid.IntRange(3, 20).Draw(t, "ingressCapacity")
		reqCount := rapid.IntRange(1, ingressCap-1).Draw(t, "reqCount")
		requestIDs := rapid.SliceOfN(GenRequestID(), reqCount, reqCount).Draw(t, "requestIDs")

		d := &recordingDispatcher{
			runBatchFunc: func(_ context.Context, _ string, items []*sv1.EnqueueRequest) (<-chan *wv1.RunBatchResponse, <-chan error, error) {
				respCh := make(chan *wv1.RunBatchResponse, 2)
				respCh <- &wv1.RunBatchResponse{
					RequestId: items[0].RequestId,
					Token:     "ok",
					IsFinal:   true,
				}
				close(respCh)
				errCh := make(chan error)
				close(errCh)
				return respCh, errCh, nil
			},
		}

		s := NewScheduler(Config{
			IngressCapacity: ingressCap,
			MaxBatchSize:    1,
			BatchWindow:     time.Millisecond,
			SessionTTL:      time.Minute,
			HeartbeatLimit:  time.Minute,
		}, d, NewFakeTimeSource(time.Now()))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go s.Run(ctx)

		type result struct {
			id     string
			err    error
			chunks int
		}
		resultCh := make(chan result, reqCount)

		for _, id := range requestIDs {
			id := id
			go func() {
				stream := &mockStream{}
				err := s.Enqueue(&sv1.EnqueueRequest{RequestId: id}, stream)
				resultCh <- result{id: id, err: err, chunks: len(stream.chunks)}
			}()
		}

		completed := 0
		for completed < reqCount {
			r := <-resultCh
			if r.err != nil {
				t.Fatalf("request %s returned unexpected error: %v", r.id, r.err)
			}
			if r.chunks == 0 {
				t.Fatalf("request %s received zero chunks", r.id)
			}
			completed++
		}

		if len(d.calls) != reqCount {
			t.Fatalf("expected %d dispatcher calls, got %d", reqCount, len(d.calls))
		}

		for _, call := range d.calls {
			if len(call.Items) != 1 {
				t.Fatalf("expected 1 item per dispatcher call, got %d", len(call.Items))
			}
		}
	})
}
