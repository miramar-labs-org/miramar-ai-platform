# Architecture Decision Records

Each ADR captures one decision as Context / Decision / Consequences. Once accepted, an
ADR is not edited to reflect new thinking — a later decision that changes course gets
its own new ADR that supersedes the old one, so the history stays honest. For the
running list of what's built vs. planned, see [`../../ROADMAP.md`](../../ROADMAP.md);
for durable project-wide philosophy, see [`../../PRINCIPLES.md`](../../PRINCIPLES.md).

This repo's ADR numbers are their own independent sequence, not a shared range with the
private EE repo's own `docs/adr/` (currently numbered from 008). Before assigning a new
number here, check that EE repo's ADR index too — this repo's own ADR-013 once collided
with EE's ADR-013 on unrelated topics before EE's was renumbered to ADR-017.

| ADR | Title | Status |
| --- | --- | --- |
| [001](001-go-cli-language.md) | CLI implementation language: Go | ✅ Accepted |
| [002](002-vllm-serving-runtime.md) | Serving runtime: vLLM | ✅ Accepted |
| [003](003-helm-chart-template-factory.md) | Deploy via Helm using a template-factory pattern | ✅ Accepted |
| [004](004-gpu-memory-headroom-tuning.md) | GPU memory headroom tuned empirically, not assumed | ✅ Accepted |
| [005](005-startup-probe-for-cold-starts.md) | Cold starts handled via `startupProbe`, not a fixed delay | ✅ Accepted |
| [006](006-ce-ee-operational-scope-boundary.md) | CE/EE split principle: operational scope, not workload type | ✅ Accepted |
| [007](007-kubernetes-api-abstraction.md) | CE deploys to any existing Kubernetes cluster; EE owns cluster lifecycle | ✅ Accepted |
| [013](013-github-action-cicd.md) | CI/CD: official reusable GitHub Action | ✅ Accepted |
| [016](016-observability-boundary.md) | Observability boundary: standard signals in CE, managed operations in EE | ✅ Accepted |
| [018](018-onprem-singleton-cloud-one-to-many.md) | On-prem infra is a shared singleton; cloud platforms are one-to-many | ✅ Accepted |
