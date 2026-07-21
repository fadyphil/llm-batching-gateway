package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/fadyphil/llm-batching-gateway/internal/gateway"
	gv1 "github.com/fadyphil/llm-batching-gateway/proto/go/gateway/v1"
	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
)

func main() {
	port := env("GATEWAY_PORT", "9000")
	schedulerAddr := env("SCHEDULER_ADDR", "localhost:9001")
	authToken := env("AUTH_TOKEN", "dev-token")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("listen port %s: %v", port, err)
	}

	conn, err := grpc.NewClient(schedulerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect to scheduler %s: %v", schedulerAddr, err)
	}

	scheduler := sv1.NewSchedulerServiceClient(conn)
	srv := gateway.NewServer(scheduler, authToken)

	gs := grpc.NewServer()
	gv1.RegisterCompletionServiceServer(gs, srv)

	log.Printf("Gateway listening on :%s, scheduler at %s", port, schedulerAddr)
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
