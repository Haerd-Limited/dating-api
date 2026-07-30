# 0002 — Question Packs as Onboarding Step

- **Status:** Accepted
- **Date:** 2026-07-30
- **Feature:** onboarding / compatibility
- **Plan:** `onboarding-question-packs_a1e8e464.plan.md` (lives outside the repo in `~/.cursor/plans/`)
- **Linear:** —
- **Reconstructed:** yes

## Context

The app was rejected from the App Store. One required change: ensure every user completes 100% of compatibility question packs before entering the matching pool, so compatibility scoring works for all users. The step must be resumable like other onboarding steps and placed before the final verification step.

## Decision

Insert `QUESTION_PACKS` into `OrderedSteps` between `PROMPTS` and `VIDEO_VERIFICATION`. Advancement requires `answered >= total` active questions (via `IsQuestionPacksComplete`). New commit endpoint `POST /api/v1/onboarding/question-packs` with no body. `GET /api/v1/onboarding/step` returns compatibility overview when current step is `QUESTION_PACKS`. **Grandfather** existing users — no migration resetting `onboarding_step`.

## Options considered

### New onboarding step before video verification — CHOSEN

Ensures question packs are mandatory for new signups without blocking users already past verification. Matches App Store requirement while minimising disruption.

### Force all existing users back to question packs — rejected

Would require a migration resetting `onboarding_step` for users at `VIDEO_VERIFICATION` or `COMPLETE`. High disruption; rejected in favour of grandfathering.

### Separate `question_packs_complete` column — rejected

Could track completion independently of onboarding step. Rejected — deriving completeness from `count(user_answers) >= count(active questions)` reuses existing data and avoids schema churn.

### Dedicated question-packs API only (no onboarding step) — rejected

FE could gate on completion without a step enum. Rejected because resumability and the standard onboarding flow (`ensureStep`, `bumpOnboardingStep`, `GET /step` content) require a first-class step.

### Backend-first deploy without FE coordination — rejected as sole strategy

BE can ship first, but old FE builds won't render `QUESTION_PACKS`. Accepted mitigation: coordinated same-day deploy or FE defensive fallback for unknown steps.

## Consequences

- `OnboardingStepsQuestionPacks` added to `internal/onboarding/domain/steps.go`.
- `compatibility.Service.IsQuestionPacksComplete` and `CountUserAnswers` repository method.
- `ErrQuestionPacksIncomplete` → 409 on premature commit.
- `TotalSteps` in onboarding responses increases by 1 automatically via `OrderedSteps`.
- Discover/matching gates unchanged for grandfathered users.

## Supersedes / Superseded by

- Supersedes: —
- Superseded by: —
