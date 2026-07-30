---
name: close-adr-on-ship
description: Mark a feature ADR as implemented when a planned feature is built and committed. Use after implementation is complete, after git commit/push of the feature, or when the user asks to close or update an ADR for shipped work.
---

# Close ADR on Ship

When a feature plan has been **built and committed**, update its ADR so `docs/decisions/` reflects reality. Decision text stays frozen; only **implementation metadata** and an **Implementation** section are added or updated.

## When to act

- Immediately after committing (or pushing) implementation work for a feature that has an ADR.
- When `/deviation-analysis` or the user notes the ADR still says "planned" or "not implemented".
- Before telling the user a plan is "done" — ADR closure is part of done.

## Find the ADR

1. Check the plan's `Decisions` section or linked ADR path.
2. Search `docs/decisions/README.md` index by feature name.
3. Grep `docs/decisions/` for the domain or plan slug.

If no ADR exists but the feature had substantive decisions, write one first using [`decision-records`](../decision-records/SKILL.md), then close it.

## What to update (same commit as implementation, or follow-up right after)

### Header metadata

In the ADR front matter (top bullet list), set or add:

```markdown
- **Feature:** <domain> (implemented)
- **Implemented:** YYYY-MM-DD
- **Commit:** `<short-or-full-sha>`
- **Linear:** <issue URL or —>
```

Remove stale phrases like `(planned — not yet implemented)`.

Do **not** change `## Decision`, `Options considered`, or rejected alternatives — those record intent at decision time.

### `## Implementation` section

Add after `## Consequences` (or before `Supersedes` if Consequences is long):

```markdown
## Implementation

Shipped in commit `<sha>` on `<branch>` (typically `main`).

| Area | Location |
|------|----------|
| Migration | `migrations/...` |
| Domain | `internal/{domain}/` |
| API | `GET /api/v1/...` |
| Config / flag | `ENABLE_...` |
| FE handoff | Linear XXX |

**Enable / deploy notes:** (migrations, env vars, flag defaults, anything ops needs before users see it)
```

Keep it factual — file paths, endpoints, flags, Linear ticket for frontend. No PII.

### README index

Update `docs/decisions/README.md` if you add a new ADR. Optional: add `(implemented)` in the Title column for shipped features.

### Ledger

Append an implementation record (do not edit existing lines):

```bash
.cursor/scripts/mark-adr-implemented.sh \
  --adr docs/decisions/NNNN-slug.md \
  --commit <sha> \
  --linear "<url>"
```

Or append manually:

```json
{"type":"implementation","ts":"2026-07-30T16:00:00Z","adr":"0003-daily-two-most-compatible","commit":"90a484a","linear":"https://...","implemented_date":"2026-07-30"}
```

## Checklist before marking done

- [ ] ADR header shows **Implemented** + **Commit**
- [ ] `## Implementation` section lists key paths and enablement notes
- [ ] Stale "planned / not implemented" wording removed
- [ ] Ledger append via script or manual JSONL line
- [ ] `docs/decisions/README.md` index accurate

## Rules

- **Never rewrite decisions** in a shipped ADR to match code drift. If implementation diverged materially, note drift in `## Implementation` and open a new ADR if the product decision changed.
- **Same PR or immediate follow-up** — do not leave ADRs saying "planned" after `main` has the code.
- **Privacy:** no user IDs, tokens, or `.env` values in ADR or ledger.

## Integration with other workflows

| Workflow | ADR action |
|----------|------------|
| `create-implementation-plan` Step 5 | Create/update ADR with decisions (**Proposed** or **Accepted**, not yet implemented) |
| Plan build + commit | **This skill** — mark implemented |
| `/deviation-analysis` finds ADR stale | **This skill** |
| `/critique-plan` | Update decisions only; do not mark implemented until code ships |
