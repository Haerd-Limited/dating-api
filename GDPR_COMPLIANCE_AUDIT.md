# GDPR Compliance Audit Report

**Date:** 23 June 2026
**Application:** Haerd Dating API (`github.com/Haerd-Limited/dating-api`)
**Status:** ⚠️ Partially compliant — core data-subject rights are implemented in code and several medium findings from the prior audit are resolved; remaining gaps are mainly consent-gate activation, off-code processor contracts, and minor export/retention refinements.

---

## Executive Summary

**Critical issues:**
- None identified at 🔴 High severity. No finding directly blocks erasure or access in code, and the prior PII-logging exposure has been remediated.

**Remaining issues (medium):**
- Consent infrastructure is complete but the gate is **disabled by default** (`ENABLE_CONSENT_GATE=false`) and depends on unshipped frontend work (HAE-433) plus policy content review (HAE-436), so production processing still lacks an enforced lawful-basis record (§5).
- Processor register exists in-repo but **DPA/SCC status and hosting region are still unconfirmed** — legal/ops must close the TODOs in `docs/processor-register.md` (§7).
- `admin_audit_log` rows accumulate indefinitely — the retention job does not purge them (§6).

**Positive findings:**
- Data export now covers 17 categories including `consents` and `device_tokens` (`internal/dataexport/service.go:229-238`, `domain/export.go:23-24`).
- PII logging at error sites has been removed; hygiene convention documented in `pkg/commonlibrary/logger/logger.go:38-40` and `AGENTS.md`.
- Admin access audit log records every `/admin` request with IP, token fingerprint, path, target, and status (`internal/middleware/admin_audit.go:27-64`, `migrations/20260623130000_create_admin_audit_log.sql`).
- Full account erasure with S3 purge, verification-code cleanup, and CASCADE across personal-data tables (`internal/user/service.go:242-294`, `migrations/20260224140000_add_analytics_cascade_on_user_delete.sql`).
- Analytics opt-out enforced centrally at `Track()` via `OptOutFunc` (`pkg/commonlibrary/analytics/analytics.go:98-101`, `cmd/main.go:160-177`).
- Automated retention purge for verification codes, events, failed verification attempts, and expired refresh tokens (`internal/retention/service.go:37-95`).
- Lawful-basis and processor registers documented at `docs/lawful-basis.md` and `docs/processor-register.md`.

---

## Detailed Findings

### 1. Lawful basis & consent — Articles 6, 7, 13 ⚠️

**Status:** Partially compliant

**Evidence:**
- `migrations/20260623120000_create_user_consents.sql:3-14` — versioned consent ledger with `accepted_at`, `revoked_at`, `ip`, `user_agent`, CASCADE on user delete.
- `internal/consent/service.go:20-25` + `internal/api/consent/handler.go` — record/list/revoke API.
- `internal/middleware/consent_required.go` + `internal/http/router/router.go:177-178` — gate returns `403 consent_required` when enabled.
- `internal/config/config.go:39,53` — `EnableConsentGate` defaults to `false` via `viper.SetDefault("ENABLE_CONSENT_GATE", false)`.
- `pkg/commonlibrary/constants/consent.go:9-10` — versions pinned to `2026-05-28` matching published policy date.
- `docs/lawful-basis.md` — in-repo Article 6 mapping with legal sign-off flags (added since prior audit).

**Gap:** Production users are not blocked from gated processing without a consent record because the feature flag is off and the frontend consent flow is not yet shipped. Policy content gaps on haerd.com remain (HAE-436). Articles 6/7/13 require informed, recorded consent *before* non-contract processing in production.

**Recommended fix:** Ship frontend consent recording (HAE-433), close policy content gaps (HAE-436), then set `ENABLE_CONSENT_GATE=true` in production. Legal should sign off on `docs/lawful-basis.md`.

**Severity:** 🟡 Medium

---

### 2. Third-party processors & international transfers — Articles 28, 44–49 ❔

**Status:** Needs verification (process largely off-code)

**Evidence:**
- AWS S3/Rekognition wired via `cfg.AWSRegion` / `cfg.AWSRekognitionRegion` (`internal/config/config.go:21-22`, `cmd/main.go`).
- OpenAI: `internal/openai`, `OpenAIAPIKey` in config.
- Twilio SMS: `internal/communication/service.go:12-56`.
- `docs/processor-register.md:5-11` — register lists processors with **TODO — confirm DPA/SCCs** on every row.

**Gap:** Code identifies sub-processors but cannot prove DPAs, SCCs, or physical data location. Articles 28 and 44–49 require contractual and transfer safeguards confirmed outside the codebase.

**Recommended fix:** Legal/ops to complete DPA status column in `docs/processor-register.md`, confirm EU region in production env if residency is required, and document transfer mechanisms per processor.

**Severity:** 🟡 Medium

---

### 3. Data retention & minimisation — Article 5.1(c), 5.1(e) ⚠️

**Status:** Partially compliant

**Evidence:**
- `internal/retention/service.go:37-95` — daily purge of verification codes (30d), events (2y), failed verification attempts (1y), expired refresh tokens (7d grace).
- `pkg/commonlibrary/constants/retention.go` — TTLs centralised.
- `migrations/20260623130000_create_admin_audit_log.sql` — new audit table with no purge job.
- `insight_snapshots`, `wrapped_annual`, conversations/messages — no inactivity-based deletion.
- `user_consents.ip` and `verification_codes.request_ip` store full IPs (`INET`).

**Gap:** `admin_audit_log` will grow without bound now that every admin request is logged (`internal/middleware/admin_audit.go:27-64`). Storage limitation (Art. 5.1(e)) requires a defined retention window for audit rows (e.g. 1–2 years). Insight snapshots and dormant conversations remain indefinitely.

**Recommended fix:** Add `admin_audit_log` to `retention.PurgeOnce` with a constant (e.g. `RetentionAdminAuditLog = 2 * 365 * 24 * time.Hour`). Optionally add TTLs for stale `insight_snapshots`/`wrapped_annual`.

**Severity:** 🟡 Medium

---

### 4. Right to access / data portability — Articles 15, 20 ⚠️

**Status:** Partially compliant (minor omissions)

**Evidence:**
- `internal/dataexport/service.go:83-112` — `ExportUserData` with 24h rate limit (`rateLimitWindow`, line 29) and request logging via `data_export_requests` (`storage/repository.go:47`).
- `internal/dataexport/service.go:115-241` — exports account, profile, preferences, photos, voice prompts, swipes, matches, conversations+messages, feedback, events, insight snapshots, verification attempts, blocks, reports, matching answers, **consents**, and **device tokens**.
- `internal/dataexport/domain/export.go:191-201` — `ConsentExport` omits raw `ip`/`user_agent` (appropriate minimisation in the export artifact).
- `internal/http/router/router.go:170` — `GET /users/me/data-export` carved out of consent gate.

**Gap:** Three user-linked tables are still absent from the export payload: `feedback_attachments` (media URLs tied to feedback), `wrapped_annual` (yearly user artifact), and `analytics_opt_out` (stored on `user_preferences` but not in `StoredDiscoverPreferences` — `internal/discover/domain/preferences.go:17-27`). Export request history (`data_export_requests`) is also not returned to the user. Article 15 requires all personal data — these are minor but real gaps.

**Recommended fix:** Add `FeedbackAttachments`, `WrappedAnnual`, and `AnalyticsOptOut bool` to `ExportPayload`; populate from feedback repo, insights repo, and `preferenceRepo.IsAnalyticsOptedOut`. Optionally include export-request timestamps.

**Severity:** 🟢 Low

---

### 5. Breach notification readiness — Articles 33, 34 ⚠️

**Status:** Partially compliant (improved)

**Evidence:**
- `internal/middleware/admin_audit.go:27-64` + `internal/auditlog/storage/repository.go:25-28` — persistent log of every admin API access (method, path, IP, token fingerprint, target, status).
- `migrations/20260224120000_create_data_export_requests.sql` — records each data-export request per user.
- `report_actions` table (`migrations/20251111091500_create_safety_tables.sql:50-58`) — moderation action trail.

**Gap:** No user login/session history table, no dedicated breach-response runbook in repo, and admin audit rows have no retention policy yet (§3). Process items (notification timelines, DPO contact) are off-code.

**Recommended fix:** Add admin-audit retention (§3). Document a breach-response runbook off-code. Consider logging failed auth attempts if login-history reconstruction becomes a requirement.

**Severity:** 🟢 Low

---

### 6. Right to erasure — Article 17 ✅

**Status:** Compliant (with low-severity caveats)

**Evidence:**
- `internal/user/service.go:242-294` — S3 purge, verification-code purge by phone/email, then user row delete.
- CASCADE verified: profiles, preferences, photos, voice, matches, messages, device tokens, safety tables, consents, analytics (`events`, `insight_snapshots`, `wrapped_annual` via `migrations/20260224140000_add_analytics_cascade_on_user_delete.sql`), feedback (`migrations/20251122000001_create_feedback_system.sql:14`).
- `internal/http/router/router.go:171` — `DELETE /users/me` bypasses consent gate.

**Gap:** Third-party durable copies (Rekognition, OpenAI, Twilio logs) are not actively deleted from code. S3 deletion failure is logged but does not abort DB deletion (lines 255-258) — acceptable trade-off but leaves orphan blobs if S3 fails.

**Recommended fix:** Document third-party deletion posture; add integration test asserting `DeleteAllUserFiles` covers all S3 prefixes.

**Severity:** 🟢 Low

---

### 7. Logging & telemetry hygiene — Article 32 / 5.1(c) ✅

**Status:** Compliant

**Evidence:**
- Prior `zap.Any("userProfile")`, message, swipe, coordinate, and phone-number log sites removed — confirmed: only one `zap.Any` remains in `internal/conversation/score/service.go:143` and it logs scoring config/length, not user content.
- `pkg/commonlibrary/logger/logger.go:38-40` — documented no-whole-struct-logging rule.
- `AGENTS.md` Best Practices item 10 — logging hygiene guideline for agents.

**Gap:** No automatic key-redaction encoder in zap (convention-only guardrail). Chi `middleware.Logger` (`router.go:105`) logs method/path/status for all requests — standard and acceptable.

**Recommended fix:** None required immediately. Optional: add a lint rule or code review checklist item banning `zap.Any` on domain structs.

**Severity:** 🟢 Low

---

### 8. Right to object / restrict processing — Articles 18, 21 ✅

**Status:** Compliant (for analytics)

**Evidence:**
- `pkg/commonlibrary/analytics/analytics.go:98-101` — `OptOutFunc` checked at single chokepoint; all six `Track()` call sites inherit it.
- `cmd/main.go:160-177` — opt-out read from DB with 60s cache; defaults to not-opted-out only on lookup error.
- `internal/http/router/router.go` — `PATCH /users/me/preferences/analytics-opt-out` inside gated routes.
- `migrations/20251116123000_add_analytics_opt_out_to_user_preferences.sql`.

**Gap:** `analytics_opt_out` is not returned on preferences read (`StoredDiscoverPreferences` lacks the field — `internal/discover/domain/preferences.go:17-27`), so the app cannot display current state without a separate call.

**Recommended fix:** Include `analytics_opt_out` in the preferences read DTO/response.

**Severity:** 🟢 Low

---

### 9. Right to rectification — Article 16 ✅

**Status:** Compliant

**Evidence:**
- Profile/onboarding update routes under `/onboarding/*` and `/users/me` (`internal/http/router/router.go:186-201, 255-268`).
- Preferences editable via discover endpoints (`internal/discover/service.go`).

**Gap:** None material.

**Severity:** 🟢 Low

---

### 10. Security of processing — Article 32 ⚠️

**Status:** Partially compliant

**Evidence:**
- OTP-only auth; verification codes hashed (`migrations/20250830172910_create_verification_codes.sql:8`).
- Secrets from env via viper (`internal/config/config.go`).
- Parameterised SQL throughout repositories; auth via context user ID in handlers.
- Admin routes protected by shared API key (`internal/middleware/admin_middleware.go:10-26`).

**Gap:** Profile fields not encrypted at application layer; TLS and disk encryption are infra-level. Raw IPs stored on consents and verification codes.

**Recommended fix:** Confirm at-rest encryption at DB/volume layer; document TLS at edge. Evaluate IP minimisation if threat model warrants.

**Severity:** 🟢 Low

---

### 11. Children's data — Article 8 ✅

**Status:** Compliant

**Evidence:**
- `internal/profile/private_methods.go:54-56` — rejects birthdates implying age `< constants.MinAge`.
- `pkg/commonlibrary/constants/constants.go` — `MinAge = 18`.

**Gap:** Self-declared age only (no ID verification) — standard but noted.

**Severity:** 🟢 Low

---

### 12. Data protection by design — Article 25 ✅

**Status:** Compliant

**Evidence:**
- Recent privacy-positive migrations: `user_consents` with CASCADE (`20260623120000`), analytics CASCADE FKs (`20260224140000`), `admin_audit_log` with token fingerprint not raw token (`20260623130000`, `internal/middleware/admin_audit.go:59-62`).
- Consent export deliberately omits IP/user-agent (`internal/dataexport/service.go:698-716`).
- Export and audit features added following existing domain/repository/service layering.

**Gap:** New tables should continue to be added to export payload checklist (§4 residual gaps).

**Severity:** 🟢 Low

---

## Cross-cutting Risks

- **Consent dormant in prod (§1 ↔ §5).** Ledger, API, and middleware exist but the flag is off and no frontend writes rows — the system neither enforces nor fully exposes consent state in production yet.
- **Off-code processor obligations (§2 ↔ §7).** `docs/processor-register.md` closes the documentation gap but DPA/SCC confirmation remains a human blocker for lawful international transfers.
- **Audit log growth (§3 ↔ §5).** Every admin request now creates a row; without retention this becomes both a storage-limitation issue and a future breach-reconstruction dataset that itself needs a policy.
- **Export completeness drift (§4 ↔ §12).** As new user-linked tables are added, the export assembler must be updated in the same PR — no automated check enforces parity today.

## Prioritised Remediation Roadmap

1. [ ] Ship frontend consent flow (HAE-433) + close policy content gaps (HAE-436), then enable `ENABLE_CONSENT_GATE=true` in production — addresses §1. Severity: 🟡.
2. [ ] Legal/ops: complete DPA/SCC column in `docs/processor-register.md` and confirm production AWS/hosting regions — addresses §2. Severity: 🟡.
3. [ ] Add `admin_audit_log` retention to `internal/retention/service.go` with a named constant — addresses §3. Severity: 🟡.
4. [ ] Add `feedback_attachments`, `wrapped_annual`, and `analytics_opt_out` to data export — addresses §4. Severity: 🟢.
5. [ ] Surface `analytics_opt_out` on preferences read response — addresses §8. Severity: 🟢.
6. [ ] Document breach-response runbook and third-party deletion posture off-code — addresses §5. Severity: 🟢.
7. [ ] Add TTL for `insight_snapshots`/`wrapped_annual` and review raw-IP necessity — addresses §3. Severity: 🟢.

## Out of Scope / Cannot Determine from Code

- Signed DPA / SCC contracts with AWS, OpenAI, Twilio, and push-delivery providers (§2).
- Actual hosting region and physical data residency of DB, S3, logs, and Railway deployment (§2).
- Whether published Privacy Policy and Terms at https://haerd.com/privacy and https://haerd.com/terms disclose all processing purposes (content review HAE-436).
- TLS termination and disk-level encryption at the infrastructure layer (§10).
- Internal employee access controls, training, and breach notification runbook/timelines (§5).
- Whether third-party processors retain durable copies after local erasure (§6).
- Legal sign-off on proposed lawful bases in `docs/lawful-basis.md`.
