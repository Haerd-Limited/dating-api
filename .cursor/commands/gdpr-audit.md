# GDPR Audit

Analyse the entire repository for GDPR compliance and write the results to a single markdown report at the repo root: `GDPR_COMPLIANCE_AUDIT.md`. Treat this as a thorough, evidence-based audit — every claim in the report must be backed by a concrete file path (with line numbers where helpful).

Read [.cursor/skills/gdpr-audit/SKILL.md](.cursor/skills/gdpr-audit/SKILL.md) first — it contains the categories, severity rubric, output structure, and anti-patterns. The command captures the workflow; the skill captures the principles.

## Step 1: Survey the repository

Build a picture of where personal data lives and flows before judging anything. Run these in parallel:

- `Glob` for `migrations/**/*.sql` to enumerate every table the codebase owns. Skim each migration to identify which tables hold personal data (name, email, photos, voice, location, messages, analytics, verification, device tokens, IPs, etc.) and which have `ON DELETE CASCADE` to `users`.
- `Glob` for `internal/**/domain/domain.go` to enumerate domains and the personal data each one models.
- `Read` `cmd/main.go` and `internal/http/router/router.go` to see which endpoints are registered, what middleware applies (auth, rate limiting, analytics opt-out, consent), and which handlers are public vs. authenticated.
- `Grep` for terms that surface privacy-relevant code paths, e.g. `DeleteAccount`, `OptOut`, `consent`, `gdpr`, `analytics`, `Track(`, `s3`, `email`, `phone`, `device_token`, `ip_address`, `user_agent`, `location`, `latitude`, `longitude`, `Rekognition`, `OpenAI`, `Twilio`, `Stripe`, `Sentry`, `RudderStack`, `password`, `bcrypt`, `encrypt`.
- `Read` `pkg/commonlibrary/analytics/analytics.go` (or equivalent) to check opt-out enforcement and which third parties receive events.
- `Read` `internal/user/service.go` (or wherever `DeleteAccount` lives) to inventory exactly what is and isn't deleted on account deletion.

Verify every concrete claim before it lands in the report. Plans/audits drift from reality fast — don't trust prior reports including the existing `GDPR_COMPLIANCE_AUDIT.md`. Re-derive every finding from the current code.

## Step 2: Audit against each GDPR area

Walk through each area defined in the skill (`.cursor/skills/gdpr-audit/SKILL.md` → "Audit areas"). For each area, produce:

- **Status:** ✅ Compliant / ⚠️ Partially compliant / ❌ Non-compliant / ❔ Needs verification (only when the codebase truly cannot tell you).
- **Evidence:** at least one file path (with line range where it sharpens the point) for every concrete claim — both for what works and what's missing.
- **Gap:** what is missing and which GDPR article it implicates.
- **Recommended fix:** concrete remediation. If a schema change is needed, sketch the migration. If a service method is needed, sketch the signature. Keep it proportional — don't ghost-write the whole feature.
- **Severity:** 🔴 High / 🟡 Medium / 🟢 Low using the rubric in the skill.

If two areas share a root cause (e.g. analytics events have no FK to `users`, which causes both incomplete deletion and unbounded retention), say so once and cross-reference rather than repeating the fix.

## Step 3: Write the report

Write to `GDPR_COMPLIANCE_AUDIT.md` at the repo root (overwrite if it exists — this is the canonical, current state).

Use the structure defined in the skill:

1. **Header** — date (today), application name (Haerd Dating API), overall compliance status.
2. **Executive summary** — 3-6 bullets of critical issues, 3-6 bullets of positive findings.
3. **Detailed findings** — one numbered subsection per audit area, in severity order (high → low). Each subsection follows the Status / Evidence / Gap / Recommended fix / Severity template.
4. **Cross-cutting risks** — short section for issues that span multiple areas (e.g. missing FK constraints on analytics tables, third-party SDKs without DPAs).
5. **Prioritised remediation roadmap** — a flat checklist ordered by severity then dependency, each item one line, each item mapping to one or more findings above.
6. **Out of scope / cannot determine from code** — explicit list of things the audit cannot answer from the codebase alone (DPA contracts, hosting region, actual privacy policy text, breach response runbook, internal access controls). This is not a hedge — be specific so the human reviewer knows what they still owe.

## Step 4: Self-check

Before reporting back to the user, verify:

- Every finding cites at least one file path. No vague "the codebase doesn't seem to…".
- Every "Non-compliant" or "Partially compliant" claim has a recommended fix.
- The report's executive summary matches the detailed findings (no surprise high-severity items only mentioned in detail).
- The roadmap items are actionable (a developer could pick one up without re-reading the audit).
- The "Out of scope" section is non-empty — there is always something a static code audit can't determine.

Then summarise to the user: total findings by severity, the path to the report, and the top 3 items they should address first. Do not modify any other file in the repo during this command.
