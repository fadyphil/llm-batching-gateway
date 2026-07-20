# ADR-0006: Milestone-Anchored Kanban Over Scrum

**Status:** Accepted
**Date:** 2026-07-20
**Milestone:** M0

## Context

A six-month solo build needs enough process to avoid drifting without ever declaring anything done, without adopting ceremony that has nothing to coordinate against.

## Decision

Kanban board with a tight WIP limit (1 card "Doing," 2 only if genuinely blocked), anchored to the milestones in `docs/PRD.md §6` as the checkpoint where "done" is actually defined. Scrum's retro is kept, as a short closeout note per milestone (`docs/ROADMAP.md`); its other ceremonies are dropped.

## Alternatives Considered

| Option | Why not chosen |
|---|---|
| Full Scrum with timeboxed sprints | Sprint commitment, velocity tracking, and most ceremonies exist to synchronize *multiple people's* uncertainty against each other. Solo, there's no one to synchronize with — a "sprint fail" with no one to answer to is just discouraging, not informative. |
| No process — pure ad hoc | Named risk: unstructured solo projects drift without a forcing function for "done," which is how portfolio repos end up abandoned mid-build. |

## Consequences

Milestone cadence is uneven — driven by real dependency completion (`docs/PRD.md §6`'s dependency graph), not a fixed two-week clock. Accepted tradeoff for a build where available time varies week to week; a fixed sprint clock would just produce sprints that fail for reasons unrelated to the work itself.

## Superseded By

—
