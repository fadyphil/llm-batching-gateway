# Glossary

Domain terms used throughout the other docs, defined once here rather than re-explained inline every time.

**Backpressure** — a system's mechanism for signaling "I'm at capacity" instead of silently queuing work indefinitely. Here: the bounded ingress channel (`docs/ARCHITECTURE.md §5`) rejecting new requests immediately once full.

**Batching (dynamic / continuous)** — grouping multiple inference requests together so a model processes them in one forward pass instead of one at a time, trading a small latency cost (waiting to accumulate a batch) for much higher throughput. "Dynamic" means the batch composition and size aren't fixed upfront — they're decided at runtime by triggers like the dual max-size/window-expiry rule in `docs/ARCHITECTURE.md §3`.

**Batch key** — the `(model, priority)` tuple that determines which open batch a request joins. Requests never share a batch across different models or priority tiers.

**BPE (Byte-Pair Encoding)** — the tokenization algorithm that breaks text into subword units by iteratively merging the most frequent adjacent byte/character pairs in a training corpus, producing a fixed vocabulary. The Rust sidecar implements this from scratch — `docs/adr/0002`.

**gRPC** — a binary RPC framework (built on HTTP/2 and Protocol Buffers) used for every internal service boundary in this system. Chosen over REST/JSON for its native support for bidirectional streaming, which the token-by-token completion flow depends on.

**KV cache (key-value cache)** — the per-request memory a transformer model keeps of previously computed attention keys/values, so it doesn't recompute them for every new token. Real KV-cache introspection/paging (à la PagedAttention) is explicitly out of scope for this project — `docs/PRD.md §5`.

**Priority tier** — `INTERACTIVE` vs `BACKGROUND`, part of the batch key. Isolates latency-sensitive traffic from bulk traffic without one starving the other.

**Property-based testing** — testing an invariant against a wide, randomly generated range of inputs (via a generator) rather than a fixed set of hand-picked examples. Used specifically for the Scheduler's concurrency-sensitive correctness properties — `docs/TESTING.md §2`.

**Quantization** — reducing a model's numerical precision (e.g., 16-bit to 4-bit weights) to shrink memory footprint and speed up inference, at some cost to output quality. Necessary here given the 4GB VRAM constraint.

**Session affinity (sticky routing)** — routing repeat requests from the same session to the same Worker when possible, rather than round-robining. `docs/adr/0001`.

**Slot (in `llama-server`)** — llama.cpp's internal unit of concurrent generation capacity; `--parallel N` configures how many requests `llama-server` can process concurrently. The Worker service proxies into these slots rather than reimplementing batching itself — `docs/ARCHITECTURE.md §6`.

**Walking skeleton** — the smallest possible end-to-end implementation of a system (every component present, doing almost nothing) built first to prove the integration works, before any real logic is added. M0's exit criteria, `docs/milestones/M0-foundation.md`.
