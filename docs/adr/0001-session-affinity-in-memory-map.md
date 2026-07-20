# ADR-0001: Session Affinity via In-Memory Sticky Map, Not Redis

**Status:** Accepted
**Date:** 2026-07-20
**Milestone:** M1

## Context

Multi-worker dispatch (M1) needs repeat requests from the same `session_id` to reliably land on the same Worker — for conversational consistency and any locality benefit from that Worker already having relevant state warm. This requires a mapping from session to worker, held somewhere, with a lifecycle (created on first request, evicted when stale).

## Decision

An in-memory map inside the Scheduler process: `sessionID -> (workerID, lastActivity)`. A sweeper goroutine evicts entries past a staleness threshold. Shape: `docs/SCHEMA.md §2`.

## Alternatives Considered

| Option | Why not chosen |
|---|---|
| External store (Redis) | Adds a network hop and an operational dependency for a problem that only matters once the Scheduler itself is horizontally scaled — which is explicitly out of scope (`docs/PRD.md §5`). Solving for a scaling scenario that doesn't exist yet is premature. |
| No session affinity — pure round-robin | Simpler, but drops both the conversational-consistency property and the interview-relevant demonstration of a real scheduling concern; not a meaningful cost saving given the in-memory map is cheap. |

## Consequences

Session state is lost on Scheduler restart — acceptable, because affinity is a soft optimization, not a correctness guarantee: a lost mapping just means the next request from that session reroutes and possibly loses any locality benefit, it doesn't produce a wrong result. This decision does not scale past one Scheduler process; if the Scheduler is ever horizontally sharded, this ADR should be revisited and superseded.

## Superseded By

—
