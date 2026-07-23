package commonv1

type Priority int32

const (
	Priority_PRIORITY_UNSPECIFIED Priority = 0
	Priority_INTERACTIVE          Priority = 1
	Priority_BACKGROUND           Priority = 2
)

type FinishReason int32

const (
	FinishReason_FINISH_REASON_UNSPECIFIED FinishReason = 0
	FinishReason_STOP                      FinishReason = 1
	FinishReason_LENGTH                    FinishReason = 2
	FinishReason_ERROR                     FinishReason = 3
)

type CompletionChunk struct {
	RequestId    string
	Token        string
	IsFinal      bool
	FinishReason FinishReason
}
