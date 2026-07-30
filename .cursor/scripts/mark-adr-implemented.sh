#!/usr/bin/env bash
# Mark a feature ADR as implemented after the plan is built and committed.
# Appends a ledger record and prints the metadata block to paste into the ADR.
#
# Usage:
#   .cursor/scripts/mark-adr-implemented.sh \
#     --adr docs/decisions/0003-daily-two-most-compatible.md \
#     --commit 90a484a \
#     --linear "https://linear.app/haerd/issue/HAE-451/..." \
#     [--date 2026-07-30]
#
# Fails open: missing args print usage and exit 0 so agents can fall back to manual edit.

set -euo pipefail

ROOT="$(git -C "$(dirname "$0")/../.." rev-parse --show-toplevel 2>/dev/null || pwd)"
LEDGER="$ROOT/docs/decisions/ledger.jsonl"
ADR=""
COMMIT=""
LINEAR="—"
DATE="$(date -u +%Y-%m-%d)"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --adr) ADR="$2"; shift 2 ;;
    --commit) COMMIT="$2"; shift 2 ;;
    --linear) LINEAR="$2"; shift 2 ;;
    --date) DATE="$2"; shift 2 ;;
    -h|--help)
      sed -n '2,12p' "$0"
      exit 0
      ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$ADR" || -z "$COMMIT" ]]; then
  echo "Usage: mark-adr-implemented.sh --adr <path> --commit <sha> [--linear URL] [--date YYYY-MM-DD]" >&2
  exit 1
fi

if [[ ! -f "$ROOT/$ADR" && ! -f "$ADR" ]]; then
  echo "ADR not found: $ADR" >&2
  exit 1
fi

ADR_PATH="$ADR"
if [[ ! -f "$ADR_PATH" ]]; then
  ADR_PATH="$ROOT/$ADR"
fi

ADR_BASENAME="$(basename "$ADR_PATH" .md)"
ADR_NUM="${ADR_BASENAME%%-*}"

mkdir -p "$(dirname "$LEDGER")"
touch "$LEDGER"

TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
RECORD=$(jq -nc \
  --arg ts "$TS" \
  --arg adr "$ADR_BASENAME" \
  --arg commit "$COMMIT" \
  --arg linear "$LINEAR" \
  --arg date "$DATE" \
  '{type:"implementation",ts:$ts,adr:$adr,commit:$commit,linear:$linear,implemented_date:$date}')

echo "$RECORD" >> "$LEDGER"

echo "Ledger appended: implementation record for $ADR_BASENAME"
echo ""
echo "--- Paste or reconcile in $ADR_PATH header ---"
echo "- **Implemented:** $DATE"
echo "- **Commit:** \`$COMMIT\`"
if [[ "$LINEAR" != "—" ]]; then
  echo "- **Linear:** $LINEAR"
fi
echo ""
echo "--- Add ## Implementation section if missing (see close-adr-on-ship skill) ---"
