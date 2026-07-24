# Code Style & Engineering Standards

This document is the enforced style guide for every language in this repository. "Enforced" means: a formatter and linter check most of this automatically (§6), and the rest is checked against the Definition of Done in `CONTRIBUTING.md §5` on every PR. Style violations without justification are a merge blocker, not a nitpick.

---

## 1. Cross-Language Principles

These apply identically to Go, Rust, and Dart. Language sections below are the *implementation* of these principles, not a separate set of rules.

### The Four Pillars (in priority order when they conflict)

1. **Readability** — write for a reader with zero context. Self-documenting names beat comments. Explicit beats clever.
2. **Debuggability** — when something breaks in production, what does a stranger need to diagnose it in 10 minutes? Fail fast, fail loud, structured logs with context. Never swallow an error silently.
3. **Testability** — if a module can't be verified in isolation (no live network/DB/process), the *design* is wrong, not the test. Dependencies are injected, not constructed inside the module that uses them.
4. **Traceability** — a requirement should be traceable to the code that implements it; a failure should be traceable to its source in minutes, via typed errors and structured logs, not string-grepping.

### SOLID, Language-Agnostic

| Principle | Rule | Common violation in this codebase's context |
| --- | --- | --- |
| Single Responsibility | One module, one reason to change | A Scheduler function that both batches requests *and* formats gRPC responses |
| Open/Closed | Extend via new code, don't edit tested code to add a case | Adding a new `Priority` tier by editing the batch-dispatch `switch` instead of a strategy the switch already supports |
| Liskov Substitution | Every implementation honors its abstraction's full contract | A `WorkerClient` mock that silently no-ops on `RunBatch` instead of returning the same error shape a real worker would |
| Interface Segregation | Small, focused contracts | A `SchedulerDependencies` interface with 12 methods when a given caller uses 2 |
| Dependency Inversion | Business logic depends on abstractions, not concrete infra | The batching algorithm importing a gRPC client type directly instead of a narrow `WorkerDispatcher` interface |

### Hard Limits

- Functions: **~40 lines.** Past that, it has more than one responsibility — split it.
- Files: **~300 lines.** Past that, split by responsibility, not arbitrarily.
- Naming names **intent**, never implementation: `assignToBatch`, not `processItem`. `WorkerUnavailableError`, not `Error1`.
- Comments explain **why**, never **what**: `// exponential backoff with jitter avoids thundering herd on 1000 simultaneous client retries` — not `// retry logic`.

### Error Handling — Universal Rule

Every fallible operation has explicit, typed handling. No empty catch/rescue blocks. No swallowed errors converted to a zero value or `null` — that turns a loud failure into silent data corruption. Every external call (network, disk, subprocess) has an explicit timeout and a defined retry policy — never fire-and-forget.

---

## 2. Go (Gateway, Auth, Scheduler, Registry, Worker, Observability)

**Formatting & linting — non-negotiable, run pre-commit:**

```bash
gofmt -l .          # must produce no output
goimports -l .       # must produce no output
golangci-lint run    # govet, staticcheck, errcheck, gocyclo, revive, unused enabled
```

**Errors:**

- Wrap with context using `%w`, never `%v`, so `errors.Is`/`errors.As` work up the call stack: `fmt.Errorf("dispatch batch %s: %w", batchID, err)`.
- Define sentinel or typed errors for anything a caller needs to branch on: `var ErrBatchFull = errors.New("batch at max size")`, not a bare string compared with `strings.Contains`.
- `panic` is reserved for programmer errors caught at startup (bad config). Never panic on bad input from a network boundary — return an error.

**Concurrency:**

- `context.Context` is always the first parameter and is always propagated — a client disconnect must actually cancel downstream work, not leak it.
- Every goroutine's lifetime is owned by something: an `errgroup.Group`, a `sync.WaitGroup`, or a context cancellation. No naked `go func(){...}()` with no way to know when it's done or to stop it.
- Channel direction and buffering are documented at the declaration: `incoming chan *pendingRequest // bounded: this IS the backpressure mechanism, see ARCHITECTURE.md §8`.
- Shared mutable state gets exactly one owner goroutine where possible (see the Scheduler's single-threaded batch-state design in `ARCHITECTURE.md`); where a mutex is unavoidable, the guarded fields are documented next to the `sync.Mutex` field, not scattered.

**Naming & structure:**

- `MixedCaps`, never underscores. Short receiver names (`s *Scheduler`, not `scheduler *Scheduler`).
- Interfaces are named for behavior, kept small (`WorkerDispatcher`, `TokenCounter`), and defined at the *consumer*, not the implementer — this is what makes Dependency Inversion real in Go rather than aspirational.
- Package layout: `internal/<service>/` per service, no service imports another service's `internal/` package directly — they talk over the proto-defined gRPC boundary, same as they would over the network in production.

---

## 3. Rust (Tokenizer Sidecar)

**Formatting & linting — non-negotiable, run pre-commit:**

```bash
cargo fmt --check
cargo clippy --all-targets -- -D warnings   # clippy::pedantic baseline, warnings are errors
```

**Errors:**

- Library code (the BPE tokenizer core) uses `thiserror` for typed, specific error enums. The `tonic` service boundary translates those into gRPC status codes explicitly — no `anyhow::Error` leaking across the service boundary.
- `anyhow` is acceptable only at the binary's `main.rs` for top-level error reporting, never inside the tokenizer library itself.
- `.unwrap()` and `.expect()` are **forbidden outside test code.** Every fallible call is propagated with `?` or explicitly matched. If a genuinely-impossible-to-fail case needs an escape hatch, it gets a comment explaining the invariant that makes it safe, and `.expect("<why this can't fail>")` with a real message, never a bare `.unwrap()`.

**Style:**

- `snake_case` for everything except types/traits (`PascalCase`). No stringly-typed data where an enum expresses the same thing — token categories, error kinds, and vocab-lookup results are all enums, not strings compared against magic values.
- The BPE implementation is hand-rolled (see `docs/adr/0002-hand-rolled-bpe-over-crate.md`) — keep the merge-rule application and vocab lookup as pure functions separate from the `tonic` service wrapper, so the tokenizer core is testable with zero gRPC machinery involved, consistent with the Testability pillar.

---

## 4. Dart / Flutter (Client)

**Formatting & linting — non-negotiable, run pre-commit:**

```bash
dart format --set-exit-if-changed .
flutter analyze --fatal-infos
```

Lint ruleset: `flutter_lints` as the floor; project-specific rules layered in `analysis_options.yaml` for anything this style guide requires that the default set doesn't catch (e.g., banning business logic in widget `build()` methods where feasible to lint).

**Architecture — matches your existing stack, applied here specifically:**

- Clean architecture layering: `presentation` (widgets + BLoC) → `domain` (entities, use cases, repository interfaces) → `data` (repository implementations, gRPC client wrappers). Dependencies point inward only — `domain` never imports anything from `data` or `presentation`.
- BLoC per screen-level concern (`PlaygroundBloc`, `OpsDashboardBloc`), events and states as sealed/`freezed` classes, never a raw `Map` or loosely-typed payload crossing a BLoC boundary.
- Repository pattern wraps the generated gRPC client — the BLoC layer depends on a repository *interface*, injected, never on the generated gRPC stub directly. This is what makes the streaming token-reveal UI testable without a live Scheduler.
- Dependency injection via constructor injection (or `get_it` if a service locator is preferred) — no static singletons holding gRPC channels.

**Naming & state:**

- Immutable state classes (`freezed` recommended given the existing stack's code-gen comfort). No `late` mutable fields on a BLoC's state class as a substitute for proper state transitions.
- Stream subscriptions (the token-by-token gRPC stream) are always cancelled in `close()`/`dispose()` — a leaked stream subscription against a long-lived streaming RPC is a real resource leak, not a theoretical one, given this client's whole point is long-lived streams.

---

## 5. Protocol Buffers

- Linted with `buf lint` against the default ruleset — run in CI on every PR that touches `.proto` files.
- `snake_case` field names, `PascalCase` message/service names, `SCREAMING_SNAKE_CASE` enum values with an explicit `_UNSPECIFIED = 0` default for every enum.
- Field numbers are never reused once a proto file has been referenced by a merged PR — `buf breaking` runs in CI against `main` to catch this automatically.
- Every RPC and message gets a one-line doc comment stating its contract, especially around what happens on the edges (`// Returns NOT_FOUND if worker_id has no active heartbeat within registry.stale_threshold`).

---

## 6. Enforcement

| Layer | Mechanism |
| --- | --- |
| Local, pre-commit | Git hook running the format/lint commands above per changed file's language (template in `docs/RUNBOOK.md` once local dev setup lands) |
| CI, per PR | Same checks, plus full test suite, plus `buf breaking` for proto compatibility |
| AI-agent-authored code | The `titan-engineer` skill's Quality Gates are the enforcement layer at generation time — this document is what those gates check *against* for this specific repo |

Anything in this document that a linter can't catch (naming for intent, comment quality, whether an ADR was needed) is checked against `CONTRIBUTING.md §5`'s Definition of Done during self-review.
