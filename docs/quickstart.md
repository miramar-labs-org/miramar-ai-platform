# Quickstart

This is the real, tested v0.1 golden path, validated against a live cluster
(self-hosted k3s on a DGX Spark — see
[`docs/supported-configurations.md`](supported-configurations.md)).
Run `miramar doctor` first — it read-only checks
that a working GPU-enabled cluster and `kubectl` access are already in place, and reports
what it found about storage, ingress, and observability integrations on the cluster; see
[`docs/prerequisites.md`](prerequisites.md) for what it expects.

All commands default to namespace `miramar-ai-platform` and release name `miramar`;
override with the persistent `-n`/`--namespace` and `--release-name` flags if needed.

## Preflight

```
go run ./cmd/miramar doctor
```

Read-only — confirms the cluster is reachable, a GPU is schedulable, and reports what
else it found (storage classes, ingress controllers, observability integrations) before
`deploy` touches anything.

## Zero-flag path

Deploys the built-in default model (`Qwen/Qwen2.5-1.5B-Instruct`) with no configuration:

```
go run ./cmd/miramar deploy
```

This creates the `miramar-ai-platform` namespace (if it doesn't already exist) and
installs the `serving-vllm` template's Helm chart into it. First run pulls the vLLM
image and downloads model weights from Hugging Face, so it can take several minutes —
the CLI's own `--wait-timeout` (default 15m) may need raising for larger models (see
below); the underlying pod keeps starting even if the CLI's wait times out, because a
Kubernetes `startupProbe` — not a fixed delay — gates readiness until the endpoint is
actually up.

Confirm it's healthy and serving:

```
go run ./cmd/miramar validate
```

This polls for a ready pod, checks `GET /v1/models`, and sends a small built-in set of
prompts to `POST /v1/chat/completions`, reporting pass/fail per prompt.

Tear it down:

```
go run ./cmd/miramar undeploy
```

This removes the Helm release's resources but leaves the `miramar-ai-platform`
namespace in place. To remove the namespace too:

```
go run ./cmd/miramar undeploy --purge-namespace
```

## Customize-then-deploy path

To deploy a different model, use `miramar new` to copy the template out to a local
directory, edit it, then deploy from that copy instead of the embedded default:

```
go run ./cmd/miramar new --model mistralai/Mistral-7B-Instruct-v0.3 --dir /tmp/serving-test
# optionally edit /tmp/serving-test/values.yaml further by hand — e.g. resources,
# vllm.gpuMemoryUtilization, probe timing
go run ./cmd/miramar deploy --chart-dir /tmp/serving-test --wait-timeout 20m
go run ./cmd/miramar validate
go run ./cmd/miramar undeploy --purge-namespace
```

`--wait-timeout` was raised here because a larger, not-yet-cached model can take longer
to download than the 15-minute default — this is a client-side wait on top of the
`startupProbe`, not a correctness issue: if it's exceeded, the deploy command exits
non-zero but the pod keeps starting underneath, and a follow-up `deploy` (or `kubectl
get pod`) against the same release picks up the already-healthy result once it's ready.

## Notes from real runs

- Qwen2.5-1.5B-Instruct (zero-flag default) reaches a healthy pod within the default
  15-minute wait on a warm image cache.
- Mistral-7B-Instruct-v0.3 (customized `--chart-dir` example above) took over 13 minutes
  for weight download alone on first pull — hence the longer `--wait-timeout` above.
- GPU memory is confirmed fully released after `undeploy` (no compute processes left on
  the GPU) in both cases.
