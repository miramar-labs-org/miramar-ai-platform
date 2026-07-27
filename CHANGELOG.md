# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- `miramar doctor` — first vertical slice: read-only preflight checks for Kubernetes
  API connectivity, cluster distribution (informational), schedulable GPU, target
  namespace, and Helm release-storage access. Storage-class and required-secret checks
  are not yet implemented — see `docs/CURRENT_STATE.md`.
- Colored status output (`doctor`, `deploy`, `validate`, `uninstall`, and top-level
  errors) — green ✓ / yellow ⚠ / red ✗ on a real terminal, falling back to the original
  plain ASCII text when output is piped/redirected or `NO_COLOR` is set.

## [0.1.0] - 2026-07-26

### Added

- `miramar init` — copy a deployment template (`templates/<type>/`) to a local
  directory for customization, patching `model.id`/`model.servedName` in the copied
  `values.yaml` via a structured YAML field patch that preserves comments and key
  ordering.
- `miramar deploy` — deploy one model endpoint via a Helm chart, sourced either from an
  embedded template (`--type serving-vllm`, the default) or a customized local copy
  (`--chart-dir`). Validated end to end on self-hosted k3s (DGX Spark) with both paths.
- `miramar validate` — poll for a ready pod, check `GET /v1/models`, and smoke-test
  `POST /v1/chat/completions` against a small built-in prompt set.
- `miramar uninstall` — remove the Helm release; `--purge-namespace` additionally
  deletes the namespace.
- Template-factory pattern (`templates/<type>/` + `templates/embed.go` +
  `internal/template`) so adding a second serving runtime or model type later only
  requires a new template directory, not new factory logic.
- Direct dependency on `helm.sh/helm/v3` (and its `k8s.io/{api,apimachinery,cli-
  runtime,client-go}` transitive pins) for native Helm SDK install/upgrade/uninstall,
  replacing any need to shell out to the `helm` binary.

### Fixed

- `serving-vllm` template's default `vllm.gpuMemoryUtilization` lowered from vLLM's own
  default of `0.90` to `0.75` — on DGX Spark's GB10 unified memory, only ~105 GiB of
  ~122 GiB is actually free at boot, and `0.90` caused vLLM to refuse to start.
- `serving-vllm` template's pod spec gained a `startupProbe` to gate readiness/liveness
  until the endpoint responds, replacing a fragile fixed `livenessProbe.
  initialDelaySeconds` that couldn't account for cold-start time varying widely by
  model size (a few minutes for a 1.5B model, 15+ minutes for a 7B model on first
  pull).
