# 0001 — Two Active Match Cap

- **Status:** Accepted
- **Date:** 2026-04-30
- **Feature:** interaction / matches
- **Plan:** `two-match-cap_d61780bd.plan.md` (lives outside the repo in `~/.cursor/plans/`)
- **Linear:** [HAE-411](https://linear.app/haerd/issue/HAE-411/implement-2-matchactive-chat-limit-feature)
- **Reconstructed:** yes

## Context

Haerd needed a hard limit on simultaneously active matches so users focus on fewer conversations. The cap had to be enforced at match-creation time (reciprocal like), surfaced to the frontend for pre-gating, and safe under concurrent reciprocal-like races.

## Decision

Enforce a symmetric cap of **2 active matches per user** at `CreateSwipe` on the matchable branch. Reject with dedicated 409 errors if either actor or target is at cap. Surface `active_matches_count`, `match_slot_limit`, and per-like `target_at_match_limit` on `GET /api/v1/likes`. Use `pg_advisory_xact_lock` on both users inside the existing transaction before counting. Grandfather existing users over the cap — no backfill migration.

## Options considered

### Symmetric cap (actor OR target at 2) — CHOSEN

Both users must have room before a new match is created. Two distinct errors (`ErrMatchLimitReached`, `ErrTargetMatchLimitReached`) map to different user-facing 409 messages. Prevents either side from accumulating a 3rd match.

### Actor-only cap — rejected

Simpler to implement but allows a user at 0/2 to match someone already at 2/2, creating asymmetric overload on the target. Does not match the product intent of limiting active conversations for both parties.

### Swipe recorded even when at cap — rejected

Some designs record the like and defer match creation. Rejected because the plan explicitly requires the swipe **not** be persisted when the cap blocks match creation — the check runs before `InsertSwipe`.

### Optimistic count check without advisory locks — rejected

Two concurrent reciprocal likes could both pass the count check and create a 3rd match. Advisory locks (sorted user IDs to avoid deadlock) were chosen over application-level mutexes because the check and insert already run inside a Postgres transaction.

### New endpoint for match-slot status — rejected

A dedicated `GET /match-slots` would work but adds surface area. Rejected in favour of additive fields on the existing `GET /api/v1/likes` response the FE already calls.

### Backfill / force-unmatch over-quota users — rejected

Migration complexity and user disruption. Cap only gates **new** match creation; users already over 2 keep their matches.

## Consequences

- `maxActiveMatches = 2` constant in `internal/interaction/service.go`.
- `CountActiveMatches`, `CountActiveMatchesForUsers`, `LockUsersForMatchCreation` added to interaction repository.
- Frontend must handle two new 409 messages and render N/2 from likes response.
- Discover quota and match cap are independent — cap applies only on matchable branch.

## Supersedes / Superseded by

- Supersedes: —
- Superseded by: —
