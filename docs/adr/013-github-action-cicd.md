# ADR-013: CI/CD: official reusable GitHub Action

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

Users need a way to run `deploy`/`validate`/`uninstall` from CI without hand-rolling
workflow steps.

## Decision

An official reusable GitHub Action wrapping `deploy`/`validate`/`uninstall`, living in
this repo.

## Consequences

A candidate for early standalone publication as `miramar/deploy-action`, ahead of CE's
own maturity — people may adopt the Action before the CLI itself reaches broad adoption,
which is a valid adoption funnel in its own right (see the Adoption milestones section
of [`ROADMAP.md`](../../ROADMAP.md)).
