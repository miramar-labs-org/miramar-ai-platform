# Architecture (target, v0)

This describes the architecture for the first golden path only — deploying and
validating one model endpoint on an existing GPU-enabled Kubernetes cluster. It is
deliberately scoped smaller than a general-purpose AI platform architecture; broader
capabilities are deferred until this narrow path is working and validated by real usage
(see [`ROADMAP.md`](../ROADMAP.md)).

`deploy`, `validate`, `uninstall`, `init`, and `doctor` are implemented and validated
against a live cluster (see
[`docs/supported-configurations.md`](supported-configurations.md)). `doctor` covers API
connectivity, distribution, schedulable GPU, storage classes, ingress controller,
observability integrations (see [ADR-016](adr/016-observability-boundary.md)), target
namespace, and Helm release storage. Only API connectivity, schedulable GPU, and the two
deploy-path authorization checks (target namespace access, Helm release-storage access)
can FAIL; storage-class, ingress-controller, and observability-integration discovery are
informational or warning-level only, even on a list error. If the target namespace
doesn't exist yet, the Helm release-storage check is skipped (WARN) rather than run for
real, since `deploy` creates the namespace itself and some clusters only grant
namespace-scoped RBAC once it exists. Required-secret checks are
not yet implemented, since no template currently declares one.

## Components

```
User
  │
  ▼
miramar CLI
  │
  ├── doctor      — read-only check of cluster prerequisites
  ├── init        — copies a deployment template to a local directory for customization
  ├── deploy      — deploys one model endpoint via a Helm chart
  ├── validate    — confirms the endpoint is healthy and serving requests
  └── uninstall   — tears the deployment back down
        │
        ▼
Kubernetes API (existing cluster, brought by the user)
        │
        ▼
vLLM (serving runtime)
        │
        ▼
OpenAI-compatible HTTP endpoint
```

## Extensibility

**Templates are Miramar's extension mechanism.** `templates/<type>/` (loaded through
`internal/template`, see [ADR-003](adr/003-helm-chart-template-factory.md)) is the one
place a new serving runtime, deployment shape, or future workload type plugs in — not a
plugin API, not CRDs, not a workflow engine. `deploy --type <name>` and
`init --type <name>` already take the template type as a value, so adding
`serving-sglang`, `serving-triton`, or another future template is adding a
directory under `templates/`, not new factory logic. This keeps the CLI's surface area
fixed while what it can deploy grows.

## Design principles for v0

- **The user brings the cluster.** This project does not provision Kubernetes,
  GPU drivers, or node pools in v0 — it assumes a working GPU-enabled cluster already
  exists and the user has `kubectl` access to it.
- **One golden path, done well, beats broad coverage done thinly.** v0.1 targets exactly
  one serving runtime and one deployment shape rather than trying to support every
  combination of cloud, GPU, and runtime from day one. See
  [`docs/supported-configurations.md`](supported-configurations.md).
- **Every step is independently verifiable.** `doctor` and `validate` exist specifically
  so a user (or this project's own CI) can confirm state at each stage rather than
  trusting that a `deploy` succeeded just because the command didn't error.
- **Thin wrapper, not a reimplementation.** The CLI is intended to invoke existing,
  well-understood primitives (Kubernetes API, Helm) rather than reimplement scheduling,
  networking, or orchestration logic itself.
- **Namespace-scoped, not cluster-owning.** All v0.1 resources live in a single
  namespace (`miramar-ai-platform`) so this can run alongside other workloads on a
  shared cluster — it never assumes it owns the whole cluster.

These are v0.1-scoped implementation principles. For durable, project-wide philosophy
that outlives any one release, see [`PRINCIPLES.md`](../PRINCIPLES.md).

## Decisions

Rationale for each decision below has been extracted into individual ADRs — see
[`docs/adr/`](adr/README.md) for the full Context/Decision/Consequences record of each.

| Decision | ADR |
| --- | --- |
| CLI implementation language: Go | [ADR-001](adr/001-go-cli-language.md) |
| Serving runtime: vLLM | [ADR-002](adr/002-vllm-serving-runtime.md) |
| Deploy via Helm using a template-factory pattern | [ADR-003](adr/003-helm-chart-template-factory.md) |
| GPU memory headroom tuned empirically, not assumed | [ADR-004](adr/004-gpu-memory-headroom-tuning.md) |
| Cold starts handled via `startupProbe`, not a fixed delay | [ADR-005](adr/005-startup-probe-for-cold-starts.md) |

The v0.1 cluster target itself — self-hosted k3s on the author's DGX Spark — is recorded
in [`ROADMAP.md`](../ROADMAP.md#release-01--package-what-already-works); why CE's scope
isn't actually k3s-specific is [ADR-007](adr/007-kubernetes-api-abstraction.md).
