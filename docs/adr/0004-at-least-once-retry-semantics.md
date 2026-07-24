# ADR-0004: At-Least-Once Delivery on Worker-Crash Retry, Not Exactly-Once

**Status:** Accepted
**Date:** 2026-07-20
**Milestone:** M3

## Context

When a Worker crashes mid-batch (detected via missed Registry heartbeats), the in-flight requests in its batch need a defined fate: retried elsewhere, or dropped. Guaranteeing each request completes *exactly once* — never retried into a visible duplicate, never lost — requires durable dedup state (an idempotency log, or a distributed-transaction boundary between the Scheduler's dispatch decision and the Worker's execution).

## Decision

At-least-once: on Worker failure, in-flight requests are requeued to a healthy Worker. A client may in rare cases observe a stream restart (tokens begin again from the start of that request) rather than a seamless resume — but never silent loss.

## Alternatives Considered

| Option | Why not chosen |
| --- | --- |
| Exactly-once via a durable dedup log / distributed transaction | Real engineering cost — durable state, idempotency keys, likely an external store — disproportionate to what a single-process portfolio scheduler needs to prove. Explicitly named as out of scope, `docs/PRD.md §5`. |
| Best-effort / drop in-flight requests on Worker failure | Silent data loss is a strictly worse failure mode than a visible stream restart — it violates the Debuggability and Traceability pillars in `CODE_STYLE.md §1`: a dropped request with no trace is undiagnosable, a restarted stream is a visible, testable, explainable behavior. |

## Consequences

Client-visible behavior on Worker crash is a stream restart, not a transparent resume — this must be documented as expected behavior (`docs/ARCHITECTURE.md §8`) and explicitly covered by a chaos test (`docs/TESTING.md §3`), not treated as an edge case nobody checks. In exchange, the system has zero risk of silently losing a request, which is the property actually worth demonstrating in an interview — "here's how it fails, and here's proof it doesn't fail silently" is a stronger claim than an unproven "it never fails."

## Superseded By

—
