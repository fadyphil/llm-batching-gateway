# Runbook — Local Development

**Status: scaffold.** This file's structure is defined now so it exists and gets filled in as each piece lands, rather than being retrofitted from memory later. Populate each section as its infrastructure is actually built — don't fill in commands that don't run yet.

## Prerequisites

*(TBD at M0 — Go version, Rust toolchain version, Flutter SDK version, `buf` CLI, `protoc` plugins)*

## Running the Full Stack

*(TBD once `docker-compose.yml` exists — target milestone: M0/M1 boundary)*

```bash
# placeholder — real command lands once docker-compose.yml is written
docker-compose up
```

## Running Services Individually (local dev loop)

*(TBD at M0 — `go run` per service, `cargo run` for the tokenizer, `flutter run` for the client)*

## Running Tests

*(TBD at M0 — per-language test commands, referencing `docs/TESTING.md`)*

## Running the Demo Scenario

*(TBD at M-UI — the specific sequence: start the stack, open the Playground, send N concurrent requests, observe batching in the Ops dashboard, kill a Worker to observe recovery)*

## Troubleshooting

*(TBD — filled in as real issues are hit during development, not speculated upfront)*
