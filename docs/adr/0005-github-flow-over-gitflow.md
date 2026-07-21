# ADR-0005: GitHub Flow Over GitFlow / Full Trunk-Based Development

**Status:** Accepted
**Date:** 2026-07-20
**Milestone:** M0

## Context

A six-month, solo-authored build needs a git branching model. The candidate models coordinate different amounts of concurrent human activity: GitFlow assumes a team large enough to need `develop`/`release`/`hotfix` branches buffering scheduled releases from in-progress work; full trunk-based development assumes several engineers committing directly to `main` behind feature flags to avoid stepping on each other.

## Decision

GitHub Flow: `main` is always demoable, protected, and every change lands via a short-lived feature branch and a self-reviewed PR. Full detail: `docs/BRANCHING.md`.

## Alternatives Considered

| Option | Why not chosen |
| --- | --- |
| GitFlow | Solves a coordination problem — buffering unstable in-progress work from a scheduled release — that doesn't exist with one committer. The extra branch types are pure overhead here. |
| Full trunk-based development with feature flags | Feature flags exist to let multiple engineers commit incomplete work to `main` simultaneously without blocking each other. Solo, the equivalent is simpler: just don't merge until it's ready. Flag infrastructure would be unused complexity. |

## Consequences

There's no formal release branch — a milestone tag directly on `main` is the release artifact (`docs/BRANCHING.md §6`). This requires real discipline to keep `main` always demoable, since there's no buffer branch protecting it if that discipline slips.

## Superseded By

—
