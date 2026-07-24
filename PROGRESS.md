# Progress Log

Append-only log of project state. Each entry records what was done and what's next.

---

## 2026-07-22 — M0 Foundation Complete

### Done

**Proto layer:**

- Defined 5 proto files under `proto/<package>/v1/` (buf lint clean, STANDARD ruleset):
  - `common/v1/common.proto` — `Priority` enum, `CompletionChunk`
  - `gateway/v1/gateway.proto` — `CompletionService.Complete`
  - `scheduler/v1/scheduler.proto` — `SchedulerService.Enqueue`
  - `worker/v1/worker.proto` — `WorkerService.RunBatch`
  - `tokenizer/v1/tokenizer.proto` — `TokenizerService.CountTokens`
- Go code generated to `proto/go/<package>/v1/`
- `buf generate` and `buf lint` verified working

**Go services (3 of 4, plus Auth stub):**

- **Gateway** (`internal/gateway/`, `cmd/gateway/`):
  - Accepts `Complete` RPC with streaming response
  - Validates static auth token from `authorization` metadata (FR-2)
  - Forwards to Scheduler via `Enqueue` gRPC call
  - Streams scheduler responses back to client
  - 4 unit tests, race-clean

- **Scheduler** (`internal/scheduler/`, `cmd/scheduler/`):
  - Batch size 1, immediate dispatch (no timer, no batching logic yet)
  - Creates batch ID, sends single-item batch to Worker via `RunBatch`
  - Streams worker responses back to Gateway
  - 3 unit tests, race-clean

- **Worker** (`internal/worker/`, `cmd/worker/`):
  - Fires HTTP POST to `llama-server` `/completion` with `stream=true`
  - Parses SSE response, forwards tokens via gRPC `RunBatchResponse`
  - 3 unit tests (tested with `httptest` fake llama-server), race-clean

- **Auth:** inline static token check in Gateway (no separate service)

**Infrastructure:**

- `go.mod` / `go.sum` initialized
- nginx config at `deploy/nginx/nginx.conf` (grpc_pass to gateway:9000)
- cmd entrypoints for all 3 services (env var config)

**Testing:**

- End-to-end smoke test (`internal/smoke/smoke_test.go`) — spins up fake llama-server, all 3 Go services, sends Complete RPC via gRPC, verifies 3 tokens received
- `go test -race -count=1 ./...` passes (11 tests)
- `go build ./...` passes
- `gofmt -l .` and `goimports -l .` clean (proto generated code excluded)

**Docs updated:**

- AGENTS.md (repo state, proto layout)
- ROADMAP.md (M0 marked Complete)
- milestones/M0-foundation.md (marked Complete)
- SCHEMA.md (proto paths, message names, enum values)
- ARCHITECTURE.md (sequence diagram wire names)
- README.md (status)

**Not done from M0 exit criteria:**

- Registry service skeleton (deferred to M1 — not needed for single-worker M0)
- Tagged release `v0.1.0-M0` (not tagged yet)

### Next

The roadmap splits into 3 parallel tracks:

#### Track 1: M1 — Core Scheduler (Must — highest priority)

The thesis of the project. Real batching logic:

- Dual-trigger batch dispatch (max-size OR window-expiry, whichever hits first)
- Bounded ingress channel as backpressure mechanism
- Per-(model, priority) batch key isolation
- Multi-worker dispatch and session affinity
- Registry with gRPC heartbeat health tracking
- Property-based tests with `pgregory.net/rapid`

#### Track 2: M-Tok — Rust Tokenizer Sidecar (Must)

- Hand-rolled BPE tokenizer as `tonic` gRPC service
- BPE core as pure functions (no gRPC dependency)
- Testable against reference tokenizer output

#### Track 3: M-UI — Flutter Playground (Must)

- Streaming token-by-token reveal
- Model selector, session management
- Clean architecture (BLoC, repository pattern)
