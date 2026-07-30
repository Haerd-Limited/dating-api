# Decision Records

Permanent record of product and architectural decisions for the Haerd dating API, including the alternatives that were rejected at the time.

## Two tiers

| Tier | Location | Purpose |
|------|----------|---------|
| **Ledger** | [`ledger.jsonl`](ledger.jsonl) | Append-only raw capture of every decision question and answer |
| **ADRs** | `NNNN-slug.md` | Curated, human-readable write-ups for substantive decisions |

The ledger is the audit trail. ADRs are what you read when revisiting a choice.

## ADR index

| # | Title | Status | Date | Reconstructed |
|---|-------|--------|------|---------------|
| [0001](0001-two-match-cap.md) | Two active match cap | Accepted | 2026-04-30 | yes |
| [0002](0002-question-packs-onboarding-step.md) | Question packs as onboarding step | Accepted | 2026-07-30 | yes |
| [0003](0003-daily-two-most-compatible.md) | Daily two most compatible (implemented) | Accepted | 2026-07-30 | partial |
| [0004](0004-decision-records-system.md) | Decision records system | Accepted | 2026-07-30 | no |

**Next free number:** 0005

When adding an ADR, update this table in the same commit.

## How decisions are captured

### AskQuestion (automatic — when Cursor supports it)

A `postToolUse` hook at [`.cursor/hooks/log-decision.sh`](../.cursor/hooks/log-decision.sh) is registered in [`.cursor/hooks.json`](../.cursor/hooks.json).

**Known limitation:** As of Cursor 3.x, the `AskQuestion` tool does **not** fire `postToolUse` hooks ([confirmed bug](https://forum.cursor.com/t/askquestion-tool-does-not-trigger-cursor-hooks/152230)). The hook script is ready for when this is fixed. Until then, agents must use prose capture (below) after every `AskQuestion` answer.

To verify whether hooks are active: Cursor **Settings → Hooks** tab, or the **Hooks** output channel. Cloning this repo gives you the hook files, but does not guarantee they are running — restart Cursor after changing `hooks.json`.

To probe the hook input schema manually, use [`.cursor/hooks/probe.sh`](../.cursor/hooks/probe.sh) and inspect `/tmp/askquestion-hook-probe.json`.

### Prose decisions (agent compliance required)

When the user decides something in conversation ("go ahead", "use London time", "approve"), the agent calls:

```bash
.cursor/scripts/record-decision.sh \
  --question "Which timezone should the daily drop use?" \
  --chosen "Single global 19:00 Europe/London" \
  --rejected "Per-user timezone column" \
  --rejected "Fixed UTC"
```

This is the **primary capture path** until the AskQuestion hook bug is fixed.

### Promotion to ADR

Substantive decisions (schema, API contract, product behaviour, security, architecture) get promoted into an ADR using the [`decision-records` skill](../.cursor/skills/decision-records/SKILL.md). Promotion **appends** a record to the ledger — it never edits existing lines:

```json
{"type": "promotion", "ts": "2026-07-30T16:00:00Z", "question_id": "structure", "adr": "0004"}
```

When a feature is **built and committed**, mark its ADR implemented using the [`close-adr-on-ship` skill](../.cursor/skills/close-adr-on-ship/SKILL.md) and `.cursor/scripts/mark-adr-implemented.sh`. Append-only ledger record:

```json
{"type": "implementation", "ts": "2026-07-30T16:00:00Z", "adr": "0003-daily-two-most-compatible", "commit": "90a484a", "linear": "https://...", "implemented_date": "2026-07-30"}
```

## Reading the ledger

List unpromoted decisions (anti-join on promotion records):

```bash
jq -rs '
  (map(select(.type=="promotion") | .question_id) | unique) as $done
  | map(select(.type=="decision" and (.question_id | IN($done[]) | not)))
  | .[] | "\(.ts)  \(.source)  \(.question_id)  \(.question[0:70])"
' docs/decisions/ledger.jsonl
```

Count records (validates JSONL):

```bash
jq -s 'length' docs/decisions/ledger.jsonl
```

Show recent decisions:

```bash
tail -5 docs/decisions/ledger.jsonl | jq .
```

## Ledger schema

Each `type: "decision"` record:

```json
{
  "type": "decision",
  "ts": "2026-07-30T15:36:02Z",
  "source": "askquestion | prose",
  "session": null,
  "branch": "main",
  "question_id": "structure",
  "question": "...",
  "options": [
    {"id": "twotier", "label": "...", "chosen": true},
    {"id": "per-decision", "label": "...", "chosen": false}
  ],
  "custom_answer": null
}
```

- **`custom_answer`:** populated when the user picks "Other" with free text. In that case every option may be `chosen: false` — that is normal.
- **`source`:** `askquestion` = hook-written (deterministic); `prose` = script-written (agent judgement).

## Privacy

The ledger is committed to git. It must **never** contain PII: user IDs, coordinates, phone numbers, message content, or profile data. Follow `AGENTS.md` rule 10 (GDPR logging hygiene). Redact before promoting to an ADR.

## Conventions

- **Numbering:** sequential `0001-slug.md`. If two branches land the same number, renumber the later ADR and update references.
- **One ADR per feature**, accumulating that feature's decisions.
- **Never edit a published ADR** to reverse a decision. Write a new ADR that supersedes it.
- **`Reconstructed: yes`** on backfilled ADRs — rationale was inferred from plan files, not recorded at decision time.
- **Merge safety:** `ledger.jsonl` uses `merge=union` in [`.gitattributes`](../.gitattributes) so concurrent appends do not conflict.

## Template

Copy [`TEMPLATE.md`](TEMPLATE.md) when writing a new ADR.
