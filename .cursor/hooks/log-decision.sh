#!/usr/bin/env bash
# Append AskQuestion decisions to docs/decisions/ledger.jsonl.
# Fails open on every path — a logging hook must never block agent work.
#
# NOTE: As of Cursor 3.x, AskQuestion does NOT fire postToolUse hooks
# (confirmed bug: forum.cursor.com/t/askquestion-tool-does-not-trigger-cursor-hooks).
# This script is ready for when that is fixed. Until then, agents must call
# .cursor/scripts/record-decision.sh after the user answers a question.
set -euo pipefail

LEDGER="${LEDGER:-docs/decisions/ledger.jsonl}"

input="$(cat 2>/dev/null || true)"
if [[ -z "$input" ]]; then
  exit 0
fi

tool_name="$(echo "$input" | jq -r '.tool_name // empty' 2>/dev/null || true)"
case "$tool_name" in
  AskQuestion|AskUserQuestion) ;;
  *)
    exit 0
    ;;
esac

mkdir -p "$(dirname "$LEDGER")"
touch "$LEDGER"

branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")"
ts="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

echo "$input" | jq -c --arg ts "$ts" --arg branch "$branch" '
  (.tool_output // "") as $raw_out |
  (if ($raw_out | type) == "string" and ($raw_out | length) > 0 then
     ($raw_out | fromjson? // {})
   elif ($raw_out | type) == "object" then
     $raw_out
   else
     {}
   end) as $out |
  (.conversation_id // .session_id // .session // null) as $session |
  (.tool_input.questions // [])[] |
  . as $q |
  ($out.answers // $out // {}) as $answers |
  ($answers[$q.id] // $answers[($q.id | tostring)] // null) as $ans |
  (if ($ans | type) == "object" then ($ans.custom // $ans.text // $ans.value // null) else $ans end) as $raw |
  (if $raw != null and ([$q.options[]?.id] | any(. == $raw or (.|tostring) == ($raw|tostring))) then null else $raw end) as $custom |
  (if $custom != null then
     ([$q.options[]? | . + {chosen: false}])
   else
     ([$q.options[]? | . + {chosen: ((.id == $raw) or (.id == ($raw | tostring)))}])
   end) as $opts |
  {
    type: "decision",
    ts: $ts,
    source: "askquestion",
    session: $session,
    branch: $branch,
    question_id: ($q.id // "unknown"),
    question: ($q.prompt // $q.question // ""),
    options: $opts,
    custom_answer: $custom
  }
' >> "$LEDGER" 2>/dev/null || true

exit 0
