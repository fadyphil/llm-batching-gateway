package commonv1

type Priority int32

const (
	Priority_PRIORITY_UNSPECIFIED Priority = 0
	Priority_INTERACTIVE          Priority = 1
	Priority_BACKGROUND           Priority = 2
)
