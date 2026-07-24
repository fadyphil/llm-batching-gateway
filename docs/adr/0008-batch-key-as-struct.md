# ADR-0008: batchKey as struct, not string concatenation

**Status:** Proposed
**Date:** 2026-07-23
**Milestone:** M1

## Context

The Scheduler groups incoming requests into batches keyed by (model, priority). Requests for different models must never share a batch; requests for different priority tiers must never share a batch. This key is used as a map key in the `openBatches` map. Go maps can use any comparable type as a key, so we have two viable options: a struct with both fields, or a string that concatenates them with a separator.

## Decision

Use a struct type: `struct { Model string; Priority Priority }` as the batch key. In Go, a struct with no slice/map fields is comparable and works directly as a map key. This gives type safety — a misspelled enum value or empty model string is caught at compile time in the struct field, not buried at runtime in a string prefix search.

## Alternatives Considered

| Option | Why not chosen |
|---|---|
| String concatenation (`Model + "-" + Priority`) | Fragile: a model named `"llama-7b-Background"` and a model `"llama-7b"` with priority `Background` would collide if the separator isn't carefully managed. More subtly, the Priority enum's String() representation could change in a proto update and silently break the key format across a binary update boundary. No compile-time checking. |
| Struct with Serialized/JSON string | Unnecessary serialization overhead on every enqueue — the struct is already comparable natively. |

## Consequences

map key comparisons are fast and type-safe. The key can be extended later (e.g., adding a user_tenant field for future multi-tenant support) without changing the keying strategy — just add a field to the struct. The only cost is slightly more verbose map lookup syntax, which is negligible in the single-goroutine event loop.

## Superseded By

<Leave blank until/unless a later ADR reverses this one. Do not edit this ADR's Decision section after acceptance — supersede it with a new ADR instead, and link both directions.>
