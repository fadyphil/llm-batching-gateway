package scheduler

import (
	"context"

	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

type WorkerDispatcher interface {
	RunBatch(ctx context.Context, workerID string, items []*sv1.EnqueueRequest) (<-chan *wv1.RunBatchResponse, <-chan error, error)
}
