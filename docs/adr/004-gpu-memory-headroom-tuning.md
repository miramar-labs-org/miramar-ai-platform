# ADR-004: GPU memory headroom tuned empirically, not assumed

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

DGX Spark's GB10 uses unified memory shared with the host. vLLM measured ~122 GiB total
but only ~105 GiB free at boot, so vLLM's own default `gpuMemoryUtilization: 0.90`
overcommitted and the engine refused to start.

## Decision

The template default is `gpuMemoryUtilization: 0.75`, leaving headroom for other host
processes, rather than trusting vLLM's own default.

## Consequences

The default model (Qwen2.5-1.5B-Instruct) and a larger validated model
(Mistral-7B-Instruct-v0.3) both fit comfortably at `0.75` — see
[`docs/prerequisites.md`](../prerequisites.md). On a smaller or more heavily-loaded GPU
this value may need lowering further, via `miramar init` + editing `values.yaml`; the
value is not auto-detected per-host in v0.1.
