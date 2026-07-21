# ADR-0003: Rust Tokenizer as a Separate gRPC Service, Not an FFI Binding

**Status:** Accepted
**Date:** 2026-07-20
**Milestone:** M-Tok

## Context

The tokenizer (Rust, `docs/adr/0002`) needs to be callable from the Go Scheduler on the hot enqueue path. Two integration shapes are viable: a networked gRPC service, or a compiled FFI/cgo binding loaded directly into the Go process.

## Decision

A standalone `tonic`-based gRPC service, called over localhost from the Scheduler — same integration pattern as every other internal service boundary in this system (`docs/ARCHITECTURE.md §1`).

## Alternatives Considered

| Option | Why not chosen |
| --- | --- |
| FFI/cgo binding into the Go process | Faster (no serialization or network hop) and simpler to deploy as one binary, but couples the Go and Rust build toolchains, complicates cross-compilation, and — most importantly for this project's actual purpose — collapses "design a polyglot service boundary" into an in-process function call. That's a materially weaker thing to demonstrate; the whole point of the sidecar is proving a real cross-language service contract, not just cross-language code reuse. |
| Rewrite the tokenizer in Go | Removes the integration question entirely, but eliminates the Rust proof point altogether. |

## Consequences

Adds one network hop and serialization cost to every enqueue, on the hot path — `docs/ARCHITECTURE.md §2` flags this explicitly as a real latency constraint the Tokenizer service must be designed around, not an afterthought. Adds one more service to run, monitor, and version independently, with its own proto contract (`docs/SCHEMA.md §1`) subject to the same `buf breaking` compatibility discipline as every other service boundary.

## Superseded By

—
