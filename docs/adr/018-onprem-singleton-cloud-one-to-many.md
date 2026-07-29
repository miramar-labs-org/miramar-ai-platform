# ADR-018: On-prem infra is a shared singleton; cloud platforms are one-to-many

**Status:** ✅ Accepted
**Date:** 2026-07-28

## Context

The org's infrastructure splits into two categories with fundamentally different
cardinality, and config/secrets ownership needs to follow that shape rather than
treating both the same way:

- **On-prem / local** (DGX, AGX): a small number of physical machines, shared across
  every repo and product that needs GPU compute. There is effectively one DGX and one
  AGX target per environment — a singleton, not a set.
- **Cloud** (GCP today; AWS/Azure are architecturally in scope per
  [CE ADR-007](007-kubernetes-api-abstraction.md), not yet built): each platform
  product can provision and own its own separate cloud footprint — its own GCP
  project, its own Workload Identity Federation pool/provider, its own service
  accounts, its own Terraform state. Nothing about the cloud side is inherently
  singular the way the DGX box is.

This surfaced concretely while wiring up `miramar-ai-platform-ee`'s Phase 5 CI
installers (see [ADR-009](https://github.com/miramar-labs-org/miramar-ai-platform-ee/blob/main/docs/adr/009-secrets-external-secrets-operator.md)):
GitHub Actions secrets/variables for GCP access (`WIF_PROVIDER`, `GCP_PROJECT_ID`,
`GKE_CLUSTER_NAME`, `GKE_ZONE`, `GCP_SERVICE_ACCOUNT`) had been set at the
**org** level, implicitly assuming one shared cloud platform for the whole org. That
assumption stopped holding the moment a second platform repo (`-ee`) existed — and in
practice, the org-level secret/variable propagation to that new repo didn't even work
reliably in GitHub itself, forcing repo-scoped copies as an immediate fix. The
underlying architectural question — should cloud identity be shared or per-platform —
predates and outlives that specific bug.

## Decision

- **On-prem/local config stays org-scoped.** DGX/AGX host IPs, SSH keys, and
  kubeconfigs represent one real shared physical resource; org-level GitHub Actions
  secrets/variables correctly model that singleton.
- **Cloud config defaults to repo (platform)-scoped, not org-scoped**, even when
  today's value happens to be identical across platforms (e.g. two platforms
  currently both pointing at the same GCP project). Each platform product owns its
  own cloud footprint going forward: its own GCP project/WIF provider/service
  accounts if/when it needs one, set as GitHub Actions secrets/variables on that
  platform's own repo.
- New cloud-facing infra should be provisioned and referenced per-platform from the
  start, not bolted onto a shared org-level identity that was only ever sized for one
  platform.

## Consequences

- A platform repo's CI workflows read cloud identity (WIF provider, project ID,
  cluster name/zone, service account emails) from that repo's own secrets/variables.
  Org-level cloud secrets/variables should be treated as legacy/transitional, not the
  pattern to extend.
- Existing org-level cloud secrets/variables (several `GCP_*`/`GKE_*`/`WIF_PROVIDER`
  entries, originally set when only one platform existed) need an audit to decide,
  per entry, whether it should be migrated to repo scope, left as a genuinely shared
  default, or deleted — not yet done, tracked as follow-up work.
- This does not change how a single platform's own Terraform/SDK code models cloud
  resources internally (see [EE ADR-010](https://github.com/miramar-labs-org/miramar-ai-platform-ee/blob/main/docs/adr/010-native-cloud-sdks-not-terraform.md))
  — it only changes where CI credentials/identifiers for reaching that cloud are
  stored.
