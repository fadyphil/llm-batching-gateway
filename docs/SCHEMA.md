# Schema Reference

Pure reference — every data shape that crosses a service boundary or lives in memory long enough to matter for correctness. No "why" here; that's `ARCHITECTURE.md` and the relevant ADRs. This file explains *what shape*, they explain *why that shape*.

**Keep this in sync with the actual `.proto` files and struct definitions.** A schema doc that's drifted from the code is worse than no schema doc — per `CONTRIBUTING.md`'s Definition of Done, any PR that changes a public contract updates this file in the same PR.

---

## 1. Wire Contracts (Protocol Buffers)

Canonical definitions live in `/proto/*.proto`; this section is the human-readable index. Full message text: `ARCHITECTURE.md §<TBD>` has illustrative snippets in context — this table is the exhaustive version.

### `gateway.proto` — client-facing

| Message/RPC | Field | Type | Notes |
|---|---|---|---|
| `CompletionService.Complete` | — | `rpc(CompletionRequest) returns (stream CompletionChunk)` | Client-facing streaming completion |
| `CompletionRequest` | `session_id` | `string` | Drives session affinity, see §2 |
| | `prompt` | `string` | |
| | `model` | `string` | Selects batch key, see `ARCHITECTURE.md` |
| | `priority` | `Priority` (enum) | `PRIORITY_UNSPECIFIED = 0`, `INTERACTIVE = 1`, `BACKGROUND = 2` |
| `CompletionChunk` | `request_id` | `string` | |
| | `token` | `string` | |
| | `is_final` | `bool` | |
| | `finish_reason` | `string` | |

### `scheduler.proto` — Gateway → Scheduler (internal)

| Message/RPC | Field | Type | Notes |
|---|---|---|---|
| `SchedulerService.Enqueue` | — | `rpc(EnqueueRequest) returns (stream CompletionChunk)` | |
| `EnqueueRequest` | `request_id` | `string` | |
| | `session_id` | `string` | |
| | `prompt` | `string` | |
| | `model` | `string` | |
| | `priority` | `Priority` | Shared enum with `gateway.proto` |
| | `token_count` | `int32` | Populated by Tokenizer before this call, see §2 pipeline note |

### `worker.proto` — Scheduler → Worker (internal)

| Message/RPC | Field | Type | Notes |
|---|---|---|---|
| `WorkerService.RunBatch` | — | `rpc(BatchDispatch) returns (stream WorkerChunk)` | |
| `BatchDispatch` | `batch_id` | `string` | |
| | `items` | `repeated BatchItem` | |
| `WorkerChunk` | `request_id` | `string` | Demux key — a single stream carries tokens for every item in the batch |
| | `token` | `string` | |
| | `is_final` | `bool` | |

### `tokenizer.proto` — Scheduler → Rust sidecar (internal)

| Message/RPC | Field | Type | Notes |
|---|---|---|---|
| `TokenizerService.CountTokens` | — | `rpc(TokenizeRequest) returns (TokenizeResponse)` | Synchronous, on the enqueue hot path — keep this fast |
| `TokenizeResponse` | `token_count` | `int32` | |
| | `exceeds_budget` | `bool` | |
| | `suggested_truncation` | `int32` | |

---

## 2. Internal State Shapes (not on the wire)

These live in process memory, defined in Go — not proto, but they're exactly as load-bearing for correctness.

```go
// Scheduler — one per enqueued request, held until dispatch
type pendingRequest struct {
    req        *pb.EnqueueRequest
    resultChan chan *pb.CompletionChunk // demux target for this request's tokens
    enqueuedAt time.Time
}

// Scheduler — one per (model, priority) batch key, open until dispatch
type openBatch struct {
    key        batchKey // (model string, priority Priority)
    items      []*pendingRequest
    timer      *time.Timer
    maxSize    int
    windowOpen time.Time
}

// Scheduler — session affinity map; sessionID -> sticky worker
// See docs/adr/0001-session-affinity-in-memory-map.md for why this is
// in-memory rather than an external store, and its eviction policy.
type sessionEntry struct {
    workerID     string
    lastActivity time.Time
}

// Registry — one per known worker
type workerRecord struct {
    workerID      string
    address       string
    loadedModel   string
    maxBatchSize  int
    lastHeartbeat time.Time
    status        WorkerStatus // HEALTHY | DEGRADED | UNREACHABLE
}
```

---

## 3. Config Schema (per service)

To be filled in as each service's actual config surface is implemented — placeholder structure so the section exists rather than being retrofitted later:

| Service | Env var | Purpose |
|---|---|---|
| *(all)* | `LOG_LEVEL` | TBD once `docs/RUNBOOK.md` exists |
| *(all)* | `SERVICE_PORT` | TBD |

---

## 4. Versioning & Compatibility

Field-number discipline and `buf breaking` CI enforcement are defined in `CODE_STYLE.md §5` — not repeated here to avoid the exact redundancy this doc is trying to prevent elsewhere.
