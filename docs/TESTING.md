# Testing Strategy

The Iron Law and the red-green-refactor cycle are defined in `CONTRIBUTING.md §1` and `§5` — not repeated here. This document is the deeper reference: what layer of test to reach for, and specifically how the batching scheduler's correctness gets proven rather than just exercised.

## 1. Test Pyramid

| Layer | Scope | Tooling |
|---|---|---|
| Unit | Pure functions, single struct/module in isolation — no live network, process, or filesystem | Go: `testing` + `testify/assert`. Rust: built-in `#[test]`. Dart: `flutter_test` + `bloc_test` for BLoC state transitions |
| Integration | A real service boundary (e.g., Scheduler ↔ a real Tokenizer instance over gRPC on localhost) | Go: `testing` with a real `tonic` process spun up in `TestMain`; no mocked gRPC clients for these |
| Property-based | Invariants that must hold across the input space, not just example inputs — see §2 | Go: `pgregory.net/rapid` |
| Chaos / Load | Whole-system behavior under concurrency and induced failure — see §3 | `ghz` (gRPC load generator) |

Every layer is testable in isolation per `CODE_STYLE.md §1`'s Testability pillar — if a unit test needs a live process to pass, that's a design problem in the code, not a reason to skip the unit test.

## 2. Property-Based Testing — the Scheduler

Example-based tests answer "does this specific case work?" Property-based tests answer "does this hold for every case?" — and the batching scheduler is exactly the kind of concurrent, stateful component where the bug lives in the input you didn't think to write an example for.

Properties to encode as generators + invariant checks, not hand-written examples:

| Property | What it catches |
|---|---|
| A dispatched batch never exceeds `maxSize` | Off-by-one or race in the size-trigger check |
| A request is never held past `window + dispatch latency` without being dispatched | Timer logic bugs, especially around batch-key creation races |
| No request is silently dropped under concurrent enqueue (every enqueued `request_id` eventually appears in exactly one dispatched batch) | Channel/goroutine handoff bugs — the class of bug that's invisible in single-threaded example tests |
| Session affinity holds under concurrent enqueue from the same session, unless the sticky worker becomes unhealthy mid-test | Races between the sweeper goroutine and enqueue path |
| Batch key isolation: a burst of requests for model A never appears in a dispatched batch keyed to model B | Batch-key assignment bugs |

`rapid` generates randomized concurrent enqueue sequences (varying request count, timing, model/priority mix) and asserts these properties hold across many runs, not one. This is the correctness backbone referenced in `docs/PRD.md §2` and `CONTRIBUTING.md §5` — it's a Must-tier deliverable, not test-suite polish.

## 3. Chaos & Load Testing (M5)

Load generation: `ghz` against the Gateway's `Complete` RPC, varying concurrent stream count to find the actual throughput ceiling on the target hardware (RTX 3050, 4GB VRAM) — the honest number goes in `docs/PERFORMANCE.md`, not an assumed one.

Chaos scenarios, each with an expected, testable outcome:

| Scenario | Expected outcome |
|---|---|
| Kill a Worker process mid-batch | In-flight requests requeue to a healthy Worker and complete — `docs/adr/0004` |
| Saturate the ingress channel | New requests receive `RESOURCE_EXHAUSTED` immediately, no hang, no memory growth |
| Burst concurrent sessions against a single sticky Worker | Session affinity holds; Worker doesn't silently drop excess load — degrades via backpressure instead |
| Registry loses contact with all Workers | Scheduler surfaces a clear error state rather than hanging or silently accepting requests it can't serve |

If M5 gets cut for time (it's Could-tier per `docs/PRD.md §2`), a manually verified failover check — kill a worker, observe recovery once, record it — is an acceptable fallback; the property-based tests in §2 are what's non-negotiable, because they're what makes the correctness claim testable rather than anecdotal.

## 4. CI Integration

Once `.github/workflows/` exists, every PR runs: format/lint checks per `CODE_STYLE.md §6`, the full unit + integration + property-based suite, and `buf breaking` for proto compatibility. Chaos/load tests run manually or on a separate scheduled job, not on every PR — they're slow and infrastructure-heavy by nature.
