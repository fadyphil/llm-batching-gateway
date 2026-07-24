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

---

## 2026-07-24 — M1 Core Scheduler Library Landed

### Done

**Proto extensions** (regenerated via `buf generate` against merged .proto tree):

- `CompletionChunk` and `FinishReason` added in `proto/common/v1/common.proto`
- `EnqueueResponse` enriched in `proto/scheduler/v1/scheduler.proto`
- Generated `proto/go/common/v1/common.pb.go`, `proto/go/scheduler/v1/scheduler.pb.go`, `proto/go/worker/v1/worker.pb.go` updated

**Scheduler library** (`internal/scheduler/`, beside the M0 skeleton that is kept):

- `dispatcher.go` — `WorkerDispatcher` interface defined at the consumer (small, behavior-named), per CODE_STYLE §2
- `scheduler.go` — `Scheduler` struct with single-goroutine dispatch loop (no lock on the hot path, ARCHITECTURE §3)
- `config.go` — `Config`, `RealTimeSource`, `FakeTimeSource`, and the `Timer` / `TimeSource` interfaces
- `property_test.go` — `pgregory.net/rapid` property test for the "no silently dropped requests under concurrent enqueue" invariant (TESTING §2)
- `scheduler_test.go`, `generator_test.go`, `mock_test.go` — unit + generator helpers for the package
- `server.go` and `server_test.go` (M0 passthrough skeleton, batch size 1) — **kept for now**

**Registry** (`internal/registry/`, new package):

- `registry.go` — in-memory map of worker health with `sync.Mutex`, `Register`, `Heartbeat`, `Sweep`, `HealthyWorkers` methods
- `registry_test.go` — unit tests covering heartbeat TTL expiry (FR-9)

**Worker** (`internal/worker/server.go`, `server_test.go`):

- Replaced M0's single-item SSE passthrough with concurrent multi-item batching: `errgroup` of N goroutines, per-item HTTP fanout to `llama-server`, single demux goroutine for `stream.Send()` (FR-7)
- Partial-failure handler now race-clean (data-race fix landed in the same series)

**ADRs added:**

- `ADR-0008` — `batchKey` as struct, not string concatenation (M1)
- `ADR-0009` — in-memory registry for M1, separate gRPC service deferred to M3
- `docs/adr/README.md` index updated to spans 0001–0009

**Deps / build:**

- `go.mod` / `go.sum` tidied: `google.golang.org/protobuf` promoted from indirect to direct
- `cmd/worker/main.go` — identical on both sides, no integration change
- Naming fixup: `internal/scheduler/generator_test.go` adjusted `Priority_INTERACTIVE`/`Priority_BACKGROUND` to the current protoc-gen-go output `Priority_PRIORITY_INTERACTIVE`/`Priority_PRIORITY_BACKGROUND`. No behavior change.

**Verification (run on the resolution branch):**

- `gofmt -l .` — empty
- `go vet ./...` — clean
- `go build ./...` — exit 0
- `go test ./...` — all pass, including rapid property tests in `internal/scheduler/property_test.go`
- `git merge-tree` simulation against `origin/main` — 0 conflict markers, GH merge will go through cleanly

### Branch situation

`feature/m1-core-scheduler` was forked from commit `de38c1e` (the docs scaffolding before M0 work started). After PR #1 merged M0 into `main`, the m1 branch diverged from `origin/main`: 18 commits ahead on the m1 side and 11 ahead on the main side, sharing base `de38c1e`. `git merge-tree` against `origin/main` reported real conflicts (`<<<<<<<` markers, 17 hunks) on 10 files:

- `cmd/worker/main.go` (turned out identical on both sides; auto-merged)
- `docs/adr/README.md` (auto-merged cleanly: 0008/0009 appended)
- `docs/ROADMAP.md` (M1 row kept as 🚧 In Progress from m1)
- `go.mod`, `go.sum` (taken from m1, tidied)
- `internal/worker/server.go`, `server_test.go` (m1's concurrent-batch version wins)
- `proto/go/{common,scheduler,worker}/v1/*.pb.go` (regenerated from buf)

**Decisions made during the merge:**

1. **Squash-merge** onto a fresh branch `merge/m1-into-main` off `origin/main`. Chose this rather than a fast-forward / no-merge to keep one tidy integration commit subject line for the GH merge; rejected rebase per user direction (history preservation was not the goal — a clean PR diff was).
2. **`go.mod`/`go.sum` resolution**: take m1's version, then run `go mod tidy`. Did not pick by hand because dependency graph merges are unsafe to do as text (the resulting indirect/direct classification was wrong until `tidy` re-evaluated it).
3. **`.pb.go` files: regenerate** with `buf generate`, do not hand-merge. Confirmed by the `Priority_*` naming fixup (proto enum names now prefixed with enum name in current `protoc-gen-go`).

### Wiring: deferred to a follow-up card

The M1 library cannot be wired into gRPC yet. The new `Scheduler.Enqueue` takes an internal `EnqueueServer` interface (Send-only) and a `WorkerDispatcher` whose `RunBatch` signature `(ctx, workerID, []*EnqueueRequest) -> (<-chan resp, <-chan err, error)` does not match proto's `WorkerServiceClient.RunBatch(ctx, *RunBatchRequest, opts...) (grpc.ServerStreamingClient, error)`. The proto-generated `SchedulerServiceServer` interface also wants `grpc.ServerStreamingServer[EnqueueResponse]`, not the Send-only `EnqueueServer`.

The deletion of `internal/scheduler/server.go` (the M0 passthrough skeleton) is therefore **not in this merge**. Following the milestone doc's "Server.go changes: Progressive replacement across M1.1–M1.6" — server.go is supposed to die across several cards, not in a single big-bang.

The follow-up card (to be filed on the M1 milestone) covers:

1. Write a gRPC adapter inside the scheduler package that:
   - Implements `SchedulerServiceServer` (so it can be registered with `sv1.RegisterSchedulerServiceServer`)
   - Adapts `WorkerServiceClient` (gRPC) → `WorkerDispatcher` (M1 internal)
   - Adapts `grpc.ServerStreamingServer[EnqueueResponse]` → internal `EnqueueServer`
2. Update `cmd/scheduler/main.go` to instantiate the adapter with the worker gRPC client and pass it to `grpc.NewServer()`.
3. Update `internal/smoke/smoke_test.go` to drive the same wiring (currently uses `scheduler.NewServer`).
4. Once 1–3 are green and reviewed, `git rm internal/scheduler/server.go` (and `server_test.go`).

Test-first per the Iron Law (`CONTRIBUTING.md §1`, `AGENTS.md` "Non-Negotiable Rules"):

- Write a failing adapter test (round-trip via the public gRPC interface using `grpc.ClientConn` against an in-process `bufconn` listener).
- Make it green by implementing the adapter.
- Update `main.go` and `smoke_test.go` only once their existing tests fail to compile against the new constructor.

### Not done in this merge (deliberate)

- gRPC adapter for the new `Scheduler` (follow-up card, see above)
- Removal of `internal/scheduler/server.go` + `server_test.go`
- `v0.2.0-M1` annotated tag (per BRANCHING §6, tag in the merge commit that lands M1 on main, *after* the GH squash)
- Updating the M1 milestone doc's `Status:` line from "Not started" to the post-merge state

### Next

The M1 thread continues on the deferred wiring card. Once that lands and the M1 milestone closes, parallel tracks pick back up:

- **M-Tok** — Rust tokenizer sidecar, hand-rolled BPE per `ADR-0002`, `tonic` gRPC wrapper per `ADR-0003`
- **M-UI** — Flutter playground, BLoC + repository per CODE_STYLE §3

If the wiring slips past the M1 milestone close (two weeks), it rolls forward to a hot-fix card rather than blocking the milestone tag — the library is already in `main`, the wire-up is mechanical.
