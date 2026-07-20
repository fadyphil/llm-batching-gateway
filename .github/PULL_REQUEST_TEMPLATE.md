## What

<!-- One or two sentences: what does this PR change? -->

## Why

<!-- The motivation. Link the Issue/card. If this reverses or deviates from
     a prior decision, link the ADR it supersedes. -->

Closes #

## How

<!-- Key implementation details worth flagging for review — the parts that
     aren't obvious from reading the diff. -->

## Milestone

<!-- Which milestone (docs/milestones/MX-name.md) does this card belong to? -->

## Testing Evidence

<!-- TDD is mandatory per CONTRIBUTING.md §5. Show the red step happened,
     don't just assert it did. -->

- [ ] Test(s) written first; failure observed before implementation (paste the failing output or the pre-implementation commit hash)
- [ ] All tests pass locally, output pristine (no warnings, no skipped cases without a linked issue)
- [ ] New/changed failure-scenario coverage (hostile inputs, timeouts, partial failure, concurrency) — list which scenarios this PR specifically covers:
  -
- [ ] Property-based test added/updated, **if this PR touches the batching scheduler**

## ADR

<!-- Does this PR make a decision that would surprise a future reader?
     If yes, link the ADR (written BEFORE this PR, per CONTRIBUTING.md §7). -->

- [ ] N/A — no architecturally significant decision in this PR
- [ ] ADR linked: `docs/adr/00XX-title.md`

## Self-Review Checklist

<!-- Read your own diff as a stranger would before checking these. -->

- [ ] This PR is one logical, reviewable unit of work — if it's hard to summarize in one sentence above, it's probably two PRs
- [ ] `gofmt`/`cargo fmt`/`dart format` clean, linter clean, zero unexplained suppressions
- [ ] Every `titan-engineer` Quality Gate passes (error handling, isolation, state safety, observability, naming, size limits) — see `CODE_STYLE.md`
- [ ] No function >~40 lines / file >~300 lines without a documented reason it wasn't split
- [ ] No swallowed errors; no silent `nil`/zero-value returns on failure paths
- [ ] Public contracts changed (proto, exported API) → `docs/ARCHITECTURE.md` or proto comments updated in this same PR
- [ ] No `TODO` without a linked Issue
- [ ] I would be comfortable explaining every line of this diff cold, in an interview, six months from now

## Breaking Changes

- [ ] None
- [ ] Yes — described here, and `buf breaking` acknowledged if proto fields changed:
