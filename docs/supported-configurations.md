# Supported Configurations

This matrix tracks exactly what's been tested, not what's theoretically possible. A
component listed here does not mean it works in general — it means this exact
combination was validated end to end (`deploy` → `validate` → `undeploy`, both the
embedded-default and `new`-produced `--chart-dir` paths).

| Component               | Status |
| ------------------------ | ------ |
| Kubernetes distribution  | ✅ **Validated** — k3s. Other distributions are compatible by design, not yet validated — see the matrix below |
| GPU                      | ✅ **Validated** — NVIDIA GB10 (DGX Spark, unified host/GPU memory) |
| Serving runtime          | ✅ **Validated** — vLLM (`vllm/vllm-openai:latest`) |
| Installation method      | ✅ **Validated** — Helm, via the `miramar` CLI's own Helm SDK integration (embedded template and `new`-produced `--chart-dir` copy both tested) |

Models validated on this configuration: `Qwen/Qwen2.5-1.5B-Instruct` (built-in default)
and `mistralai/Mistral-7B-Instruct-v0.3` (customized copy, see
[`docs/quickstart.md`](quickstart.md)).

## Validated vs. compatible by design

**Compatibility philosophy:** Miramar targets the Kubernetes API. If a distribution
satisfies the documented prerequisites, Miramar should behave identically regardless of
how the cluster was provisioned. Only validated environments are listed as officially
tested. See [ADR-007](adr/007-kubernetes-api-abstraction.md) for the full rationale
behind this boundary.

✅ **Validated** means this exact combination was deployed, tested, and torn down end to
end. **Compatible by design** means CE's architecture supports it — `deploy` only needs
a reachable Kubernetes API and a schedulable GPU, nothing k3s-specific — but it hasn't
been tested here yet; treat it as likely-to-work, not guaranteed.

| Kubernetes distribution               | Status |
| -------------------------------------- | ------ |
| k3s (self-hosted, DGX Spark)           | ✅ **Validated** |
| GKE (existing, user-provisioned)       | 🕒 Compatible by design — not yet validated |
| kind                                    | 🕒 Compatible by design — not yet validated |
| EKS, AKS                                | 🕒 Compatible by design — not yet validated |
| OpenShift                               | ⬜ Not yet considered |

CE deploys to any of these the same way — it only cares whether `kubectl` works and a
GPU is schedulable, never who provisioned the cluster. Provisioning or upgrading a
cluster (as opposed to deploying to one that already exists) is outside this public CE
repo — see the [Community Edition scope boundary](../ROADMAP.md#community-edition-scope-boundary).

## How a second option gets added to a row

Per [`ROADMAP.md`](../ROADMAP.md) Release 0.3, a second serving runtime is planned next
for CE. Each addition gets validated end to end the same way this first combination was,
before being added as a new row entry — never assumed to work by extrapolation. The same
applies to cluster distributions: GKE/kind/EKS/AKS move from "compatible by design" to
"Validated" in the matrix above only once actually tested, even though they're already
in CE's scope by design.

Anything not listed as validated should be assumed untested, not "probably works."
