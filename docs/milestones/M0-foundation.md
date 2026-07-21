# M0 — Foundation

**Status:** ✅ Complete
**Depends on:** Nothing — this is the first milestone
**Unlocks:** M1, M-Tok, M-UI (see `docs/PRD.md §6`)

## Goal

Prove the plumbing end to end before any batching logic exists. A single request round-trips Client → nginx → Gateway → Scheduler → Worker → back to Client, with the Scheduler doing zero batching (dispatch immediately, batch size 1).

## Entry Criteria

None — this milestone starts the project.

## Exit Criteria (Definition of Done for this milestone)

- [ ] All four `.proto` files defined and lint-clean (`buf lint`): `gateway.proto`, `scheduler.proto`, `worker.proto`, `tokenizer.proto` — shapes match `docs/SCHEMA.md §1`
- [ ] Four Go service skeletons stood up (Gateway, Scheduler, Registry, Worker) plus the Auth stub, each in its own `internal/<service>/` package per `docs/CODE_STYLE.md §2`
- [ ] nginx configured with `grpc_pass` routing to the Gateway
- [ ] A single client request streams through the full path and receives real tokens back, with the Scheduler applying no batching logic yet (dispatch on receipt, batch size 1)
- [ ] Every service passes the Definition of Done in `CONTRIBUTING.md §5` — including tests written first, not retrofitted
- [ ] Tagged `v0.1.0-M0` per `docs/BRANCHING.md §6`
- [ ] `docs/ROADMAP.md` status updated to `✅ Complete` for M0

## Cards

Filed as GitHub Issues using `.github/ISSUE_TEMPLATE/feature_task.yml`, milestone set to `M0 - Foundation`:

1. Define `gateway.proto` (`CompletionService`)
2. Define `scheduler.proto` (`SchedulerService`)
3. Define `worker.proto` (`WorkerService`)
4. Define `tokenizer.proto` (`TokenizerService`) — stub only, real implementation is M-Tok
5. Gateway skeleton — accepts a stream, forwards to Scheduler
6. Scheduler skeleton — batch size 1, immediate dispatch, no timer logic yet
7. Worker skeleton — wraps a local `llama-server` instance per `docs/ARCHITECTURE.md §6`
8. Auth stub — static token check (FR-2)
9. nginx config — `grpc_pass` edge routing
10. End-to-end smoke test proving the full round trip

## Relevant PRD / Architecture Sections

- `docs/PRD.md` FR-1, FR-2
- `docs/ARCHITECTURE.md §1`, `§2`, `§6`
- `docs/SCHEMA.md §1`

## Retro

*(filled in when this milestone closes — what shipped, what got cut, one thing to do differently)*
