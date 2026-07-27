# Roadmap

This roadmap describes the intended path from an empty scaffold to a working v0.1
release that delivers one golden path end to end:

> Install and validate an OpenAI-compatible model endpoint on an existing GPU-enabled
> Kubernetes cluster.

Each release section lists what "done" looks like so scope doesn't drift while building
it.

Some items below are informed by patterns already proven out in a separate, personal
reference-architecture repository (a hybrid GCP/on-prem platform this project's author
has been building and operating). Where noted, the intent is to **learn from that
architecture's shape**, not to copy its code directly — this project starts from an
empty repo and only pulls in what's generalizable to an arbitrary user's cluster.

## Release 0.1 — package what already works

Goal: a single documented, reproducible path for one deployment target.

- ✅ Product README and supported-configuration statement (this scaffold)
- ✅ Choose the CLI implementation language/toolchain — **Go** (see
      [`docs/architecture.md`](docs/architecture.md#decisions))
- ✅ Pick one serving runtime to support first and one cluster target — **vLLM** on
      **self-hosted k3s (DGX Spark), namespace `miramar-ai-platform`** (see
      [`docs/architecture.md`](docs/architecture.md#decisions)). Other cluster
      distributions (GKE, kind, EKS, AKS) are CE-scope for `deploy` and compatible by
      design — v0.1 has just only validated the one; see the
      [Community Edition scope boundary](#community-edition-scope-boundary) below
      for the deploy-vs-provision boundary
- ✅ `miramar init` — copy a deployment template (`templates/<type>/`) to a local
      directory for customization, patching `model.id`/`model.servedName` in the copied
      `values.yaml` without disturbing comments or unrelated fields. Pulled forward from
      Release 0.3 (see the note on that release's `init` bullet below) once the
      template-factory pattern was chosen for `deploy --type` too.
- ✅ `miramar deploy` — deploy one model endpoint via a Helm chart. Chart sourcing uses
      the same template-factory pattern as `init`: `--type serving-vllm` (default) loads
      the embedded template, or `--chart-dir <path>` loads a customized copy produced by
      `init`. Validated end to end on self-hosted k3s (DGX Spark) with both the
      zero-flag default (Qwen2.5-1.5B-Instruct) and a customized `--chart-dir` copy
      (Mistral-7B-Instruct-v0.3).
- ✅ `miramar validate` — smoke-test the endpoint (poll for a ready pod → `GET
      /v1/models` → `POST /v1/chat/completions` for a small built-in prompt set → check
      responses). The deploy-then-poll-then-smoke-test shape mirrors a pattern already
      validated in the author's reference platform, generalized to not assume any
      particular org/cluster naming.
- ✅ `miramar uninstall` — clean teardown; `--purge-namespace` additionally deletes the
      namespace. Both the default (release-only) and `--purge-namespace` paths validated
      against the live cluster, including confirming GPU memory is freed afterward.
- ✅ First tagged release with pinned dependency versions — **v0.1.0**. Dependencies
      pinned via `go.mod`/`go.sum` (no floating versions); see
      [`CHANGELOG.md`](CHANGELOG.md#010---2026-07-26).

## Release 0.2 — improve installation

Goal: reduce the tribal knowledge required before `deploy` works — `doctor` is the
flagship capability of this release: not a minor utility, but the thing that
demonstrates Miramar understands the platform it's deploying to before it touches
anything. It's the first command a new user should run, before `deploy`.

- ✅ `miramar doctor` — checks Kubernetes API reachability, reports distribution as
      information only, verifies schedulable GPU resources, and discovers storage
      classes, an ingress controller, and standard observability integrations
      (metrics-server, Prometheus Operator, OpenTelemetry Operator — see
      [ADR-016](docs/adr/016-observability-boundary.md)) as informational/warning-level
      context, then checks the target namespace and exercises Helm release-storage
      access through the same SDK path used by `deploy`. Only API reachability,
      schedulable GPU, and the two deploy-path authorization checks (target namespace
      access, Helm release-storage access) can FAIL; storage-class, ingress-controller,
      and observability-integration discovery are informational only and report WARN at
      worst, even on a list error. Helm release-storage access is skipped (WARN) rather
      than checked for real when the target namespace doesn't exist yet, since some
      clusters only grant namespace-scoped RBAC once the namespace exists and `deploy`
      creates the namespace itself. It asks "can I talk to
      Kubernetes, can I schedule a GPU pod, do the prerequisites exist," never "is this
      cluster type supported," so it works the same way regardless of who provisioned
      the cluster.
- ⏳ Required-secret checks, once a template actually declares one to check for.
- ⏳ Configuration file schema (`config/*.yaml`) instead of hardcoded values
- ⏳ Better failure diagnostics when `deploy`/`validate` fail
- ⏳ Installation profiles (e.g. `--profile gke-l4`)

## Release 0.3 — improve developer experience

Goal: make it easy to go from "one model works" to "my model works."

- ✅ Project/config generator for a new model deployment (`miramar init`) — delivered
      ahead of schedule in Release 0.1 as part of the template-factory pattern; see that
      release's `init` bullet above
- ⏳ Standard metadata for a deployed model (name, runtime, resource footprint)
- ⏳ Support a second serving runtime
- ⏳ Documented, tested upgrade path between CLI versions

CE can already deploy to any existing cluster in principle — see the
[Community Edition scope boundary](#community-edition-scope-boundary)
below: deploying to a cluster someone already has is CE scope regardless of who
provisioned it; only *provisioning* one is EE. What's still open is which additional
distribution gets *validated* end to end next — GKE is the likely candidate, since it's
already compatible by design — but that's not committed to this release specifically.

## Release 0.4 — design-partner capabilities

Goal: capabilities that come directly from real deployments with early users, not
speculative feature-building.

- ⏳ `miramar support-bundle` — collect logs/config for troubleshooting
- ⏳ Observability integration surfaces for clusters the user already operates: expose
      Prometheus-compatible metrics where available, prefer OpenTelemetry for future
      trace/log correlation, document optional scrape hooks, and provide example
      dashboards or alerts without installing or managing the observability stack
      itself. See [ADR-016](docs/adr/016-observability-boundary.md).
- ⏳ Optional vendor adapters for W&B, LangSmith, and similar tools: CE may document how
      to pass user-owned credentials/configuration into workloads, but these integrations
      must remain optional and non-blocking — they are adapters on top of the core
      observability surface above, not observability primitives themselves. Managed
      organization/team integrations, secrets, retention, policy, and support are outside
      this public CE repo. `doctor` may later detect configured adapters as informational
      context, but a missing W&B/LangSmith integration must never produce a FAIL.
- ⏳ Basic policy controls (e.g. resource limits, allowed namespaces)
- ⏳ Documented, tested upgrade procedure between platform versions
- ⏳ Whatever the first two or three real users actually ask for — deliberately left
      unspecified until there are real users to ask

## Adoption milestones

Tracked separately from the engineering releases above — these are what actually
validate the product, not code shipped. Left unchecked until they happen for real, with
no target dates (this is a solo/early-stage project, not a funded startup with a sales
calendar):

- ⏳ First external user installs the CLI and completes the golden path
      (`init` → `deploy` → `validate` → `uninstall`) outside the author's own
      environment
- ⏳ First design partner — someone using CE for real work, giving ongoing feedback
- ⏳ First deployment outside the author's lab
- ⏳ First adoption of the `deploy` GitHub Action independent of the CLI itself
- ⏳ First unprompted bug report or feature request from someone who isn't the author

The qualitative bar behind all of Release 0.1–0.4: a new user should be able to install
the CLI, deploy a small model, validate it, send one inference request, and remove
everything again in well under an hour, with clear progress at every step. Equally
important — what convinces a senior AI platform engineer in the first five minutes on
the repo (README, quickstart, one clear demo) — matters as much as any single feature
on the engineering releases above.

## Community Edition scope boundary

This public repository is Community Edition. It deploys and validates AI workloads on an
existing, reachable Kubernetes cluster and deliberately does not own cluster lifecycle.

The durable boundary is recorded in ADRs rather than repeated here:

- [ADR-006](docs/adr/006-ce-ee-operational-scope-boundary.md): operational scope is the
  edition boundary.
- [ADR-007](docs/adr/007-kubernetes-api-abstraction.md): deploying to an existing
  cluster is CE scope; provisioning or managing clusters is outside this repository.
- [ADR-016](docs/adr/016-observability-boundary.md): CE exposes standard signals for
  user-owned observability systems; managed observability operations are outside this
  repository.

Detailed commercial-extension planning, implementation, and distribution decisions
intentionally live outside this public CE repository.

## Explicit non-goals for the near term

Deferred until there's evidence (a real user, a paid pilot, a repeated feature request)
that they're worth building:

- A full Kubernetes operator
- A hosted SaaS control plane
- Managed observability stack lifecycle, retention, alerting, or team-governance
  features in CE
- Multi-cloud support
- A marketplace of templates
- Formal compliance certifications
- Broad "any model, any cloud" claims

Several of the items above (a hosted SaaS control plane, formal compliance
certifications, multi-cloud/fleet management) are exactly the kind of thing the
[Community Edition scope boundary](#community-edition-scope-boundary)
above now accounts for — see that section for the current design record, still subject
to revision.
