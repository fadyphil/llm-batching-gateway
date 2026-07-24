package worker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	for _, item := range req.Items {
		if err := s.runSingle(item, stream); err != nil {
			return fmt.Errorf("item %s: %w", item.RequestId, err)
		}
	}
	return nil
}

func (s *Server) runSingle(item *wv1.BatchItem, stream grpc.ServerStreamingServer[wv1.RunBatchResponse]) error {
	body := &completionRequest{Prompt: item.Prompt, Stream: true}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(stream.Context(), http.MethodPost, s.llamaURL+"/completion", bytes.NewReader(data))
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
			return fmt.Errorf("parse SSE chunk: %w", err)
		}

		if err := stream.Send(&wv1.RunBatchResponse{
			RequestId: item.RequestId,
			Token:     chunk.Content,
			IsFinal:   chunk.Stop,
		}); err != nil {
			return fmt.Errorf("send token: %w", err)
		}

		if chunk.Stop {
			break
		}
	}

	return scanner.Err()
}
