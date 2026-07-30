---
name: decision-records
description: Record architectural and product decisions as ADRs under docs/decisions/, including the alternatives that were rejected. Use whenever the user answers a decision question, when a plan's Decisions section changes, or when the user asks about past decisions, why something was built a certain way, or wants to revisit an earlier choice.
---

# Decision Records

Record product and architectural decisions in `docs/decisions/` so future readers can see what was chosen, what was rejected, and why.

## When to act

- After the user answers an `AskQuestion` call (see capture below).
- After the user decides something in prose.
- When a plan's `Decisions` section is updated during `/critique-plan`.
- After plan approval (`create-implementation-plan` Step 5).
- When the user asks why something was built a certain way.

## Capture (do this before writing an ADR)

### AskQuestion answers

The `postToolUse` hook at `.cursor/hooks/log-decision.sh` **does not fire for AskQuestion today** (Cursor bug). After the user answers, call:

```bash
.cursor/scripts/record-decision.sh \
  --id "<question_id>" \
  --question "<question text as presented>" \
  --chosen "<label of chosen option>" \
  --rejected "<label of rejected option 1>" \
  --rejected "<label of rejected option 2>"
```

If the user picked **Other** with free text, use `--chosen "<their exact words>"` and list every offered option as `--rejected`.

### Prose decisions

When the user decides in conversation ("go ahead", "use London time", "approve"), call the same script. Reconstruct only alternatives that were genuinely on the table — do not invent options.

## Significance filter

**Promote to ADR** when the choice changes:

- Database schema or migrations
- API contract (paths, fields, status codes, error semantics)
- Product behaviour visible to users
- Security or privacy posture
- Architectural boundaries between domains

**Ledger only** (no ADR): process choices ("approve or iterate?"), model selection, formatting preferences, anything leaving no lasting artifact.

## Writing an ADR

1. Read the next free number from [`docs/decisions/README.md`](../../docs/decisions/README.md) index table.
2. Copy [`docs/decisions/TEMPLATE.md`](../../docs/decisions/TEMPLATE.md) to `docs/decisions/NNNN-slug.md`.
3. One ADR **per feature**, accumulating that feature's decisions in one document when they belong together.
4. Under **Options considered**, paste the tradeoff text **as it was presented**, not a summary.
5. Set `Reconstructed: yes` when rationale was inferred from plan files or transcripts, not recorded at decision time.
6. **Update the README index table** in the same change.
7. **Append promotion records** to the ledger (never edit existing lines):

```bash
echo '{"type":"promotion","ts":"'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'","question_id":"<id>","adr":"NNNN"}' >> docs/decisions/ledger.jsonl
```

## Rules

- **Never edit a published ADR** to reverse a decision. Write a new ADR with `Status: Superseded by NNNN` on the old one.
- **Number collisions:** if two branches both use `0004`, renumber the later ADR and update references.
- **Privacy:** redact user IDs, coordinates, phone numbers, and message content before promoting. The ledger is committed.
- **Rejected options are the point.** A future reader revisiting a decision needs to see what was on the table and why it lost.

## During `/critique-plan`

After Bucket A answers are in and the plan is patched:

1. Update or create the feature's ADR with each Bucket A decision and rejected options.
2. Call `record-decision.sh` for each answered question (hook will not have fired).
3. Append promotion records for promoted decisions.

## When implementation ships

After the plan is built and committed, close the ADR — do not leave "planned" or "not implemented" on `main`:

1. Follow [`close-adr-on-ship`](../close-adr-on-ship/SKILL.md).
2. Update header (`Implemented`, `Commit`, `Linear`) and add `## Implementation`.
3. Run `.cursor/scripts/mark-adr-implemented.sh` to append a ledger record.

Decision sections stay unchanged; only implementation metadata is added.

## Status lifecycle

`Proposed` → `Accepted` → *(implemented — document in `## Implementation`)* → `Superseded by NNNN`
