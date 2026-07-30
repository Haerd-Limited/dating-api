# 0003 — Daily Two Most Compatible

- **Status:** Accepted
- **Date:** 2026-07-30
- **Feature:** dailypicks (planned — not yet implemented)
- **Plan:** `daily-two-most-compatible_658fe2bb.plan.md` (lives outside the repo in `~/.cursor/plans/`)
- **Linear:** —
- **Reconstructed:** partial (early design decisions yes; critique-round decisions no)

## Context

Replace swiping as the primary discovery loop with a daily curated drop of two most compatible profiles. Users decide via existing `POST /api/v1/swipes/`. Matches still come only from the Likes section. Reveal is gated on zero active matches. Nightly job at 19:00 Europe/London precomputes picks.

## Decision

New `dailypicks` domain with batch state machine (`pending` → `revealed` → `completed`), single SQL scorer per viewer (not per-pair `ComputeCompatibility`), reuse swipes for decisions, `ENABLE_DAILY_PICKS` flag gating discover routes and scheduler. See plan for full schema and service design.

### Design decisions (from initial plan — Reconstructed: yes)

| Topic | Choice |
|-------|--------|
| Schedule | Single global 19:00 `Europe/London` |
| Hard filter | `seek_gender_ids` only; age/distance/religion/sexuality/ethnicity are soft |
| Queue depth | One pending batch per user, recomputed nightly |
| Reveal gate | `activeMatches == 0` |
| Target at 2/2 cap | Still eligible as picks (like waits in their Likes) |
| Decision endpoint | Reuse `POST /api/v1/swipes/` |
| Scoring | One SQL query per user; overlap ≥ 5 required |

### Critique-round decisions (Reconstructed: no)

#### Discover coexistence while flag is on

**Gate discover routes off** when `ENABLE_DAILY_PICKS` is true. Code stays; only router registration changes.

- **Rejected:** Keep discover live — old FE builds could swipe the whole userbase, defeating scarcity.
- **Rejected:** Return empty discover feed — softer but confusing empty state.

#### Ranking precedence

**Preferences outrank compatibility:** `ORDER BY prefs_satisfied DESC, compatibility_percent DESC`.

- **Rejected:** Compatibility first — best matches might ignore stated filters.
- **Rejected:** Bucketed tiers — closest to original description but most SQL complexity.

#### Mandatory mismatches and thin-overlap candidates

**Exclude both outright** from the scorer (not ranked last).

- **Rejected:** Rank mandatory mismatches last with `compatibility_percent = 1` — could serve declared dealbreakers.
- **Rejected:** Exclude mismatches but keep thin-overlap as filler — unmeasured compatibility still surfaces.

*Reversal note:* the initial plan had "mandatory mismatches rank last, not excluded". Critique changed this to hard exclusion.

#### Cold start

**On-demand generation** on first `GET /api/v1/daily-picks` when user has never had a batch.

- **Rejected:** Wait for next 19:00 drop — empty screen after long onboarding.
- **Rejected:** Generate at onboarding completion — adds scorer to onboarding transaction blast radius.

## Consequences

- Discover routes unavailable when flag is on — FE must ship picks screen before flag flip.
- Thin pools may yield 1 or 0 picks (quality over fill rate).
- `ReplacePendingBatch` must delete prior pending batch (cross-date partial unique index).
- `resolveVisibleBatch` needs per-user advisory lock (writes on read path).

## Supersedes / Superseded by

- Supersedes: discover-as-primary-loop (behavioural, not deleted)
- Superseded by: —
