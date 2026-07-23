package schedulerv1

import commonv1 "github.com/fadyphil/llm-batching-gateway/proto/go/common/v1"

type EnqueueRequest struct {
	RequestId  string
	SessionId  string
	Prompt     string
	Model      string
	Priority   commonv1.Priority
	TokenCount int32
}

type EnqueueResponse struct {
	RequestId    string
	Token        string
	IsFinal      bool
	FinishReason commonv1.FinishReason
	Error        string
}
