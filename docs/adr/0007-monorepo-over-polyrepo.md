# ADR-0007: Monorepo Over Per-Service Polyrepo

**Status:** Accepted
**Date:** 2026-07-20
**Milestone:** M0

## Context

The system spans four Go services, a Rust sidecar, and a Flutter client, all consuming a shared set of proto contracts (`docs/SCHEMA.md §1`). Repository layout — one repo for everything, or one repo per service — needs to be decided before M0's proto/skeleton work starts.

## Decision

Single monorepo, one shared `/proto` directory as the source of truth for every wire contract.

## Alternatives Considered

| Option | Why not chosen |
|---|---|
| Polyrepo, one repo per service | Has real signal value — demonstrates managing cross-repo dependencies and versioned contract publishing — but for a single developer, a proto change becomes a manually coordinated multi-repo version bump on every touch, which is daily friction with no corresponding coordination benefit, since there's no second team the repo boundary is protecting anyone from. |

## Consequences

A single CI configuration has to handle three languages (Go, Rust, Dart) — addressed with per-language jobs once `.github/workflows/` exists, not a blocker at design time. In exchange, a proto contract change is one atomic commit across every consumer, which is the actual win this decision buys: no window where services disagree about a contract's shape.

## Superseded By

—
