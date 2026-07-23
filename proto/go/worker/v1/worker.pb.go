package workerv1

type RunBatchResponse struct {
	RequestId string
	Token     string
	IsFinal   bool
}
