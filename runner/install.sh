#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEST="$HOME/.config/systemd/user"
SVC="miramar-ai-platform-runner"

mkdir -p "$DEST"

# Enable linger so the user service starts on boot without an interactive login
loginctl enable-linger "$(id -un)"

# Create the runner PAT env file if it doesn't exist; seed from current session if vars are set
RUNNER_ENV="${DEST}/${SVC}.env"
if [[ ! -f "${RUNNER_ENV}" ]]; then
    printf 'GITHUB_ORG_ADMIN_PAT=\nGITHUB_ORG_GHCR_PAT=\n' >"${RUNNER_ENV}"
    [[ -n "${GITHUB_ORG_ADMIN_PAT:-}" ]] && sed -i "s|^GITHUB_ORG_ADMIN_PAT=|GITHUB_ORG_ADMIN_PAT=${GITHUB_ORG_ADMIN_PAT}|" "${RUNNER_ENV}"
    [[ -n "${GITHUB_ORG_GHCR_PAT:-}" ]] && sed -i "s|^GITHUB_ORG_GHCR_PAT=|GITHUB_ORG_GHCR_PAT=${GITHUB_ORG_GHCR_PAT}|" "${RUNNER_ENV}"
    chmod 600 "${RUNNER_ENV}"
    echo "Created ${RUNNER_ENV} (fill in PATs if not seeded from current environment)"
else
    echo "${RUNNER_ENV} already exists — skipping"
fi

echo "Installing ${SVC}.service..."
cp "${SCRIPT_DIR}/${SVC}.service" "${DEST}/${SVC}.service"

systemctl --user daemon-reload
systemctl --user enable "${SVC}"
systemctl --user restart "${SVC}"
printf '  %-28s %s\n' "${SVC}" "$(systemctl --user is-active "${SVC}")"
