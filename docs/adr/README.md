# Architecture Decision Records — Index

Every architecturally significant decision in this project, in one place. Status meanings and the full lifecycle are defined in `CONTRIBUTING.md §7`. Add a row here in the same PR that adds or supersedes an ADR — this file is the source of truth for "what did we decide and why," not something reconstructed by reading every file in the folder.

| # | Title | Status | Milestone |
| --- | --- | --- | --- |
| [0001](0001-session-affinity-in-memory-map.md) | Session affinity via in-memory sticky map, not Redis | Accepted | M1 |
| [0002](0002-hand-rolled-bpe-over-crate.md) | Hand-rolled BPE tokenizer over the `tokenizers` crate | Accepted | M-Tok |
| [0003](0003-rust-tokenizer-as-grpc-service.md) | Rust tokenizer as a separate gRPC service, not an FFI binding | Accepted | M-Tok |
| [0004](0004-at-least-once-retry-semantics.md) | At-least-once delivery on worker-crash retry, not exactly-once | Accepted | M3 |
| [0005](0005-github-flow-over-gitflow.md) | GitHub Flow over GitFlow / full trunk-based development | Accepted | M0 |
| [0006](0006-kanban-over-scrum.md) | Milestone-anchored Kanban over Scrum | Accepted | M0 |
| [0007](0007-monorepo-over-polyrepo.md) | Monorepo over per-service polyrepo | Accepted | M0 |
| [0008](0008-batch-key-as-struct.md) | batchKey as struct, not string concatenation | Proposed | M1 |
| [0009](0009-in-memory-registry-for-m1.md) | In-memory worker registry for M1, separate gRPC service deferred to M3 | Proposed | M1 |

**"Accepted"** means the ADR is written and current. A reversed decision gets a *new* ADR that supersedes the old one — the old row stays, its Status column updates to `Superseded by 00XX`, per `CONTRIBUTING.md §7`.
