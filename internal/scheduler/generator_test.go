package scheduler

import (
	commonv1 "github.com/fadyphil/llm-batching-gateway/proto/go/common/v1"
	"pgregory.net/rapid"
)

func GenPriority() *rapid.Generator[commonv1.Priority] {
	return rapid.Custom(func(t *rapid.T) commonv1.Priority {
		return rapid.SampledFrom([]commonv1.Priority{
			commonv1.Priority_INTERACTIVE,
			commonv1.Priority_BACKGROUND,
		}).Draw(t, "priority")
	})
}

func GenModel() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-z0-9]+(-[a-z0-9]+)*`)
}

func GenRequestID() *rapid.Generator[string] {
	return rapid.StringMatching(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
}

func GenSessionID() *rapid.Generator[string] {
	return rapid.OneOf(
		rapid.Just(""),
		rapid.StringMatching(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`),
	)
}
