package scheduler

import (
	"context"
	"testing"

	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

type recordingDispatcher struct {
	runBatchFunc func(ctx context.Context, workerID string, items []*sv1.EnqueueRequest) (<-chan *wv1.RunBatchResponse, <-chan error, error)
	calls        []dispatchCall
}

type dispatchCall struct {
	WorkerID string
	Items    []*sv1.EnqueueRequest
}

func (d *recordingDispatcher) RunBatch(ctx context.Context, workerID string, items []*sv1.EnqueueRequest) (<-chan *wv1.RunBatchResponse, <-chan error, error) {
	d.calls = append(d.calls, dispatchCall{WorkerID: workerID, Items: items})
	if d.runBatchFunc != nil {
		return d.runBatchFunc(ctx, workerID, items)
	}
	return nil, nil, nil
}

func TestRecordingDispatcher_RecordsCall(t *testing.T) {
	d := &recordingDispatcher{
		runBatchFunc: func(ctx context.Context, workerID string, items []*sv1.EnqueueRequest) (<-chan *wv1.RunBatchResponse, <-chan error, error) {
			return nil, nil, nil
		},
	}

	items := []*sv1.EnqueueRequest{
		{RequestId: "req-1"},
		{RequestId: "req-2"},
	}

	_, _, err := d.RunBatch(context.Background(), "worker-1", items)
	if err != nil {
		t.Fatalf("RunBatch returned error: %v", err)
	}

	if len(d.calls) != 1 {
		t.Fatalf("expected 1 recorded call, got %d", len(d.calls))
	}
	if d.calls[0].WorkerID != "worker-1" {
		t.Fatalf("expected recorded WorkerID 'worker-1', got %q", d.calls[0].WorkerID)
	}
	if len(d.calls[0].Items) != 2 {
		t.Fatalf("expected 2 recorded items, got %d", len(d.calls[0].Items))
	}
	if d.calls[0].Items[0].RequestId != "req-1" || d.calls[0].Items[1].RequestId != "req-2" {
		t.Fatal("recorded items do not match input")
	}
}

func TestRecordingDispatcher_CumulativeCalls(t *testing.T) {
	d := &recordingDispatcher{}

	_, _, _ = d.RunBatch(context.Background(), "worker-1", []*sv1.EnqueueRequest{{RequestId: "req-a"}})
	_, _, _ = d.RunBatch(context.Background(), "worker-2", []*sv1.EnqueueRequest{{RequestId: "req-b"}})

	if len(d.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(d.calls))
	}
}

func TestRecordingDispatcher_SatisfiesInterface(t *testing.T) {
	var d WorkerDispatcher = &recordingDispatcher{}
	if d == nil {
		t.Fatal("recordingDispatcher should implement WorkerDispatcher")
	}
}

func TestRecordingDispatcher_DefaultsToNoOp(t *testing.T) {
	d := &recordingDispatcher{}
	items := []*sv1.EnqueueRequest{{RequestId: "req-1"}}
	respCh, errCh, err := d.RunBatch(context.Background(), "w1", items)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if respCh != nil {
		t.Fatal("expected nil response channel with no runBatchFunc")
	}
	if errCh != nil {
		t.Fatal("expected nil error channel with no runBatchFunc")
	}
}
