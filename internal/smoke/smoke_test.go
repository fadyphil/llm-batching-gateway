package smoke_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/fadyphil/llm-batching-gateway/internal/gateway"
	"github.com/fadyphil/llm-batching-gateway/internal/scheduler"
	"github.com/fadyphil/llm-batching-gateway/internal/worker"
	commonv1 "github.com/fadyphil/llm-batching-gateway/proto/go/common/v1"
	gv1 "github.com/fadyphil/llm-batching-gateway/proto/go/gateway/v1"
	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

func TestEndToEnd(t *testing.T) {
	llama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"content\":\"Hello\",\"stop\":false}\n\n")
		fmt.Fprint(w, "data: {\"content\":\" world\",\"stop\":false}\n\n")
		fmt.Fprint(w, "data: {\"content\":\"\",\"stop\":true}\n\n")
	}))
	defer llama.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerAddr := serveWorker(ctx, t, llama.URL)
	schedulerAddr := serveScheduler(ctx, t, workerAddr)
	gatewayAddr := serveGateway(ctx, t, schedulerAddr, "test-token")

	time.Sleep(50 * time.Millisecond)

	conn, err := grpc.NewClient(gatewayAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("connect gateway: %v", err)
	}
	defer conn.Close()

	client := gv1.NewCompletionServiceClient(conn)
	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "test-token"))

	stream, err := client.Complete(authCtx, &gv1.CompleteRequest{
		SessionId: "s1",
		Prompt:    "hi",
		Model:     "default",
		Priority:  commonv1.Priority_PRIORITY_INTERACTIVE,
	})
	if err != nil {
		t.Fatalf("Complete RPC: %v", err)
	}

	var tokens []string
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		tokens = append(tokens, chunk.Token)
		if chunk.IsFinal {
			break
		}
	}

	if len(tokens) != 3 {
		t.Fatalf("got %d tokens; want 3: %v", len(tokens), tokens)
	}
	t.Logf("E2E OK — %d tokens: %v", len(tokens), tokens)
}

func listenTCP(t *testing.T) net.Listener {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Helper()
		t.Fatalf("listen: %v", err)
	}
	return lis
}

func serveGateway(ctx context.Context, t *testing.T, schedulerAddr, authToken string) string {
	lis := listenTCP(t)

	conn, err := grpc.NewClient(schedulerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("gateway dial scheduler: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	schedClient := sv1.NewSchedulerServiceClient(conn)
	srv := gateway.NewServer(schedClient, authToken)

	gs := grpc.NewServer()
	gv1.RegisterCompletionServiceServer(gs, srv)

	go func() {
		<-ctx.Done()
		gs.GracefulStop()
	}()
	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("gateway serve: %v", err)
		}
	}()

	return lis.Addr().String()
}

func serveScheduler(ctx context.Context, t *testing.T, workerAddr string) string {
	lis := listenTCP(t)

	conn, err := grpc.NewClient(workerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("scheduler dial worker: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	workerClient := wv1.NewWorkerServiceClient(conn)
	srv := scheduler.NewServer(workerClient)

	gs := grpc.NewServer()
	sv1.RegisterSchedulerServiceServer(gs, srv)

	go func() {
		<-ctx.Done()
		gs.GracefulStop()
	}()
	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("scheduler serve: %v", err)
		}
	}()

	return lis.Addr().String()
}

func serveWorker(ctx context.Context, t *testing.T, llamaURL string) string {
	lis := listenTCP(t)

	srv := worker.NewServer(llamaURL)

	gs := grpc.NewServer()
	wv1.RegisterWorkerServiceServer(gs, srv)

	go func() {
		<-ctx.Done()
		gs.GracefulStop()
	}()
	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("worker serve: %v", err)
		}
	}()

	return lis.Addr().String()
}
