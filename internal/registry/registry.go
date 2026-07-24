package registry

import (
	"sync"
	"time"
)

type WorkerStatus int

const (
	WorkerHealthy     WorkerStatus = iota
	WorkerDegraded    WorkerStatus = iota
	WorkerUnreachable WorkerStatus = iota
)

type WorkerRecord struct {
	WorkerID      string
	Address       string
	LoadedModel   string
	MaxBatchSize  int
	LastHeartbeat time.Time
	Status        WorkerStatus
}

type Registry struct {
	mu             sync.Mutex
	workers        map[string]*WorkerRecord
	heartbeatLimit time.Duration
}

func NewRegistry(heartbeatLimit time.Duration) *Registry {
	return &Registry{
		workers:        make(map[string]*WorkerRecord),
		heartbeatLimit: heartbeatLimit,
	}
}

func (r *Registry) Register(workerID, address, loadedModel string, maxBatchSize int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.workers[workerID]; ok {
		existing.Address = address
		existing.LoadedModel = loadedModel
		existing.MaxBatchSize = maxBatchSize
		existing.LastHeartbeat = time.Now()
		existing.Status = WorkerHealthy
		return
	}

	r.workers[workerID] = &WorkerRecord{
		WorkerID:      workerID,
		Address:       address,
		LoadedModel:   loadedModel,
		MaxBatchSize:  maxBatchSize,
		LastHeartbeat: time.Now(),
		Status:        WorkerHealthy,
	}
}

func (r *Registry) Heartbeat(workerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, ok := r.workers[workerID]; ok {
		w.LastHeartbeat = time.Now()
		if w.Status == WorkerUnreachable {
			w.Status = WorkerHealthy
		}
	}
}

func (r *Registry) Sweep() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, w := range r.workers {
		if time.Since(w.LastHeartbeat) > r.heartbeatLimit {
			w.Status = WorkerUnreachable
		}
	}
}

func (r *Registry) HealthyWorkers() []WorkerRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []WorkerRecord
	for _, w := range r.workers {
		if w.Status == WorkerHealthy {
			result = append(result, *w)
		}
	}
	return result
}

func (r *Registry) Lookup(workerID string) (WorkerRecord, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	w, ok := r.workers[workerID]
	if !ok {
		return WorkerRecord{}, false
	}
	return *w, true
}

func (r *Registry) MarkUnreachable(workerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if w, ok := r.workers[workerID]; ok {
		w.Status = WorkerUnreachable
	}
}

func (r *Registry) Unregister(workerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.workers, workerID)
}
