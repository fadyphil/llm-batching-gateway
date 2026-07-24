package scheduler

import (
	"context"
	"time"

	commonv1 "github.com/fadyphil/llm-batching-gateway/proto/go/common/v1"
	sv1 "github.com/fadyphil/llm-batching-gateway/proto/go/scheduler/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pendingRequest struct {
	req        *sv1.EnqueueRequest
	resultChan chan *commonv1.CompletionChunk
	enqueuedAt time.Time
}

type EnqueueServer interface {
	Send(*sv1.EnqueueResponse) error
}

type Scheduler struct {
	incoming   chan *pendingRequest
	dispatcher WorkerDispatcher
	timeSource TimeSource
	cfg        Config
}

func NewScheduler(cfg Config, dispatcher WorkerDispatcher, timeSource TimeSource) *Scheduler {
	return &Scheduler{
		incoming:   make(chan *pendingRequest, cfg.IngressCapacity),
		dispatcher: dispatcher,
		timeSource: timeSource,
		cfg:        cfg,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case pr := <-s.incoming:
			s.dispatch(ctx, pr)
		}
	}
}

func (s *Scheduler) dispatch(ctx context.Context, pr *pendingRequest) {
	defer close(pr.resultChan)

	respCh, errCh, err := s.dispatcher.RunBatch(ctx, "", []*sv1.EnqueueRequest{pr.req})
	if err != nil {
		return
	}

	for respCh != nil || errCh != nil {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-respCh:
			if !ok {
				respCh = nil
				continue
			}
			chunk := &commonv1.CompletionChunk{
				RequestId: resp.RequestId,
				Token:     resp.Token,
				IsFinal:   resp.IsFinal,
			}
			pr.resultChan <- chunk
			if resp.IsFinal {
				return
			}
		case _, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			return
		}
	}
}

func (s *Scheduler) Enqueue(req *sv1.EnqueueRequest, stream EnqueueServer) error {
	pr := &pendingRequest{
		req:        req,
		resultChan: make(chan *commonv1.CompletionChunk, 10),
		enqueuedAt: s.timeSource.Now(),
	}

	select {
	case s.incoming <- pr:
	default:
		return status.Error(codes.ResourceExhausted, "scheduler saturated")
	}

	for chunk := range pr.resultChan {
		err := stream.Send(&sv1.EnqueueResponse{
			RequestId:    chunk.RequestId,
			Token:        chunk.Token,
			IsFinal:      chunk.IsFinal,
			FinishReason: chunk.FinishReason,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
