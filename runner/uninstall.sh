#!/usr/bin/env bash
set -euo pipefail

DEST="$HOME/.config/systemd/user"
SVC="miramar-ai-platform-runner"

echo "Removing ${SVC}..."
systemctl --user stop "${SVC}" 2>/dev/null || true
systemctl --user disable "${SVC}" 2>/dev/null || true
rm -f "${DEST}/${SVC}.service"

systemctl --user daemon-reload
echo "Done"
