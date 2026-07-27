# AGENTS.md

This file provides repository-level instructions for coding agents working on Miramar AI Platform.

## Read before making changes

Review these files in this order:

1. `README.md` — current product scope and user-facing golden path
2. `ROADMAP.md` — release scope, sequencing, and adoption milestones
3. `PRINCIPLES.md` — durable project-wide decision principles
4. `docs/architecture.md` — current v0 architecture
5. `docs/adr/README.md` and relevant ADRs — accepted design decisions
6. `docs/supported-configurations.md` — validated versus compatible environments
7. `docs/CURRENT_STATE.md` — current implementation state and immediate work

When these documents disagree:

* An accepted ADR is authoritative for the specific decision it records.
* `PRINCIPLES.md` is authoritative for durable project philosophy.
* `ROADMAP.md` is authoritative for planned release scope.
* The implementation and tests are authoritative for what currently exists.
* Flag contradictions rather than silently choosing one interpretation.

## Current product boundary

Miramar Community Edition deploys and validates AI workloads on an existing, reachable Kubernetes cluster.

Community Edition may inspect cluster prerequisites but must not create, destroy, resize, or otherwise own the cluster lifecycle.

Cluster provisioning and organizational-scale platform operations belong to the separately planned Enterprise Edition. Do not add Enterprise Edition code to this repository unless explicitly instructed and supported by a new accepted ADR.

## Implementation approach

* Prefer small, reviewable vertical slices.
* Preserve the existing Go and Cobra command structure.
* Use existing Kubernetes and Helm primitives rather than reimplementing them.
* Keep operations transparent and independently verifiable.
* Prefer explicit errors with actionable remediation.
* Preserve the zero-flag golden path.
* Avoid abstractions that are only justified by speculative future features.
* Do not broaden supported claims beyond configurations actually validated.
* Distinguish “validated” from “compatible by design.”

## Scope discipline

Do not implement later roadmap features merely because they are described in `ROADMAP.md`.

Before starting work:

1. Identify the release and roadmap item the change belongs to.
2. Confirm that the requested behavior does not conflict with an accepted ADR.
3. State any assumptions that materially affect the implementation.

Do not:

* Copy implementation code from the author’s separate reference-platform repository.
* Introduce cloud-specific behavior into the CE deployment path unless it is optional and isolated.
* Add cluster provisioning to CE.
* Introduce license enforcement or pricing code.
* Claim support for an unvalidated environment.
* Replace accepted architectural decisions without creating a superseding ADR.
* perform broad refactors unrelated to the requested task.

## Tests and validation

Changes should normally include:

* Unit tests for new logic
* Failure-path tests where practical
* Updated command help or documentation when CLI behavior changes
* Updates to `docs/supported-configurations.md` only after real validation
* Updates to `docs/CURRENT_STATE.md` when implementation status materially changes

Before considering work complete, run the repository’s documented formatting, test, build, and lint commands.

Do not mark an environment or workflow as validated based only on compilation or mocked tests.

## Documentation rules

* Put durable philosophy in `PRINCIPLES.md`.
* Put release sequencing and planned scope in `ROADMAP.md`.
* Put current system shape in `docs/architecture.md`.
* Put individual architectural decisions in `docs/adr/`.
* Put transient implementation status and the immediate next task in `docs/CURRENT_STATE.md`.
* Avoid repeating the same detailed explanation across multiple files; link to the authoritative source instead.
* Use colored status icons for pass/fail/done/pending-shaped content in any markdown
  file: ✅ done/accepted/validated/implemented, 🚧 stub/in-progress, ⏳ planned/deferred
  (roadmap items), 🕒 deferred (ADR status) or compatible-by-design-not-yet-validated,
  ⬜ not yet considered. Don't pair an icon with a GitHub checkbox (`- [x]`/`- [ ]`) on
  the same line — the checkbox already encodes done/pending, so use a plain `- ✅`/`- ⏳`
  bullet instead of `- [x] ✅`/`- [ ] ⏳`.

## Working style

When asked to implement a feature:

1. Inspect the relevant code and authoritative documentation.
2. Summarize the intended change and files likely to be affected.
3. Implement the smallest complete slice.
4. Run appropriate tests and checks.
5. Report:

   * what changed
   * what was validated
   * what remains unvalidated
   * any documentation or ADR implications

If a request conflicts with repository decisions, stop and explain the conflict before modifying code.
