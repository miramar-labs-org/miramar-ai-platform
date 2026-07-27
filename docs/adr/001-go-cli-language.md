# ADR-001: CLI implementation language: Go

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

The project needed to choose an implementation language and toolchain for the `miramar`
CLI before any code was written.

## Decision

Go. Chosen because it's the ecosystem norm for Kubernetes-adjacent tooling (`kubectl`,
`helm`, `k9s`, `terraform` are all Go); `client-go` and the Helm SDK let the CLI drive
the Kubernetes API and invoke Helm natively instead of shelling out; it compiles to a
single static binary per OS/arch, which matters for friction-free distribution to
enterprise buyers who don't want to install a Python/Node runtime for a
security-adjacent CLI; and Cobra's subcommand dispatch maps directly onto the
`doctor`/`new`/`deploy`/`validate`/`undeploy` shape (see
[`docs/architecture.md`](../architecture.md#components)).

## Consequences

Native Kubernetes API and Helm SDK access without shelling out to `kubectl`/`helm`
subprocesses. Single static binary simplifies distribution (no runtime dependency to
document or troubleshoot). Commits the project to Go's ecosystem and tooling for future
CLI work.
