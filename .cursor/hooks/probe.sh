#!/usr/bin/env bash
# Throwaway probe: dumps postToolUse stdin to /tmp for schema inspection.
# Usage: register on postToolUse with no matcher, trigger AskQuestion, read
# /tmp/askquestion-hook-probe.json
set -euo pipefail

PROBE_FILE="${PROBE_FILE:-/tmp/askquestion-hook-probe.json}"
cat > "$PROBE_FILE"
exit 0
