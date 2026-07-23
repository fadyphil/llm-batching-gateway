package scheduler

import "testing"

func TestWorkerDispatcherInterfaceIsDefined(t *testing.T) {
	var d WorkerDispatcher
	if d != nil {
		t.Fatal("expected nil WorkerDispatcher")
	}
}
