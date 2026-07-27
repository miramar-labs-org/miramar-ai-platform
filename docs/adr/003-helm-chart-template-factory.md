# ADR-003: Deploy via Helm using a template-factory pattern

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

The project needed a deploy mechanism, and a way to source charts that works both for a
zero-flag default deploy and for a customized deploy (different model, different
resource settings) without duplicating chart logic.

## Decision

Deploy via a Helm chart, not plain Kubernetes manifests — trading some up-front
machinery (values, templating) for upgrade/rollback semantics and a head start on
templatizing across the second serving runtime already planned for Release 0.3.

Chart sourcing uses a template-factory pattern, pulled forward from Release 0.3 rather
than deferred. Charts live under `templates/<type>/` (one directory per project type —
today just `templates/serving-vllm/`), embedded into the binary via
`templates/embed.go`. `miramar deploy --type <type>` (default `serving-vllm`) loads a
template directly from the embedded filesystem; `miramar init --type <type> --dir <path>`
copies the same template out to a directory the user owns, optionally patching
`model.id`/`model.servedName` in the copied `values.yaml`; `miramar deploy --chart-dir
<path>` then deploys from that customized copy. Adding a second project type later means
adding a sibling `templates/<type>/` directory and one entry in `templates.Available` —
no change to the factory logic itself.

This mirrors the project-template-factory pattern already proven in the author's
reference platform (`create-project.yaml` + `templates/new-project-<type>/`), with one
deliberate improvement: that reference pattern does raw `sed` placeholder substitution
into plain manifests, while `init`'s value-patching here operates on the `values.yaml`
AST via `gopkg.in/yaml.v3`'s `Node` API — a structured field patch, not text
substitution, so comments and key ordering in the copied file are preserved exactly.

## Consequences

Helm gives upgrade/rollback semantics for free. Adding a new project type is additive
(new template directory + registry entry), not a factory-logic change, which keeps the
cost of Release 0.3's second-runtime work bounded. The AST-based value patch is more
robust than text substitution but is more code to maintain than a `sed` one-liner — a
deliberate trade accepted for correctness on user-owned files.
