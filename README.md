# Miramar AI Platform

[![Build CLI](https://github.com/miramar-labs-org/miramar-ai-platform/actions/workflows/build.yaml/badge.svg)](https://github.com/miramar-labs-org/miramar-ai-platform/actions/workflows/build.yaml)
[![Build Runner Image](https://github.com/miramar-labs-org/miramar-ai-platform/actions/workflows/build-runner-image.yaml/badge.svg)](https://github.com/miramar-labs-org/miramar-ai-platform/actions/workflows/build-runner-image.yaml)
[![Go Version](https://img.shields.io/badge/go-1.25-00ADD8?logo=go)](go.mod)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**Deploy production-grade AI inference to Kubernetes with one opinionated CLI.**

A Kubernetes-native AI deployment platform — not a general-purpose AI platform. Miramar
targets exactly one thing well: getting a model endpoint running, validated, and torn
down on a GPU-enabled Kubernetes cluster you already have, without you assembling GPU
scheduling, a serving runtime, and health/validation tooling by hand.

## Who this is for

AI platform teams, applied AI engineers, and consultancies that need a repeatable,
documented path to running a model on their own Kubernetes infrastructure — without
spending months learning and wiring together Kubernetes, GPU scheduling, serving
runtimes, and health/validation tooling by hand.

## What problem this solves

Standing up private AI infrastructure usually means assembling Kubernetes, GPU device
plugins, a serving runtime, health checks, and rollback/cleanup procedures from scratch —
often taking a team months before the first model is actually serving traffic reliably.
Miramar AI Platform aims to package that path into a small set of commands, starting with
a single well-tested workflow rather than trying to cover every possible configuration on
day one.

## What exists today

The `miramar` CLI (Go + Cobra) deploys, validates, and tears down a vLLM model endpoint
on a GPU-enabled Kubernetes cluster, plus a template-factory command for customizing the
deployment before applying it:

- ✅ `miramar doctor` — read-only preflight check of cluster prerequisites (API
  connectivity, distribution, schedulable GPU, storage classes, ingress controller,
  observability integrations, target namespace, Helm release storage). Run this first —
  it's how Miramar shows it understands the platform it's deploying to, before it
  touches anything.
- ✅ `miramar new` — copy a deployment template to a local directory for customization
- ✅ `miramar deploy` — deploy one model endpoint via a Helm chart (embedded default, or a
  customized copy via `--chart-dir`)
- ✅ `miramar validate` — confirm the endpoint is healthy and serving requests
- ✅ `miramar undeploy` — tear the deployment back down, optionally purging the namespace

## Golden path

The first concrete workflow this project targets:

> Deploy and validate an OpenAI-compatible model endpoint on an existing GPU-enabled
> Kubernetes cluster.

Zero-flag path, using the built-in default (Qwen2.5-1.5B-Instruct on self-hosted k3s):

```
miramar doctor         # confirm the cluster is ready before touching anything
miramar deploy         # deploy the default model endpoint
miramar validate       # confirm the endpoint is healthy and serving
miramar undeploy       # tear it back down
```

Customize-then-deploy path, for a different model:

```
miramar doctor
miramar new --model mistralai/Mistral-7B-Instruct-v0.3 --dir ./my-model
# edit ./my-model/values.yaml further by hand if needed
miramar deploy --chart-dir ./my-model
miramar validate
miramar undeploy --purge-namespace
```

Both paths are validated against a live cluster — see
[`docs/supported-configurations.md`](docs/supported-configurations.md). See
[`ROADMAP.md`](ROADMAP.md) for what's still ahead,
[`docs/architecture.md`](docs/architecture.md) for the shape of the system, and
[`PRINCIPLES.md`](PRINCIPLES.md) for the philosophy behind the design choices.

## What this is not (yet)

To keep this honest about scope:

- Not a managed SaaS or hosted control plane
- Not a Kubernetes operator
- Not multi-cluster — `deploy` targets one cluster at a time, whichever you already
  have `kubectl` access to (self-hosted or cloud-managed); only one cluster/runtime
  combination has actually been validated end to end so far, see
  [`docs/supported-configurations.md`](docs/supported-configurations.md)
- Not covered by any compliance certification or commercial SLA
- Not accepting outside code contributions yet — see [`CONTRIBUTING.md`](CONTRIBUTING.md)

## License

Apache License 2.0 — see [`LICENSE`](LICENSE).
