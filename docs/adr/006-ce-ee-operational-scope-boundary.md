# ADR-006: CE/EE split principle: operational scope, not workload type

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

As the project's scope grows beyond the v0.1 golden path, a principled boundary is
needed between lightweight Community Edition work and operational capabilities that are
outside this public repository. The goal is to avoid deciding scope ad hoc per feature.

## Decision

Split by operational scope, not workload type. Self-contained, single-team,
bring-your-own-cluster work stays CE. Platform-team-style lifecycle ownership and
organizational operations are outside this public CE repository.

## Consequences

Gives a repeatable test for where any new capability belongs, instead of a per-feature
argument each time. The cluster-specific corollary of this principle is spelled out
separately in [ADR-007](007-kubernetes-api-abstraction.md).
