# ADR-0009: In-memory worker registry for M1, separate gRPC service deferred to M3

**Status:** Proposed
**Date:** 2026-07-23
**Milestone:** M1

## Context

M1 requires the Scheduler to know which workers are healthy in order to dispatch batches (FR-4) and requeue on failure (FR-6). This requires a registry that stores worker metadata (ID, address, loaded model, max batch size, last heartbeat, status). Two fundamental options: an in-memory data structure inside the Scheduler process, or a standalone gRPC microservice.

## Decision

For M1, the worker registry lives as an in-memory data structure (`map[string]*WorkerRecord` + `sync.Mutex`) in the `internal/registry/` package, accessed directly within the Scheduler process. A separate Registry gRPC service (with its own proto, heartbeat RPCs, and standalone process) is deferred to M3 when Observability needs it as an independent data source. The `internal/registry/` package is structured so that it can be extracted behind a gRPC service boundary later without changing the interface its callers depend on.

## Alternatives Considered

| Option | Why not chosen |
|---|---|
| Full Registry gRPC service from M1 | Adds a process boundary, deployment dependency (separate binary, port, health-check wiring), and proto RPC definition before the internal data structure's access patterns are even validated by tests. Would slow down the Scheduler's own correctness work — the core deliverable of M1 — for the sake of a service boundary that doesn't add value until M3. |
| External store (Redis, etc.) | Overkill for a single-process scheduler. ADR-0001 already rejected this for session affinity on similar grounds. |

## Consequences

The Registry struct and its callers must be kept in the same process for M1. The package structure (`internal/registry/registry.go`) is designed so that M3 can wrap it behind a gRPC server with minimal refactoring — the Registry itself doesn't change, only a new `internal/registry/server.go` is added that calls its methods via a thin gRPC handler. Session state is lost on Scheduler restart (same as ADR-0001's consequence); this is acceptable because registry state is derived from worker heartbeats which resume on restart.

## Superseded By

<Leave blank until/unless a later ADR reverses this one. Do not edit this ADR's Decision section after acceptance — supersede it with a new ADR instead, and link both directions.>
