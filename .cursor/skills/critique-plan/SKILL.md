---
name: critique-plan
description: Review and critique an implementation plan against the actual codebase, surface ambiguous or undecided sections, ask the user to resolve real ones, and patch the plan in place. Use when the user asks to review, critique, audit, sanity-check, or stress-test a plan file (typically under .cursor/plans/), or whenever the user references a plan file and asks for feedback before approving it.
---

# Critique Plan

This skill defines how to review an implementation plan in this repo. It pairs with the `/critique-plan` slash command (`.cursor/commands/critique-plan.md`) — the command captures the procedural workflow, and this skill captures the principles the agent applies any time the user asks for a plan review, even outside the command.

## Core principle

**Trust nothing in the plan that is verifiable in the codebase.** Plans drift from reality fast. Before judging the plan, verify every concrete claim with `Grep`, `Glob`, and `Read`. Specifically:

- File paths and line numbers (off-by-many is common).
- Struct names, interface methods, error variables.
- Whether a "dedicated handler" actually exists for the write path the plan describes — or whether it's bundled into a larger handler (very common in this repo's profile/onboarding domains).
- Whether interfaces have generated mocks (`go:generate mockgen` directive + sibling `_mock.go`) — adding methods requires `make mock`.
- Whether validators run unconditionally or only when a payload field is non-nil (`if x != nil { … }` gates are easy to miss and cause production regressions).
- Whether constants the plan bumps have other callers whose behaviour changes.

Run these verifications in parallel — never serialize lookups when they're independent.

## Categorise findings

Sort everything you find into three buckets:

### Bucket A — Real ambiguities (need a decision from the user)

A finding is a real ambiguity only if:
- Multiple valid choices materially change implementation, schema, API shape, or UX.
- The plan silently introduces a breaking change for existing users / data.
- The plan picks a user-visible value (column name, category string, magic constant) without input.
- The plan's enforcement scope is unclear (would this rule fire on unrelated edits? Block a read path?).

Rules for Bucket A:
- Cap at 3-4 questions per round.
- Ask all of them in a single `AskQuestion` call.
- Each option must label its tradeoff, not just the choice. Bad: "Yes / No". Good: "Grandfather existing users — only enforce on new submissions; today's validator already only fires on upsert, so this is automatic" vs. "Force re-onboard — add a migration that resets `onboarding_step` for affected users".
- If the user's answers raise new ambiguities, ask another short round before patching.

### Bucket B — Specification gaps (fix without asking)

Things that are obviously missing or wrong but where the right fix is unambiguous:
- Imprecise file paths → pin to exact paths and line numbers.
- Missing tooling steps the codebase requires (`make mock` after interface changes, `make entity` after migrations, `make lint`, etc.).
- Missing test cases that follow obviously from the rule being added (especially the negative-path / no-op-path that proves a gate works).
- Frontend / cross-team coordination notes when the plan changes an API contract.
- Wiring or mock regeneration the plan omitted.

Note these in the critique to the user, but don't ask permission — patch them in.

### Bucket C — Strengths

3-5 things the plan got right. This calibrates the user's trust and signals what not to change.

## Workflow

1. **Read the plan in full.** Re-read `.cursor/commands/create-implementation-plan.md` (the source of acceptance criteria).
2. **Verify in parallel.** Cross-check every concrete claim against the codebase.
3. **Sort findings into Buckets A/B/C.**
4. **Present the critique** in a single message:
   - Strengths (Bucket C).
   - Issues — split into "Real ambiguities — need your decision" (Bucket A), "Specification gaps I'll fix without asking" (Bucket B), and "Minor things".
   - One `AskQuestion` call covering all Bucket A items.
5. **Wait for answers.** Do nothing while waiting.
6. **Patch the plan in place** with `StrReplace`. Do NOT call `CreatePlan` — plan files are markdown and edited directly. Plan mode allows markdown edits to plan files.
   - Update the YAML frontmatter `todos:` list (new entries get `id`, `content`, `status: pending`).
   - Record every Bucket A decision in the `Decisions` section with rationale.
   - Absorb every Bucket B fix into the relevant `Parts` subsection. Pin file paths and line numbers.
   - Add new test cases to `Validation` for any new runtime behaviour.
7. **Self-check and report.** Run the plan against the acceptance criteria from `create-implementation-plan.md`. Report pass/fail per criterion. Ask whether to iterate again or treat as approved.

## Anti-patterns

- **Manufacturing ambiguities to look thorough.** If a section is fine, leave it alone. Bucket A is reserved for decisions that genuinely require the user.
- **Vague critique.** "Inaccurate file path" is useless. "§7 says handler `UpsertVoicePrompts` exists at `internal/api/profile/handler.go`, but voice-prompt writes flow through the general profile-update handler at line N" is useful.
- **Re-creating the plan.** Never call `CreatePlan` again — patch in place with `StrReplace`. Calling `CreatePlan` would create a sibling file and orphan the user's annotations.
- **Patching before the user answers.** Bucket A questions block the patch step. Apply Bucket B fixes only after Bucket A is resolved (some Bucket B items may evolve based on Bucket A answers).
- **Running `make` targets, editing non-plan files, or committing.** Plan mode rules apply throughout this skill.

## Tone

- Be specific.
- Distinguish "this is wrong" from "this is missing". Both warrant a fix; only the former warrants the framing "you got this wrong".
- The plan is a draft. The user owns it. The skill's job is to make it harder for the plan to lie to its future implementer.
