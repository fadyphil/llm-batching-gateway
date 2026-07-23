package worker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/grpc"

	wv1 "github.com/fadyphil/llm-batching-gateway/proto/go/worker/v1"
)

type Server struct {
	wv1.UnimplementedWorkerServiceServer
	llamaURL string
	client   *http.Client
}

func NewServer(llamaURL string) *Server {
	return &Server{
		llamaURL: llamaURL,
		client:   http.DefaultClient,
	}
}

type completionRequest struct {
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type completionChunk struct {
	Content string `json:"content"`
	Stop    bool   `json:"stop"`
}

func (s *Server) RunBatch(req *wv1.RunBatchRequest, stream grpc.ServerStreamingServer[wv1.RunBatchResponse]) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var wg sync.WaitGroup
	demux := make(chan *wv1.RunBatchResponse)
	errCh := make(chan error, len(req.Items))

	for _, item := range req.Items {
		item := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.runItem(ctx, item, demux); err != nil {
				errCh <- err
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for resp := range demux {
			if err := stream.Send(resp); err != nil {
				cancel()
				return
			}
		}
	}()

	wg.Wait()
	close(demux)
	<-done

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (s *Server) runItem(ctx context.Context, item *wv1.BatchItem, demux chan<- *wv1.RunBatchResponse) error {
	body := &completionRequest{Prompt: item.Prompt, Stream: true}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.llamaURL+"/completion", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("llama-server returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")

		var chunk completionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return fmt.Errorf("parse SSE: %w", err)
		}

		select {
		case demux <- &wv1.RunBatchResponse{
			RequestId: item.RequestId,
			Token:     chunk.Content,
			IsFinal:   chunk.Stop,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}

		if chunk.Stop {
			break
		}
	}

	return scanner.Err()
}
