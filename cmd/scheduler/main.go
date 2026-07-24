package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/fadyphil/llm-batching-gateway/internal/scheduler"
	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

func main() {
	port := env("SCHEDULER_PORT", "9001")
	workerAddr := env("WORKER_ADDR", "localhost:9002")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen port %s: %v", port, err)
	}

	conn, err := grpc.NewClient(workerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect to worker %s: %v", workerAddr, err)
	}

	worker := wv1.NewWorkerServiceClient(conn)
	srv := scheduler.NewServer(worker)

	gs := grpc.NewServer()
	sv1.RegisterSchedulerServiceServer(gs, srv)

	log.Printf("Scheduler listening on :%s, worker at %s", port, workerAddr)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
