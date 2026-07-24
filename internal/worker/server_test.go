package worker_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	// Fake llama-server that streams 3 tokens via SSE
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
