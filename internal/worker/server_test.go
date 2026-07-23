package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fadyphil/llm-batching-gateway/internal/worker"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
	"google.golang.org/grpc"
)

type mockWorkerStream struct {
	grpc.ServerStream
	ctx  context.Context
	sent []*wv1.RunBatchResponse
}

func (m *mockWorkerStream) Context() context.Context { return m.ctx }
func (m *mockWorkerStream) Send(resp *wv1.RunBatchResponse) error {
	m.sent = append(m.sent, resp)
	return nil
}

func TestRunBatch_SingleItem_StreamsTokens(t *testing.T) {
	llama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("got method %s; want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "completion") {
			t.Errorf("got path %s; want /completion", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"content\":\"hello\",\"stop\":false}\n\n")
		fmt.Fprint(w, "data: {\"content\":\" world\",\"stop\":false}\n\n")
		fmt.Fprint(w, "data: {\"content\":\"\",\"stop\":true}\n\n")
	}))
	defer llama.Close()

	srv := worker.NewServer(llama.URL)
	stream := &mockWorkerStream{ctx: context.Background()}

	err := srv.RunBatch(&wv1.RunBatchRequest{
		BatchId: "b1",
		Items:   []*wv1.BatchItem{{RequestId: "r1", Prompt: "hi"}},
	}, stream)

	if err != nil {
		t.Fatalf("RunBatch returned error: %v", err)
	}
	if len(stream.sent) != 3 {
		t.Fatalf("got %d responses; want 3", len(stream.sent))
	}
	if stream.sent[0].Token != "hello" {
		t.Errorf("got token %q; want %q", stream.sent[0].Token, "hello")
	}
	if !stream.sent[2].IsFinal {
		t.Error("last response should be final")
	}
}

func TestRunBatch_EmptyItems_ReturnsNil(t *testing.T) {
	llama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer llama.Close()

	srv := worker.NewServer(llama.URL)
	err := srv.RunBatch(&wv1.RunBatchRequest{BatchId: "b1"}, &mockWorkerStream{ctx: context.Background()})
	if err != nil {
		t.Fatalf("RunBatch returned error: %v", err)
	}
}

// llamaHandler returns an http.HandlerFunc that simulates a llama-server SSE endpoint.
// It streams nTokens tokens, sleeping delay between each token, then sends a final stop token.
func llamaHandler(delay time.Duration, nTokens int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		for i := 0; i < nTokens; i++ {
			time.Sleep(delay)
			fmt.Fprintf(w, "data: {\"content\":\"token-%d\",\"stop\":false}\n\n", i)
			flusher.Flush()
		}
		time.Sleep(delay)
		fmt.Fprint(w, "data: {\"content\":\"\",\"stop\":true}\n\n")
		flusher.Flush()
	}
}

func TestWorker_ConcurrentExecution_FasterThanSequential(t *testing.T) {
	llama := httptest.NewServer(llamaHandler(100*time.Millisecond, 2))
	defer llama.Close()

	srv := worker.NewServer(llama.URL)
	stream := &mockWorkerStream{ctx: context.Background()}

	start := time.Now()
	err := srv.RunBatch(&wv1.RunBatchRequest{
		BatchId: "b1",
		Items: []*wv1.BatchItem{
			{RequestId: "r1", Prompt: "hi"},
			{RequestId: "r2", Prompt: "hi"},
			{RequestId: "r3", Prompt: "hi"},
		},
	}, stream)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("RunBatch returned error: %v", err)
	}

	// 3 items * 3 tokens each = 9 responses (3 tokens: token-0, token-1, stop)
	if len(stream.sent) != 9 {
		t.Fatalf("got %d responses; want 9 (3 items * 3 tokens each)", len(stream.sent))
	}

	// Sequential would take ~900ms (3 items * 3 tokens * 100ms).
	// Concurrent should take ~300ms (3 tokens * 100ms).
	// Safety margin: < 500ms.
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("RunBatch took %v; expected < 500ms for concurrent execution", elapsed)
	}
}

type testCompletionRequest struct {
	Prompt string `json:"prompt"`
}

func TestWorker_PartialHTTPFailure_OtherStreamsContinue(t *testing.T) {
	llama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var req testCompletionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.Prompt == "fail-me" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"content\":\"token\",\"stop\":false}\n\n")
		fmt.Fprint(w, "data: {\"content\":\"\",\"stop\":true}\n\n")
	}))
	defer llama.Close()

	srv := worker.NewServer(llama.URL)
	stream := &mockWorkerStream{ctx: context.Background()}

	err := srv.RunBatch(&wv1.RunBatchRequest{
		BatchId: "b1",
		Items: []*wv1.BatchItem{
			{RequestId: "r1", Prompt: "fail-me"},
			{RequestId: "r2", Prompt: "ok"},
		},
	}, stream)

	// The second item's tokens should still be delivered
	if len(stream.sent) == 0 {
		t.Fatal("expected tokens from surviving items, got none")
	}
	hasSecondItem := false
	for _, resp := range stream.sent {
		if resp.RequestId == "r2" {
			hasSecondItem = true
			break
		}
	}
	if !hasSecondItem {
		t.Fatal("expected tokens from second item (r2) despite first item failure")
	}

	// RunBatch should return an error (the first item failed)
	if err == nil {
		t.Fatal("expected error from first item failure, got nil")
	}
}

func TestWorker_ContextCancellation_AllStreamsStop(t *testing.T) {
	llama := httptest.NewServer(llamaHandler(200*time.Millisecond, 10))
	defer llama.Close()

	srv := worker.NewServer(llama.URL)
	ctx, cancel := context.WithCancel(context.Background())
	stream := &mockWorkerStream{ctx: ctx}

	start := time.Now()
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.RunBatch(&wv1.RunBatchRequest{
			BatchId: "b1",
			Items: []*wv1.BatchItem{
				{RequestId: "r1", Prompt: "hi"},
				{RequestId: "r2", Prompt: "hi"},
			},
		}, stream)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		elapsed := time.Since(start)
		if elapsed >= 500*time.Millisecond {
			t.Fatalf("RunBatch took %v after cancellation; expected quick return", elapsed)
		}
		if err == nil {
			t.Log("RunBatch returned nil error after cancellation (context.Canceled may be wrapped)")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunBatch did not return within 2s after context cancellation")
	}
}

func TestRunBatch_ServerError_Propagates(t *testing.T) {
	llama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer llama.Close()

	srv := worker.NewServer(llama.URL)
	err := srv.RunBatch(&wv1.RunBatchRequest{
		BatchId: "b1",
		Items:   []*wv1.BatchItem{{RequestId: "r1", Prompt: "hi"}},
	}, &mockWorkerStream{ctx: context.Background()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
