# ADR-0002: Hand-Rolled BPE Tokenizer Over the `tokenizers` Crate

**Status:** Accepted
**Date:** 2026-07-20
**Milestone:** M-Tok

## Context

The Rust sidecar (`docs/adr/0003`) needs to count and validate tokens before a request is admitted to a batch (FR-8). A production-grade BPE implementation already exists as a well-tested crate (HuggingFace's `tokenizers`).

## Decision

Hand-roll BPE encode/decode against a bundled public vocab/merges file, rather than depending on the `tokenizers` crate.

## Alternatives Considered

| Option | Why not chosen |
| --- | --- |
| Wrap the `tokenizers` crate | Ships faster and is production-battle-tested, but reduces this milestone to "I called a library" — a materially weaker answer to "walk me through something you built from first principles" than implementing BPE encode/decode directly, which is consistent with this project's existing from-scratch precedent (a prior Rust project implementing Git internals). |
| Implement tokenization in Go instead of a separate Rust service | Removes the cross-language boundary entirely, which undercuts the actual reason the sidecar exists — see `docs/adr/0003`. |

## Consequences

Costs roughly 2-3 extra days versus wrapping a crate. Requires sourcing and bundling a public vocab/merges file rather than deriving one from scratch. Correctness must be verified against a reference tokenizer's output on a fixed test corpus (`docs/PRD.md` FR-8's acceptance criteria) — a hand-rolled implementation that silently diverges from a real BPE vocabulary is a worse outcome than not building it at all, so this verification step is non-negotiable, not optional polish.

## Superseded By

—
