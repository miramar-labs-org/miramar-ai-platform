# ADR-005: Cold starts handled via `startupProbe`, not a fixed delay

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

vLLM's first-run cold start (HF weight download, then CUDA graph capture) varies widely
by model size — a few minutes for a 1.5B model, 15+ minutes for a 7B model on first
pull.

## Decision

Use a Kubernetes `startupProbe` (30 minutes of allowance) to fully suppress
readiness/liveness checks until the endpoint responds, instead of tuning
`livenessProbe.initialDelaySeconds` to a single fixed number.

## Consequences

No per-model tuning of a fixed delay is needed before deploying a new model size — a
fixed number would break for any model larger than whatever it was tuned against. The
30-minute allowance is itself an assumption that may need revisiting for much larger
models in the future (see Release 0.3's second-runtime work in
[`ROADMAP.md`](../../ROADMAP.md)).
