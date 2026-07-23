package workerv1

import "google.golang.org/grpc"

type RunBatchRequest struct {
	BatchId string
	Items   []*BatchItem
}

type BatchItem struct {
	RequestId string
	Prompt    string
}

type RunBatchResponse struct {
	RequestId string
	Token     string
	IsFinal   bool
}

type WorkerServiceServer interface {
	RunBatch(*RunBatchRequest, grpc.ServerStreamingServer[RunBatchResponse]) error
	mustEmbedUnimplementedWorkerServiceServer()
}

type UnimplementedWorkerServiceServer struct{}

func (UnimplementedWorkerServiceServer) RunBatch(*RunBatchRequest, grpc.ServerStreamingServer[RunBatchResponse]) error {
	return nil
}
func (UnimplementedWorkerServiceServer) mustEmbedUnimplementedWorkerServiceServer() {}

func RegisterWorkerServiceServer(s grpc.ServiceRegistrar, srv WorkerServiceServer) {
	s.RegisterService(&WorkerService_ServiceDesc, srv)
}

func _WorkerService_RunBatch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(RunBatchRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkerServiceServer).RunBatch(m, &grpc.GenericServerStream[RunBatchRequest, RunBatchResponse]{ServerStream: stream})
}

var WorkerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "worker.v1.WorkerService",
	HandlerType: (*WorkerServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RunBatch",
			Handler:       _WorkerService_RunBatch_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "worker/v1/worker.proto",
}
