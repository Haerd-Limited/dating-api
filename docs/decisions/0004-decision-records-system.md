# 0004 — Decision Records System

- **Status:** Accepted
- **Date:** 2026-07-30
- **Feature:** developer tooling / documentation
- **Plan:** `decision_records_system_3cb9a9d6.plan.md` (lives outside the repo in `~/.cursor/plans/`)
- **Linear:** —
- **Reconstructed:** no

## Context

Decisions and rejected alternatives lived only in `~/.cursor/plans/` — outside version control, mixed across projects, lost when plans are abandoned. The team wanted every decision and option tracked so future readers can understand why the app is shaped as it is and revisit choices.

## Decision

Two-tier system under `docs/decisions/`:

1. **Ledger** (`ledger.jsonl`) — append-only raw capture of every decision.
2. **ADRs** (`NNNN-slug.md`) — curated per-feature write-ups with verbatim rejected options.

Hook at `.cursor/hooks/log-decision.sh` on `postToolUse` (ready for AskQuestion when Cursor fixes the hook bug). Prose capture via `.cursor/scripts/record-decision.sh` is the **primary path** until then. Skill, `AGENTS.md` section, and command updates wire capture into planning workflows. Backfill ADRs 0001–0003 from recent features.

## Options considered

### Two tiers: ledger + ADRs — CHOSEN

Ledger captures everything verbatim; ADRs promote substantive decisions into readable history. Avoids hundreds of one-file-per-decision ADRs while keeping nothing lost.

- **Rejected:** Classic one ADR per decision — directory grows fast; process meta-decisions pollute it.
- **Rejected:** One ADR per feature only, no ledger — loses the exhaustive audit trail.

### Hook plus skill — CHOSEN

Hook for deterministic capture; skill for significance judgement and ADR prose.

- **Rejected:** Hook only — raw JSON without context is auditable but unreadable.
- **Rejected:** Skill only — depends on agent compliance; decisions slip through.

### Backfill three recent features — CHOSEN

0001 two-match cap, 0002 question packs, 0003 daily picks — gives immediate value.

- **Rejected:** Daily picks only — history starts mid-story.
- **Rejected:** Start fresh — leaves the "how did we get here" gap open.

### `docs/decisions/` committed — CHOSEN

Both ledger and ADRs in git under `docs/`.

- **Rejected:** Gitignore ledger — exhaustive record lost on fresh clone.
- **Rejected:** `.cursor/decisions/` — buried in agent tooling, invisible to non-engineers.

### Sequential `0001-slug.md` numbering — CHOSEN (critique round)

Canonical, short citable IDs for supersede references.

- **Rejected:** Date-prefixed filenames — collision-proof but no short ID.
- **Rejected:** Hybrid `0004-2026-07-30-slug` — long filenames without preventing number collisions.

### Prose decisions captured too — CHOSEN (critique round)

`record-decision.sh` + `AGENTS.md` instruction. Primary path until AskQuestion hooks work.

- **Rejected:** AskQuestion only — misses conversational decisions.
- **Rejected:** End-of-session sweep — lost if session ends abruptly.

### Hook writes immediately in plan mode — CHOSEN (critique round)

Nothing lost, including abandoned plans. Accepted cost: ledger diffs during planning.

- **Rejected:** Buffer during plan mode — abandoned plan decisions lost permanently.
- **Rejected:** Gitignored staging file — reverses committed-everything choice.

### Calls made without asking (plan author)

- **JSONL ledger, not markdown append** — bash + `jq` is escape-safe; markdown breaks on newlines in option labels.
- **`AGENTS.md` pointer, not new `.cursor/rules/`** — already always-applied; avoids parallel mechanism.

## Consequences

- `.gitattributes` `merge=union` on ledger to prevent merge conflicts.
- Promotion appends `type: promotion` records — never edits existing ledger lines.
- `custom_answer` field for AskQuestion "Other" free-text path.
- AskQuestion hook bug documented in README — agents must call `record-decision.sh` after every answered question until Cursor fixes it.
- Critique-plan and create-implementation-plan commands updated with ADR steps.

## Supersedes / Superseded by

- Supersedes: informal plan-file-only decision tracking
- Superseded by: —
