# Current State

This document is a concise handoff for maintainers and coding agents. It records the present implementation state, immediate priority, and important validation gaps.

It is intentionally transient:

* `ROADMAP.md` remains the source of truth for planned releases.
* `PRINCIPLES.md` remains the source of truth for durable project philosophy.
* `docs/architecture.md` describes the current system shape.
* `docs/adr/` records accepted architectural decisions.
* This file records what is true **right now** and what should happen next.

Update this file when a meaningful implementation milestone is completed or when the immediate development priority changes.

## Current release

Latest tagged release: `v0.2.0` (doctor's flagship checkset; see `CHANGELOG.md`)

Release 0.1 is complete according to `ROADMAP.md`. Release 0.2's flagship `doctor`
capability shipped in `v0.2.0`; the remaining Release 0.2 items (secret checks, config
schema, failure diagnostics, installation profiles) are still open — see "Immediate
priority" below.

Implemented commands:

* ✅ `miramar new`
* ✅ `miramar deploy`
* ✅ `miramar validate`
* ✅ `miramar undeploy`
* ✅ `miramar doctor` (flagship Release 0.2 check set — see "Immediate priority" below
  for what remains)

## Current golden path

The validated workflow is:

```bash
miramar doctor
miramar deploy
miramar validate
miramar undeploy
```

A customized deployment can be generated and deployed with:

```bash
miramar new --model <model-id> --dir ./my-model
miramar deploy --chart-dir ./my-model
miramar validate
miramar undeploy --purge-namespace
```

See `README.md` for the user-facing workflow and `docs/supported-configurations.md` for the exact validated environment.

## Validated implementation

The following have been validated end to end on the author’s self-hosted k3s cluster running on DGX Spark:

* Embedded default vLLM deployment
* Qwen2.5-1.5B-Instruct zero-flag path
* Customized chart-directory deployment
* Mistral-7B-Instruct-v0.3 customized path
* Endpoint readiness polling
* `GET /v1/models`
* `POST /v1/chat/completions`
* Release-only undeploy
* Namespace-purge undeploy
* GPU memory release after teardown
* `miramar doctor`'s full v0.2 check set: API connectivity, distribution detection,
  schedulable-GPU detection, storage-class discovery, ingress-controller discovery,
  observability-integration discovery, target-namespace check, Helm release-storage
  access

Do not infer validation on other Kubernetes distributions from this list.

## Compatibility boundary

Community Edition is designed to deploy to any existing Kubernetes cluster reachable through the user’s current `kubectl` context.

Currently:

* k3s on DGX Spark is validated.
* GKE, EKS, AKS, kind, and other conformant Kubernetes environments are compatible by design but are not yet validated unless explicitly listed in `docs/supported-configurations.md`.
* Cluster provisioning and lifecycle management are outside CE scope.

See ADR-007 for the accepted deploy-versus-provision boundary.

## Immediate priority

`miramar doctor` is implemented and validated against a live cluster (`internal/doctor`,
wired into `cmd/miramar/doctor.go`), and is the flagship capability of Release 0.2 — the
first command a new user should run, before `deploy` touches anything. It is read-only
and covers:

1. Kubernetes API connectivity (server version, reachability) — can FAIL
2. Cluster distribution — best-effort, informational only, never gates
3. Schedulable GPU — sums allocatable `nvidia.com/gpu` across nodes — can FAIL
4. Storage classes — discovers what's registered via `StorageV1().StorageClasses()`;
   WARN if none found or if the list call itself errors (e.g. an RBAC gap), informational
   otherwise. No template declares `storageClassName` yet, so this exists to demonstrate
   platform awareness ahead of a template that needs one, not because anything currently
   depends on it — a list error here is never a FAIL.
5. Ingress controller — discovers registered `IngressClass`es; WARN if none found or if
   the list call errors. The golden path doesn't expose an Ingress today, so this is
   informational only and never FAILs.
6. Observability integrations — discovers metrics-server, Prometheus Operator, and
   OpenTelemetry Operator via `Discovery().ServerResourcesForGroupVersion`; WARN if none
   found. Per [ADR-016](adr/016-observability-boundary.md), this is discovery-only — CE
   never installs or manages an observability stack, so absence is never a FAIL.
7. Target namespace — WARN if missing, since `deploy` creates it automatically; FAIL if
   the `Get` call itself errors (e.g. an RBAC gap), since that signals a permissions
   problem that would also block `deploy` — unlike checks 4–6, this is a deploy-path
   authorization check, not pure discovery.
8. Helm release-storage access — exercises the same `action.Configuration` path
   `deploy` uses, catching RBAC gaps a plain API-connectivity check would miss — can FAIL,
   but is skipped (WARN) instead of run for real when check 7 has already confirmed the
   target namespace doesn't exist yet, since some clusters only grant namespace-scoped
   RBAC once the namespace exists, and `deploy` creates the namespace itself.

Only checks 1, 3, 7, and 8 (API connectivity, schedulable GPU, target-namespace access,
Helm release-storage access) can FAIL; checks 2, 4, 5, and 6 are pure discovery and
report PASS/WARN at worst, even when the underlying list/discovery call errors. Check 8
never runs for real (and thus never FAILs) on a fresh cluster where the target namespace
doesn't exist yet — this keeps a fresh-cluster `doctor` run from hard-failing before
`deploy` ever gets a chance to create that namespace.

Deliberately still deferred: required-secret checks. No concrete prerequisite yet — the
only secret hook (`hfTokenSecretName`) is commented out and unwired. Add this once a
template actually depends on it, rather than validating something nothing yet needs.

Next candidates for further `doctor` work or the rest of Release 0.2:

* Secret checks, once there's a concrete dependency to check
* `--json` output (deferred from this slice — see open questions below)
* Configuration file schema (`config/*.yaml`)
* Better failure diagnostics when `deploy`/`validate` fail
* Installation profiles (e.g. `--profile gke-l4`)

Vendor-specific observability adapters (W&B, LangSmith, and similar) are explicitly out
of scope for `doctor` and for CE's core observability surface — see `ROADMAP.md`
Release 0.4 for the adapter boundary. A missing vendor adapter must never produce a
FAIL, the same discovery-only rule as the standards-based checks above.

## Constraints for `doctor`

* Read-only by default
* No cluster provisioning
* No cluster-distribution allowlist
* Clear PASS, WARN, and FAIL results
* Actionable remediation text
* Machine-readable output should be considered, but should not delay the first useful vertical slice unless already required
* Any future “apply fixes” behavior must be explicit and opt-in
* Checks should be individually testable
* Distribution-specific checks should extend generic checks rather than replace them
* Observability checks, when added, should discover existing user-managed systems rather than installing or managing them
* Observability checks should prefer open standards such as Kubernetes APIs, Prometheus-compatible metrics, and OpenTelemetry over vendor-specific integrations

## Known open implementation questions

These are not accepted decisions unless promoted into ADRs:

* Resolved for the first slice: doctor checks live in a single `internal/doctor` package
  mirroring `internal/validator`'s shape (`Options` in, `Run(ctx, opts) (*Report, error)`
  out); Helm is checked through the project's existing SDK integration
  (`action.Configuration` + a read-only `action.NewList`), not an external `helm`
  executable, since `deploy` never shells out to one; the schedulable-GPU check sums
  allocatable `nvidia.com/gpu` across all nodes.
* Whether the first release needs JSON output — still deferred, no consumer needs it yet
* Whether missing optional secrets are WARN or FAIL — open, since no secret check exists
  yet (see "Immediate priority")
* How installation profiles interact with generic doctor checks
* Resolved: observability discovery (storage classes, ingress controller, standard
  observability integrations) belongs in baseline `doctor` output, not behind a flag —
  all discovery-only, WARN at most, per ADR-016
* Whether/how `doctor` should surface configured vendor observability adapters
  (W&B, LangSmith) as informational context — open, deferred to Release 0.4 alongside
  the adapters themselves (see `ROADMAP.md`)

Resolve implementation-local questions in code and tests. Create an ADR only when a decision is architecturally significant or difficult to reverse.

## Handoff log

### 2026-07-26 (latest)

* Enterprise Edition planning and implementation moved into a separate private
  repository, outside this public repo entirely. See the
  [Community Edition scope boundary](../ROADMAP.md#community-edition-scope-boundary)
  for what stays out of CE.
* This public CE repo's git history was reset to a single root commit
  (`chore: reset public CE history`) at the point of the split, so no pre-split commit
  ever referencing EE-adjacent content remains reachable in this now-public repo's
  history. Verified nothing was lost in that reset: the reset commit's full 71-file tree
  matches the current working tree exactly, including this session's own work (the
  `checkHelmReleaseStorage` namespace-missing fix, the Go 1.25.12 toolchain bump).
* Verified no EE planning/implementation detail leaked into CE: the CE `docs/adr/`
  index cleanly skips the EE-only ADR numbers with no dangling links, and
  `ROADMAP.md`'s Community Edition scope boundary section states plainly that detailed
  commercial-extension planning lives outside this repo.
* Going forward: any new planning/implementation doc for EE-owned operational
  capabilities (see [ADR-006](adr/006-ce-ee-operational-scope-boundary.md) and
  [ADR-016](adr/016-observability-boundary.md) for the boundary) belongs in the EE
  repo, not here. This CE repo is public; treat every file added to it as public by
  default.

### 2026-07-26 (even later)

* External review implemented: fixed a fresh-cluster failure mode where
  `checkHelmReleaseStorage` ran an unconditional real Helm release-storage (Secret)
  listing even when `checkNamespace` had already determined the target namespace
  doesn't exist yet — on clusters granting only namespace-scoped RBAC (available once
  the namespace exists), that listing could Fail before `deploy` ever gets a chance to
  create the namespace, contradicting the README's "run `doctor` first" golden path.
* `checkNamespace` now returns whether the namespace was confirmed missing;
  `checkHelmReleaseStorage` skips the real check (WARN) when it was. Helm storage still
  FAILs when the namespace exists but access is genuinely broken.
* Added `TestCheckHelmReleaseStorage_SkipsWhenNamespaceMissing` and
  `TestFreshCluster_NamespaceAndHelmStorageNeverFail` to lock in the behavior.

### 2026-07-26 (later)

* External review implemented: README tagline/positioning, doctor-first golden path
  framing across README/quickstart, `docs/architecture.md`'s new "Extensibility"
  section (templates as the extension mechanism), `PRINCIPLES.md`'s "Observe, don't
  own" wording, and `ROADMAP.md`'s doctor-as-flagship framing for Release 0.2.
* `internal/doctor` grew three new discovery-only checks: storage classes, ingress
  controller, and observability integrations (metrics-server, Prometheus Operator,
  OpenTelemetry Operator via API-group discovery) — all WARN-at-most, never FAIL, per
  ADR-016. Validated against the live DGX Spark k3s cluster.
* Added Release 0.4 roadmap wording for W&B/LangSmith as optional, non-blocking vendor
  adapters — explicitly not core observability primitives.

### 2026-07-26

* Release 0.1 scope completed.
* CE scope boundary clarified: CE deploys to any existing cluster; cluster lifecycle is
  outside this public repo.
* `PRINCIPLES.md` and CE-facing ADRs are present.
* `miramar doctor`'s first vertical slice implemented (`internal/doctor`) and validated
  against the live DGX Spark k3s cluster: API connectivity, distribution detection,
  schedulable GPU, target namespace, Helm release-storage access. Storage-class and
  secret checks deliberately deferred — no template depends on either yet. Unit tests
  cover the fakeable checks via `client-go/kubernetes/fake`.
* Next recommended implementation focus: storage-class/secret checks once a template
  needs them, or the rest of Release 0.2 (config file schema, failure diagnostics,
  installation profiles).

## Updating this file

Keep this document concise.

When updating it:

* Replace stale current-state material rather than appending an indefinite diary.
* Keep only a small handoff log of major recent transitions.
* Move durable decisions into an ADR.
* Move future commitments into `ROADMAP.md`.
* Move supported-environment claims into `docs/supported-configurations.md`.
