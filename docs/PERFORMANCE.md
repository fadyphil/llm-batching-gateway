# Performance & Benchmarks

**Status: scaffold — no measured numbers yet.** This file exists now so its structure is defined against `docs/PRD.md §4`'s targets, but every number below is filled in from real `ghz` load-test runs during M5, not estimated. A performance doc with invented numbers is worse than an empty one.

## Test Environment

*(TBD at M5 — exact hardware: RTX 3050 4GB VRAM, model + quantization used, `llama-server` flags)*

## Methodology

*(TBD at M5 — `ghz` configuration, concurrency levels tested, duration per run)*

## Results

| Metric | Target (`docs/PRD.md §4`) | Measured | Notes |
| --- | --- | --- | --- |
| Dispatch decision latency (interactive priority) | Within configured window | *TBD* | |
| Max sustained concurrent streams without request loss | *TBD — hardware-bound, this run finds the real ceiling* | *TBD* | |
| Batch size distribution under load | N/A — descriptive | *TBD* | Direct evidence for the batching thesis, `docs/ARCHITECTURE.md §9` |
| Worker-crash recovery time (requeue to completion) | Zero silent loss | *TBD* | `docs/adr/0004` |

## Known Limitations

*(TBD — honest gaps between what was tested and what a production system would need, e.g. single-machine only, no network-partition testing)*
