# Architecture Decision Records

Each ADR captures one decision as Context / Decision / Consequences. Once accepted, an
ADR is not edited to reflect new thinking — a later decision that changes course gets
its own new ADR that supersedes the old one, so the history stays honest. For the
running list of what's built vs. planned, see [`../../ROADMAP.md`](../../ROADMAP.md);
for durable project-wide philosophy, see [`../../PRINCIPLES.md`](../../PRINCIPLES.md).

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
