# Prerequisites (target, v0)

This page states what a user's environment needs for the v0.1 golden path
(`deploy`/`validate`/`undeploy`/`new`), validated against a real DGX Spark cluster —
see [`docs/supported-configurations.md`](supported-configurations.md). `miramar doctor`
checks API connectivity, distribution, schedulable GPU, target namespace, and Helm
release-storage access automatically — run it first. It does not yet check storage
class or secret prerequisites; verify those by hand for now.

## Requirements for v0.1

- An existing k3s cluster with at least one GPU-enabled node — for v0.1, this is the
  author's own DGX Spark environment (see
  [`docs/supported-configurations.md`](supported-configurations.md))
- `kubectl` configured with access to that cluster
- The NVIDIA device plugin already installed on the cluster
- Sufficient GPU memory to run a vLLM-served model. Measured on DGX Spark's GB10
  (unified memory shared with the host): ~122 GiB total, only ~105 GiB free at boot, so
  the template's `vllm.gpuMemoryUtilization` defaults to `0.75` (not vLLM's own default
  of `0.90`, which overcommits and refuses to start). The default model
  (Qwen2.5-1.5B-Instruct) fits comfortably at this setting; a larger model — Mistral-7B-
  Instruct-v0.3 (~13.5 GiB checkpoint) was validated as a customized `--chart-dir`
  deploy — also fits, but on a smaller or more heavily-loaded GPU this value may need
  lowering further via `miramar new` + editing `values.yaml`.
- Outbound network access from the cluster to pull container images and model weights.
  A cold pull of a multi-GB model can take several minutes to over ten minutes
  (Mistral-7B's weight download alone took ~14 minutes on first pull) — see
  [`docs/quickstart.md`](quickstart.md) for the `--wait-timeout` implications.
- All resources are created in a dedicated `miramar-ai-platform` namespace, created by
  `deploy` itself — the namespace is not required to pre-exist

## Explicitly not required for v0.1

- Cluster provisioning is out of scope — this project does not create the Kubernetes
  cluster or GPU node pool for you.
- GKE / other cloud-managed clusters are not required for v0.1 — v0.1 only validated
  self-hosted k3s. They're compatible by design (CE only needs a reachable Kubernetes
  API and a schedulable GPU) but not yet tested; see
  [`docs/supported-configurations.md`](supported-configurations.md) for the
  validated-vs-compatible-by-design matrix.

`miramar doctor`'s first vertical slice (API connectivity, distribution, schedulable
GPU, target namespace, Helm release storage) has been run against the real DGX Spark
cluster referenced above. This page will be revised further once storage-class and
secret checks are added — see `docs/CURRENT_STATE.md`.
