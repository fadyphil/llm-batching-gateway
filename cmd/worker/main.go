package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/fadyphil/llm-batching-gateway/internal/worker"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

func main() {
	port := env("WORKER_PORT", "9002")
	llamaURL := env("LLAMA_URL", "http://localhost:8080")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen port %s: %v", port, err)
	}

	srv := worker.NewServer(llamaURL)

	gs := grpc.NewServer()
	wv1.RegisterWorkerServiceServer(gs, srv)

	log.Printf("Worker listening on :%s, llama-server at %s", port, llamaURL)
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
