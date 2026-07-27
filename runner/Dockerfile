FROM debian:bookworm-slim

ARG RUNNER_VERSION=2.335.1
ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive \
    RUNNER_USER=runner \
    RUNNER_HOME=/home/runner

# Bare minimum to run the GitHub Actions runner agent and build the Go CLI.
# Go itself is NOT baked in — actions/setup-go installs the pinned version
# per workflow run. Add tooling here only when a workflow actually needs it.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        jq \
        libssl3 \
        libkrb5-3 \
        zlib1g \
        libicu72 \
    && rm -rf /var/lib/apt/lists/*

LABEL org.opencontainers.image.version="${RUNNER_VERSION}"

RUN useradd -m -s /bin/bash runner

WORKDIR /home/runner

RUN case "${TARGETARCH}" in \
        amd64) RUNNER_ARCH="x64" ;; \
        arm64) RUNNER_ARCH="arm64" ;; \
        *) echo "Unsupported arch: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    curl -fsSL \
        "https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz" \
        -o runner.tar.gz \
    && tar xzf runner.tar.gz \
    && rm runner.tar.gz

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN chown -R runner:runner /home/runner

USER runner

ENTRYPOINT ["/entrypoint.sh"]
