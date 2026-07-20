# AGENTS.md

Guidance for AI agents (OpenCode / Claude / Copilot) working in this repo. Read before writing any code or docs.

## Repo State

**Pre-implementation, docs-only.** No source, no `go.mod`/`Cargo.toml`/`pubspec.yaml`, no CI workflows, no `opencode.json`, no.lockfiles. Current milestone: M0 (not started). `docs/` *is* the project — treat it as the source of truth, not README prose. Code laid down in M0 must match the contracts in `docs/SCHEMA.md` and the boundaries in `docs/ARCHITECTURE.md`.

When you find yourself about to write code, **read in this order first**:
1. `CONTRIBUTING.md` (binding ground rules — same standard for AI and humans)
2. `CODE_STYLE.md` (enforced per-language style + the four pillars)
3. `docs/TESTING.md` (TDD workflow + the property-test mandate for the Scheduler)
4. `docs/ARCHITECTURE.md` + `docs/SCHEMA.md` (the spec you implement against)
5. `docs/adr/README.md` index, then any ADR touching the area you're in

## Non-Negotiable Rules (you will be blocked on these)

- **Iron Law: no production code without a failing test first.** "Wrote tests after" is explicitly rejected — tests that pass immediately prove nothing. If code exists before its test, delete the code and start over. See `CONTRIBUTING.md §1`, `docs/TESTING.md`.
- **ADR before implementation, not after.** Any decision with a real "why not X?" alternative gets an ADR (`docs/adr/template.md`, numbered `00XX-kebab-case.md`) *before* you build. The ADR index (`docs/adr/README.md`) must be updated in the same PR.
- **Property-based tests are mandatory for the batching Scheduler** (`pgregory.net/rapid`), not optional polish. Invariants to encode are listed in `docs/TESTING.md §2`.
- **No `TODO` without a linked GitHub Issue.** No merging with a "fix later" gap.
- **AI-authored commits get a trailer:** `Co-Authored-By: <agent> <noreply@...>` appended to the commit message. Not a disclaimer — required attribution.

## Commands (per language, run pre-commit)

Currently aspirational — no toolchains are installed/pinned yet (M0 prerequisite). Once the stack lands, these are the exact checks; do not substitute weaker defaults:

```bash
# Go (Gateway, Auth, Scheduler, Registry, Worker, Observability)
gofmt -l .                       # must produce no output
goimports -l .                    # must produce no output
golangci-lint run                 # govet, staticcheck, errcheck, gocyclo, revive, unused

# Rust (Tokenizer sidecar)
cargo fmt --check
cargo clippy --all-targets -- -D warnings   # pedantic baseline, warnings = errors

# Dart / Flutter (Client)
dart format --set-exit-if-changed .
flutter analyze --fatal-infos

# Protobuf
buf lint           # every PR touching .proto
buf breaking --against main   # proto compatibility; field numbers never reused
```

No single project-wide test runner — each language owns its own. There is no `Makefile`/`justfile`/`package.json` script; run per-language. `docs/RUNBOOK.md` is a scaffold, do not invent commands it doesn't list.

## Architecture Boundaries That Aren't Obvious From Filenames

- **Monorepo** (`docs/adr/0007`): one shared `/proto` dir is the source of truth for every wire contract. A proto change is one atomic commit across all consumers — never edit a proto in isolation assuming consumers will catch up.
- **Go services live under `internal/<service>/`** and **never import another service's `internal/`** — they talk over the proto-defined gRPC boundary, same as over the network. Treat the process boundary as the module boundary.
- **Interfaces are defined at the consumer, not the implementer** (`CODE_STYLE.md §2`): `WorkerDispatcher`, `TokenCounter`, etc. — small and behavior-named. Don't ship a 12-method `SchedulerDependencies`.
- **Rust tokenizer is a separate `tonic` gRPC service, not an FFI/cgo binding** (`docs/adr/0003`). The BPE core (merge rules, vocab lookup) is pure functions, kept separate from the `tonic` wrapper so it's testable with zero gRPC. Hand-rolled, not the `tokenizers` crate (`docs/adr/0002`).
- **Scheduler is single-goroutine, no lock on the hot path.** Batch state stays single-threaded by construction; `sync.Mutex` only around Registry + session-affinity map. Batch key = `(model, priority)`, not global. One open batch per key at a time. See `docs/ARCHITECTURE.md §3`.
- **Backpressure = one bounded ingress channel, period.** No unbounded secondary queue anywhere. Saturation → `RESOURCE_EXHAUSTED` immediately, never buffer-and-hope (`docs/ARCHITECTURE.md §5`).
- **At-least-once, not exactly-once** on worker crash (`docs/adr/0004`). Clients may see a stream restart; never silent loss. Don't reach for a dedup log.
- **Worker is deliberately dumb** — it fires N concurrent HTTP requests at a local `llama-server` (`--parallel N`) and demuxes. Do not reimplement GPU-level batching; the engineering value is upstream in the Scheduler's dispatch decision (`docs/ARCHITECTURE.md §6`).

## Style Rules That Differ From Defaults

- **Go:** `MixedCaps`, short receivers (`s *Scheduler`). `context.Context` always first, always propagated. Wrap errors with `%w` (never `%v`); sentinels/typed errors for anything a caller branches on. `panic` only for startup-config errors, never network input. No naked goroutines — lifetime owned by `errgroup`/`WaitGroup`/context.
- **Rust:** `.unwrap()`/`.expect()` forbidden outside tests. Library uses `thiserror`; `anyhow` only in `main.rs`. `clippy::pedantic` is the baseline. No stringly-typed data where an enum fits (token categories, error kinds, vocab-lookup results).
- **Dart:** Clean architecture layers `presentation → domain → data`, dependencies inward only. BLoC per screen; events/states are `freezed`/sealed, never raw `Map`s. Repository wraps the generated gRPC stub; BLoC depends on a repository interface, not the stub. Cancel every stream subscription in `dispose()` — leaked subs on long-lived streams are real leaks here.
- **Proto:** `snake_case` fields, `PascalCase` messages, `SCREAMING_SNAKE_CASE` enums with explicit `_UNSPECIFIED = 0`. Every RPC + message gets a one-line doc comment stating edge behavior. Field numbers never reused after a merged PR references the file.
- **Hard limits:** functions ~40 lines, files ~300 lines. Names express intent (`assignToBatch`, not `processItem`). Comments explain *why* only.

## Git Workflow

- **GitHub Flow.** `main` is always demoable, no `develop` branch. One short-lived branch per card. `docs/BRANCHING.md` is authoritative.
- **Branch naming:** `<type>/<short-kebab-case>` — `feature/`, `fix/`, `refactor/`, `test/`, `docs/`, `chore/`. Type matches the commit-type table in `CONTRIBUTING.md §4`.
- **Commit format:** Conventional Commits, `<type>(<scope>): <subject>`. Scopes are fixed: `gateway`, `auth`, `scheduler`, `worker`, `registry`, `tokenizer`, `observability`, `flutter`, `nginx`, `proto`, `docs`. Subject ≤50 chars, imperative, no period. Body explains *why*.
- **Rebase before PR:** `git fetch origin && git rebase origin/main`. Resolve conflicts locally, never in the PR.
- **Squash merge by default** (one card → one commit on `main`). Regular merge commit only when the red/green/refactor sequence is itself the documentation (judgment call).
- **Tag every milestone close:** annotated `vX.Y.0-<milestone-id>` (e.g. `v0.1.0-M0`). `CHANGELOG.md` is generated against tags.
- **Broken `main`: revert first, fix properly second.** Never hot-patch forward under pressure.

## What's Explicitly Out of Scope — Don't Build It

From `docs/PRD.md §5`: KV-cache paging/introspection, horizontal scheduler sharding, exactly-once delivery via dedup, OAuth/per-user quotas/multi-tenant auth, horizontal worker autoscaling. If a request seems to need one of these, push back in the PR rather than silently expanding scope.

## When You Don't Know

Say so in the PR description. Don't silently guess on module caller expectations or failure semantics — `CONTRIBUTING.md §1.4` requires explicit assumptions when blast radius is unclear. Open a `docs`-scope issue if a doc gap caused the confusion, and fix the gap in the same PR.
