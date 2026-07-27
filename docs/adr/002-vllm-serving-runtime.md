# ADR-002: Serving runtime: vLLM

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

The v0.1 golden path needed exactly one serving runtime to support first, rather than
trying to support multiple runtimes from day one.

## Decision

vLLM. It serves an OpenAI-compatible HTTP API (`/v1/chat/completions`) out of the box,
which is exactly the golden-path promise in [`README.md`](../../README.md) — no adapter
layer needed. It's open source and self-hosted with no external account/API-key
dependency, which matters for the target buyer segments (healthcare, defense, finance
procurement teams don't want a new third-party account in the critical path). It also
matches a pattern already proven in the author's reference platform: deploy →
`kubectl rollout status` → port-forward smoke-test against `/v1/chat/completions`.

## Consequences

The golden path is OpenAI-API-compatible from the start, with no translation layer to
build or maintain. No third-party account/API key required anywhere in the golden path,
consistent with the project's self-hosted-first posture (see
[`PRINCIPLES.md`](../../PRINCIPLES.md)). Support for a second serving runtime is
deferred to Release 0.3 (see [`ROADMAP.md`](../../ROADMAP.md)); the template-factory
pattern (see [ADR-003](003-helm-chart-template-factory.md)) is what makes adding one
tractable rather than a rewrite.
