# ADR-007: CE deploys to any existing Kubernetes cluster; EE owns cluster lifecycle

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

Applying [ADR-006](006-ce-ee-operational-scope-boundary.md)'s operational-scope
principle specifically to Kubernetes clusters required a clear line for the single most
consequential category, since it determines whether CE can address the by-far most
common real-world case: an existing, already-provisioned cluster, self-hosted or
cloud-managed.

## Decision

Deploying to an existing, reachable cluster — self-hosted or cloud-managed — is CE
scope. CE treats the cluster as a Kubernetes API endpoint and never cares who created
it; it only needs a reachable API and a schedulable GPU.

Provisioning or managing that cluster's lifecycle is outside this public CE repository.

This boundary was chosen because applications target the Kubernetes API rather than the
mechanism that created the cluster, providing a cleaner architectural separation between
Community and Enterprise editions.

## Consequences

`doctor`'s preflight/platform-detection checks (see [`ROADMAP.md`](../../ROADMAP.md)
Release 0.2) only ever ask "can I talk to Kubernetes, can I schedule a GPU pod, do the
prerequisites exist," never "is this cluster type supported." See
[`docs/supported-configurations.md`](../supported-configurations.md) for the
compatibility philosophy and the validated-vs-compatible-by-design matrix — what's
actually been tested versus architecturally in scope under this boundary.
