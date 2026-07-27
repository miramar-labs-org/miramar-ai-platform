# Principles

[`ROADMAP.md`](ROADMAP.md) is what's being built and when. [`docs/adr/`](docs/adr/README.md)
records why individual decisions were made. This file is different from both: durable,
project-wide philosophy meant to resolve future design debates that neither the roadmap
nor an existing ADR covers yet. Unlike an ADR, it isn't tied to one decision — but like
the rest of this project's documentation, it stays grounded in things already decided
here, not aspirational boilerplate.

- **Kubernetes native.** Target the Kubernetes API itself, not a specific distribution
  or provisioning mechanism — see [ADR-007](docs/adr/007-kubernetes-api-abstraction.md).
- **Self-hosted first.** No forced third-party account or API key in the golden path —
  see [ADR-002](docs/adr/002-vllm-serving-runtime.md)'s vLLM rationale.
- **Simple before complete.** One golden path done well beats broad coverage done
  thinly — see [`docs/architecture.md`](docs/architecture.md#design-principles-for-v0).
- **Opinionated over configurable.** One runtime, one deploy shape, and sane defaults
  come before an options matrix.
- **Developer experience first.** Install, deploy, validate, and remove a model in well
  under an hour, with clear progress at every step — see
  [`ROADMAP.md`](ROADMAP.md#adoption-milestones)'s adoption-milestones bar.
- **Transparent automation.** Every step is independently verifiable (`doctor`,
  `validate`) — never a silent black box that just claims success.
- **Observe, don't own.** CE exposes Kubernetes-native and open-standard observability
  surfaces (Prometheus-compatible metrics, future OpenTelemetry hooks) and may detect
  what a user already runs, but never installs, upgrades, or manages an observability
  stack itself — see [ADR-016](docs/adr/016-observability-boundary.md).
- **Composable, not monolithic.** A thin wrapper over existing primitives (`kubectl`,
  Helm); this public repo stays useful on its own and does not hide core CE behavior
  behind license gates.
- **Portable.** Apache 2.0 for CE, no cloud or vendor lock-in.
