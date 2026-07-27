# ADR-016: Observability boundary: standard signals in CE, managed operations in EE

**Status:** ✅ Accepted
**Date:** 2026-07-26

## Context

Miramar needs an observability story for deployed model endpoints without turning
Community Edition into a monitoring platform. Users may already run Prometheus,
Grafana, OpenTelemetry collectors, hosted tracing products, experiment trackers, or
other observability systems. Treating each product integration as a first-class
architecture concern would push CE toward a kitchen-sink integration matrix and weaken
the project's Kubernetes-native, lightweight posture.

The CE/EE boundary already follows operational scope rather than workload type (see
[ADR-006](006-ce-ee-operational-scope-boundary.md)). Applying that principle to
observability requires separating signal exposure from operational ownership.

## Decision

Community Edition exposes and documents Kubernetes-native and open-standard
observability surfaces for systems the user already operates. This includes health and
status surfaced through Kubernetes APIs, Prometheus-compatible metrics where available,
optional scrape annotations or ServiceMonitor/PodMonitor templates when they do not add
a hard runtime dependency, and future OpenTelemetry hooks for trace/log correlation.

Community Edition may detect existing observability resources and report them as
informational or warning-level context, but missing observability integrations must not
block the self-hosted golden path. CE does not install, upgrade, retain, govern, or
operate Prometheus, Grafana, OpenTelemetry collectors, hosted tracing systems, W&B,
LangSmith, MLflow, or equivalent tools.

Enterprise Edition owns managed observability operations where operational ownership
creates value: deploying and upgrading observability components, retention, alert/SLO
packs, RBAC, team workflows, multi-cluster/fleet dashboards, and supported integrations
with hosted or enterprise observability products. Vendor-specific integrations can be
adapters at this layer, not core CE architecture.

## Consequences

CE remains lightweight and standards-oriented while still fitting into real Kubernetes
platform environments. Users can connect Miramar-managed workloads to their existing
monitoring stack without Miramar becoming responsible for that stack's lifecycle.

Prometheus, Grafana, OpenTelemetry, W&B, LangSmith, MLflow, and similar tools should not
be listed as equivalent architectural primitives. Prometheus-compatible metrics and
OpenTelemetry are standards-oriented integration surfaces; hosted or vendor-specific
products are optional adapters or EE-managed integrations.

`miramar doctor` should not fail because observability is absent. Future observability
checks should be discovery-only unless a specific template or user-selected profile has
made an observability dependency explicit.
