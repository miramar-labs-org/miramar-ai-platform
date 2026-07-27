#!/usr/bin/env bash
# Show status of the miramar-ai-platform-runner user service.

SVC="miramar-ai-platform-runner"

STATE=$(systemctl --user is-active "${SVC}" 2>/dev/null || true)
ENABLED=$(systemctl --user is-enabled "${SVC}" 2>/dev/null || true)
printf '  %-28s active=%-12s enabled=%s\n' "${SVC}" "${STATE}" "${ENABLED}"

echo ""
systemctl --user status "${SVC}" --no-pager --lines=0 2>&1 || true
