# ADR-XXXX: <Short, decision-focused title>

**Status:** Proposed | Accepted | Rejected | Superseded by ADR-XXXX
**Date:** YYYY-MM-DD
**Milestone:** <which milestone this decision belongs to, if any>

## Context

What is the problem? What forces — technical constraints, hardware limits (RTX 3050 / 4GB VRAM), timeline, the six-month solo-build constraint — are pushing on this decision? State the problem before the answer; a reader should be able to stop here and predict roughly where this is going.

## Decision

The actual decision, stated in one or two sentences, unhedged. Not "we could consider" — "we will."

## Alternatives Considered

| Option | Why not chosen |
| --- | --- |
| <Alternative A> | <Concrete reason — cost, complexity, scope fit, not just "worse"> |
| <Alternative B> | |

## Consequences

What does this decision make easier? What does it make harder later? What's the honest cost being accepted — this is where "we're accepting at-least-once delivery instead of exactly-once because distributed transactions are a much larger project" kind of statements belong. If this decision needs revisiting at a specific trigger point (e.g., "reconsider if worker count exceeds N"), say so explicitly.

## Superseded By

<Leave blank until/unless a later ADR reverses this one. Do not edit this ADR's Decision section after acceptance — supersede it with a new ADR instead, and link both directions.>
