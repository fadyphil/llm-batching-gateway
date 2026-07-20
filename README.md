# LLM Inference Gateway

*A batching scheduler, polyglot service mesh, and failure-recovery story for LLM inference — the infrastructure is the project, not the model. See `docs/PRD.md §1` for the full thesis.*

<!-- badges: TBD once CI exists — build status, license -->

**Status:** Pre-implementation — see [`docs/ROADMAP.md`](docs/ROADMAP.md) for current milestone progress.

<!-- demo gif / video link: TBD once M-UI + M1 land -->

## What This Is

A gRPC-based inference gateway in Go, fronting a local `llama.cpp` backend, with:
- A batching Scheduler that groups concurrent requests by model and priority, with provable correctness properties (`docs/TESTING.md §2`)
- A hand-rolled BPE tokenizer as an independent Rust service, not a library call (`docs/adr/0002`, `docs/adr/0003`)
- Session-affinity routing, backpressure, and at-least-once failure recovery — all demonstrated, not just claimed (`docs/adr/0001`, `docs/adr/0004`)
- A Flutter client that makes the batching and failure recovery visible in real time

## Documentation

| Doc | What's in it |
|---|---|
| [`docs/PRD.md`](docs/PRD.md) | Requirements, MVP scope, functional/non-functional requirements, what's explicitly out of scope |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | System design, request lifecycle, the Scheduler's internals, and why each major piece is shaped the way it is |
| [`docs/SCHEMA.md`](docs/SCHEMA.md) | Every data shape — wire contracts and internal state — as pure reference |
| [`docs/TESTING.md`](docs/TESTING.md) | Test pyramid, the property-based testing strategy, chaos/load approach |
| [`docs/ROADMAP.md`](docs/ROADMAP.md) | External-facing milestone status |
| [`docs/PERFORMANCE.md`](docs/PERFORMANCE.md) | Measured numbers against the NFR targets (populated after load testing) |
| [`docs/GLOSSARY.md`](docs/GLOSSARY.md) | Domain terms used throughout |
| [`docs/adr/`](docs/adr/README.md) | Every architecturally significant decision, with rejected alternatives and consequences |
| [`docs/milestones/`](docs/milestones/) | Per-milestone exit criteria and card breakdown |
| [`docs/RUNBOOK.md`](docs/RUNBOOK.md) | How to run the stack locally (populated as infra lands) |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Ground rules, TDD workflow, Definition of Done — binding on human and AI-agent contributions alike |
| [`docs/BRANCHING.md`](docs/BRANCHING.md) | Git workflow, branch protection, tagging convention |
| [`CODE_STYLE.md`](CODE_STYLE.md) | Enforced style and engineering standards, per language |

## Tech Stack

Go (Gateway, Auth, Scheduler, Registry, Worker, Observability) · Rust (Tokenizer, `tonic`) · Flutter (client) · gRPC/Protocol Buffers · nginx · `llama.cpp`

## Getting Started

See [`docs/RUNBOOK.md`](docs/RUNBOOK.md) — populated once local dev infrastructure exists (M0).

## License

[MIT](LICENSE)
