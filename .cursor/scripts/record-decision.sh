#!/usr/bin/env bash
# Append a prose decision to docs/decisions/ledger.jsonl.
# Called by agents when the user decides something in conversation.
#
# Usage:
#   .cursor/scripts/record-decision.sh \
#     --question "Which timezone?" \
#     --chosen "19:00 Europe/London" \
#     --rejected "Per-user timezone" \
#     --rejected "Fixed UTC"
#
# Or with a structured question id:
#   .cursor/scripts/record-decision.sh --id discover --question "..." --chosen "..."
set -euo pipefail

LEDGER="${LEDGER:-docs/decisions/ledger.jsonl}"

question=""
question_id=""
chosen=""
declare -a rejected=()

usage() {
  cat <<'EOF'
Usage: record-decision.sh --question TEXT --chosen TEXT [--id ID] [--rejected TEXT]...
EOF
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --question)
      question="${2:-}"
      shift 2
      ;;
    --id)
      question_id="${2:-}"
      shift 2
      ;;
    --chosen)
      chosen="${2:-}"
      shift 2
      ;;
    --rejected)
      rejected+=("${2:-}")
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      ;;
  esac
done

if [[ -z "$question" || -z "$chosen" ]]; then
  usage
fi

if [[ -z "$question_id" ]]; then
  question_id="$(echo "$question" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '-' | sed 's/^-//;s/-$//' | cut -c1-48)"
fi

mkdir -p "$(dirname "$LEDGER")"
touch "$LEDGER"

branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")"
ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

jq -nc \
  --arg ts "$ts" \
  --arg branch "$branch" \
  --arg question_id "$question_id" \
  --arg question "$question" \
  --arg chosen "$chosen" \
  --argjson rejected "$(printf '%s\n' "${rejected[@]:-}" | jq -R -s 'split("\n") | map(select(length>0))')" \
  '{
    type: "decision",
    ts: $ts,
    source: "prose",
    session: null,
    branch: $branch,
    question_id: $question_id,
    question: $question,
    options: (
      [{id: "chosen", label: $chosen, chosen: true}]
      + ($rejected | map({id: ("rejected-" + (.[0:24])), label: ., chosen: false}))
    ),
    custom_answer: null
  }' >> "$LEDGER"

exit 0
