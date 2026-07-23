package registry_test

import (
	"testing"
	"time"

	"github.com/fadyphil/llm-batching-gateway/internal/registry"
)

func TestRegister_NewWorker_Healthy(t *testing.T) {
	r := registry.NewRegistry(10 * time.Second)
	r.Register("w1", "localhost:8081", "llama3", 4)

	rec, ok := r.Lookup("w1")
	if !ok {
		t.Fatal("expected worker to be found after Register")
	}
	if rec.WorkerID != "w1" {
		t.Fatalf("expected WorkerID w1, got %s", rec.WorkerID)
	}
	if rec.Address != "localhost:8081" {
		t.Fatalf("expected Address localhost:8081, got %s", rec.Address)
	}
	if rec.LoadedModel != "llama3" {
		t.Fatalf("expected LoadedModel llama3, got %s", rec.LoadedModel)
	}
	if rec.MaxBatchSize != 4 {
		t.Fatalf("expected MaxBatchSize 4, got %d", rec.MaxBatchSize)
	}
	if rec.Status != registry.WorkerHealthy {
		t.Fatalf("expected Status WorkerHealthy, got %v", rec.Status)
	}
	if rec.LastHeartbeat.IsZero() {
		t.Fatal("expected LastHeartbeat to be set")
	}
}

func TestHeartbeat_Missed_MarksUnreachable(t *testing.T) {
	r := registry.NewRegistry(50 * time.Millisecond)
	r.Register("w1", "localhost:8081", "llama3", 4)

	time.Sleep(60 * time.Millisecond)
	r.Sweep()

	rec, ok := r.Lookup("w1")
	if !ok {
		t.Fatal("expected worker to exist after Sweep")
	}
	if rec.Status != registry.WorkerUnreachable {
		t.Fatalf("expected Status WorkerUnreachable after missed heartbeat, got %v", rec.Status)
	}
}

func TestHeartbeat_Refreshes_StaysHealthy(t *testing.T) {
	r := registry.NewRegistry(100 * time.Millisecond)
	r.Register("w1", "localhost:8081", "llama3", 4)

	time.Sleep(40 * time.Millisecond)
	r.Heartbeat("w1")
	time.Sleep(70 * time.Millisecond)
	r.Sweep()

	rec, ok := r.Lookup("w1")
	if !ok {
		t.Fatal("expected worker to exist")
	}
	if rec.Status != registry.WorkerHealthy {
		t.Fatalf("expected Status WorkerHealthy after recent heartbeat, got %v", rec.Status)
	}
}

func TestHealthyWorkers_ReturnsOnlyHealthy(t *testing.T) {
	r := registry.NewRegistry(50 * time.Millisecond)
	r.Register("w1", "localhost:8081", "llama3", 4)
	r.Register("w2", "localhost:8082", "mistral", 2)

	time.Sleep(60 * time.Millisecond)
	r.Sweep()

	healthy := r.HealthyWorkers()
	if len(healthy) != 0 {
		t.Fatalf("expected 0 healthy workers, got %d", len(healthy))
	}

	r.Heartbeat("w2")
	healthy = r.HealthyWorkers()
	if len(healthy) != 1 {
		t.Fatalf("expected 1 healthy worker, got %d", len(healthy))
	}
	if healthy[0].WorkerID != "w2" {
		t.Fatalf("expected w2, got %s", healthy[0].WorkerID)
	}
}

func TestMarkUnreachable_Explicit(t *testing.T) {
	r := registry.NewRegistry(10 * time.Second)
	r.Register("w1", "localhost:8081", "llama3", 4)

	r.MarkUnreachable("w1")
	rec, ok := r.Lookup("w1")
	if !ok {
		t.Fatal("expected worker to exist after MarkUnreachable")
	}
	if rec.Status != registry.WorkerUnreachable {
		t.Fatalf("expected Status WorkerUnreachable, got %v", rec.Status)
	}
}

func TestUnregister_Removes(t *testing.T) {
	r := registry.NewRegistry(10 * time.Second)
	r.Register("w1", "localhost:8081", "llama3", 4)

	r.Unregister("w1")
	_, ok := r.Lookup("w1")
	if ok {
		t.Fatal("expected worker to be removed after Unregister")
	}
}

func TestLookup_NonExistent_ReturnsFalse(t *testing.T) {
	r := registry.NewRegistry(10 * time.Second)
	_, ok := r.Lookup("nonexistent")
	if ok {
		t.Fatal("expected Lookup of nonexistent worker to return false")
	}
}

func TestRegister_ReRegister_RenewsHealth(t *testing.T) {
	r := registry.NewRegistry(50 * time.Millisecond)
	r.Register("w1", "localhost:8081", "llama3", 4)

	time.Sleep(60 * time.Millisecond)
	r.Sweep()

	rec, ok := r.Lookup("w1")
	if !ok {
		t.Fatal("expected worker to exist")
	}
	if rec.Status != registry.WorkerUnreachable {
		t.Fatalf("expected Status WorkerUnreachable after missed heartbeat, got %v", rec.Status)
	}

	r.Register("w1", "localhost:8081", "llama3", 4)
	rec, ok = r.Lookup("w1")
	if !ok {
		t.Fatal("expected worker to exist after re-register")
	}
	if rec.Status != registry.WorkerHealthy {
		t.Fatalf("expected Status WorkerHealthy after re-register, got %v", rec.Status)
	}
}
