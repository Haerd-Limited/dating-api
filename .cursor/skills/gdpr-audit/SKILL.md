---
name: gdpr-audit
description: Audit the repository for GDPR compliance and write a structured markdown report to GDPR_COMPLIANCE_AUDIT.md at the repo root. Use when the user asks for a GDPR audit, privacy review, data-protection check, or "is this codebase GDPR compliant" report.
---

# GDPR Audit

This skill defines how to audit this repository for GDPR compliance and produce a single markdown report. It pairs with the `/gdpr-audit` slash command (`.cursor/commands/gdpr-audit.md`) — the command captures the procedural workflow, this skill captures the principles, audit areas, severity rubric, and output structure.

## Core principle

**Every finding must be backed by a file path.** A GDPR audit that says "the codebase doesn't seem to handle X" is useless. A finding is only credible if it points at concrete code (or the absence of expected code in a specific place). Verify every claim with `Grep`, `Glob`, and `Read` — never trust a previous audit, including any existing `GDPR_COMPLIANCE_AUDIT.md`.

## Audit areas

Walk through every area below. Skipping an area is only acceptable if the repo provably has no surface for it (e.g. no children's data because there's a hard age check at signup — and you've cited that check).

### 1. Right to access / data portability — Articles 15, 20

- Is there an endpoint that returns *all* of a user's personal data in a structured format (JSON/CSV)?
- Does it cover every table that holds user-linked data (profile, preferences, photos, voice, messages, matches, swipes, feedback, analytics events, insights, verification attempts, consents, device tokens)?
- Is it rate-limited and audit-logged?

### 2. Right to erasure — Article 17

- Is there a `DeleteAccount` (or equivalent) endpoint? Find it and read it end-to-end.
- For *every* table that references `users`, does deletion either CASCADE from the FK or get explicitly handled in `DeleteAccount`?
- Are S3 / blob-storage objects deleted (photos, voice prompts, verification videos)?
- Are third-party records deleted (Rekognition, OpenAI, analytics destinations, push-notification tokens, payment provider customer IDs)?
- Are logs that contain personal data anonymised or purged?

### 3. Right to rectification — Article 16

- Can users update their own data through standard endpoints? (Usually yes; confirm with paths.)

### 4. Right to object / restrict processing — Articles 18, 21

- Are there opt-outs for analytics / marketing / profiling? Where are they enforced — at write time, at read time, or both?
- Is the opt-out *checked at every call site*, or only in a central `Track()` method that some code paths bypass?

### 5. Lawful basis & consent — Articles 6, 7, 13

- Is there a consent table, with versioning, accepted-at, IP, user-agent, and revocation?
- Is privacy-policy / terms acceptance tracked at signup?
- Is the legal basis documented anywhere (code comments, README, schema)?

### 6. Data retention & minimisation — Article 5.1(c) and 5.1(e)

- Are retention periods defined anywhere (constants, config, scheduled jobs)?
- Is there an automated purge for analytics events, old verification attempts, expired tokens, inactive conversations, audit logs?
- Does the schema collect more than it needs (e.g. raw IP for every request when a coarse country code would do)?

### 7. Third-party processors & international transfers — Articles 28, 44–49

- List every external service the code calls: AWS S3 / Rekognition, OpenAI, RudderStack / analytics destinations, Twilio / SMS, Sentry, Stripe, push-notification providers.
- Note the region implied by SDK config where visible.
- Flag that DPAs / SCCs / hosting region are non-code artefacts the human reviewer must confirm.

### 8. Security of processing — Article 32

- Are passwords hashed? With what algorithm? (Look for `bcrypt`, `argon2`.)
- Are sensitive fields encrypted at rest where appropriate?
- Are secrets read from env / secret manager rather than committed?
- Is TLS enforced at the edge? (Usually infra; flag as out-of-scope but call out anything visible in code.)
- Are there obvious injection / authz risks visible in handlers (raw SQL, missing user-ID-from-context checks)?

### 9. Logging & telemetry hygiene

- Do logs include personal data (emails, phone numbers, raw request bodies)?
- Is there a structured logger config that scrubs known-sensitive fields?

### 10. Children's data — Article 8

- Is there an age check? Where? Is it enforced server-side?

### 11. Breach notification readiness — Articles 33, 34

- This is largely process, not code, but: is there any structured audit trail that would help reconstruct a breach (admin-action logs, login history, data-export logs)?

### 12. Data protection by design — Article 25

- Are new domains scaffolded with privacy in mind (CASCADE FKs, opt-out respected, no over-collection)? Cite a couple of recent migrations as evidence either way.

## Severity rubric

- 🔴 **High** — directly blocks a fundamental data-subject right (erasure, access, consent), or exposes personal data inappropriately. Anything that would cause a regulator to act first.
- 🟡 **Medium** — partial compliance, weak controls, or process gaps that don't immediately violate a right but accumulate risk (incomplete retention policy, opt-out not enforced everywhere, missing audit logging).
- 🟢 **Low** — hygiene items, documentation gaps, or improvements that strengthen compliance posture without addressing a current violation.

When in doubt between two levels, pick the higher one and explain the trade-off in the finding.

## Output structure

Write the report to `GDPR_COMPLIANCE_AUDIT.md` at the repo root, overwriting any prior version. Structure:

```markdown
# GDPR Compliance Audit Report

**Date:** <today, human-readable>
**Application:** <app name from go.mod / README>
**Status:** <✅ Compliant | ⚠️ Partially compliant | ❌ Non-compliant>

---

## Executive Summary

**Critical issues:**
- <bullet — one line each, each one maps to a numbered finding below>

**Positive findings:**
- <bullet — what the codebase already gets right, with file path>

---

## Detailed Findings

### 1. <Audit area> — <Article(s)> <status emoji>

**Status:** <Compliant / Partially compliant / Non-compliant / Needs verification>

**Evidence:**
- `<file path>:<line range>` — <what this code does or fails to do>
- `<file path>:<line range>` — <…>

**Gap:** <what's missing and which article it implicates>

**Recommended fix:**
<concrete remediation: schema sketch, service method signature, endpoint contract — proportional, not a full implementation>

**Severity:** 🔴 High / 🟡 Medium / 🟢 Low

---

### 2. <next area> …
```

After the per-area findings:

```markdown
## Cross-cutting Risks

- <short list of issues that span multiple areas, e.g. "Analytics tables lack FK to users — causes both incomplete deletion (§2) and unbounded retention (§6).">

## Prioritised Remediation Roadmap

1. [ ] <action item> — addresses §<finding number>(s). Severity: 🔴/🟡/🟢.
2. [ ] …

## Out of Scope / Cannot Determine from Code

- <DPA / SCC contracts with third-party processors>
- <Hosting region and physical data location>
- <Privacy policy and terms-of-service text>
- <Internal access controls, employee training, breach response runbook>
- <anything else the static code audit cannot answer>
```

Order detailed findings by severity (high → low). The roadmap orders by severity then by dependency (a fix that unblocks others comes first).

## Anti-patterns

- **Vague claims.** "The codebase doesn't appear to handle deletion well" is useless. "`internal/user/service.go:216-250` deletes from `users` and S3 but never touches `events`, `feedback`, or `insight_snapshots`, none of which have CASCADE FKs (`migrations/20251116120000_create_analytics_tables.sql`)" is useful.
- **Trusting the previous audit.** Re-derive everything from the current code. Schema and handlers change.
- **Manufacturing high-severity items.** A polish-level concern is 🟢, not 🔴. Inflated severity makes the report unactionable.
- **Ghost-writing the whole feature.** Recommended fixes should sketch the shape (table columns, method signature, endpoint path), not implement the feature.
- **Hiding what the audit can't answer.** The "Out of Scope" section is mandatory — it tells the human reviewer what they still need to verify off-code.
- **Editing other files.** This skill produces exactly one artefact: `GDPR_COMPLIANCE_AUDIT.md` at the repo root. No code changes.

## Tone

- Be specific, evidence-based, and proportional. The report's job is to make non-compliance impossible to dismiss and easy to fix.
- Distinguish "this is wrong" from "this is missing" from "this can't be determined from code". All three are valid findings, but they need different framings.
- Don't apologise for findings. The audit is not adversarial; it's a tool the team uses to prioritise.
