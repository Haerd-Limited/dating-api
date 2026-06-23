# Lawful Basis Register (Article 6 GDPR)

This document maps each processing purpose implemented in the Haerd Dating API to its proposed lawful basis. **Legal sign-off required** before treating this as authoritative.

| Processing purpose | Data involved | Lawful basis (proposed) | Code / schema evidence | Notes |
|---|---|---|---|---|
| Account creation & authentication (OTP) | Phone, email, verification codes | **Contract** (Art. 6(1)(b)) — necessary to provide the service | `internal/auth/`, `verification_codes` table | OTP codes hashed at rest |
| Profile & onboarding | Name, birthdate, photos, voice, location, preferences | **Contract** (Art. 6(1)(b)) | `internal/onboarding/`, `internal/profile/`, `user_profiles` | 18+ enforced server-side |
| Matching & discovery | Profile attributes, swipes, compatibility answers | **Contract** (Art. 6(1)(b)) | `internal/discover/`, `internal/interaction/`, `internal/compatibility/` | Core product feature |
| Messaging | Message text, voice notes | **Contract** (Art. 6(1)(b)) | `internal/conversation/` | Between matched users |
| Push notifications | Device tokens | **Contract** (Art. 6(1)(b)) or **Consent** if marketing | `device_tokens`, `internal/notification/` | **Legal review:** transactional vs promotional |
| Analytics & product insights | Events, session data, insight snapshots | **Consent** (Art. 6(1)(a)) — opt-out available | `events`, `user_preferences.analytics_opt_out`, `pkg/commonlibrary/analytics/` | Opt-out enforced at `Track()` |
| Safety & moderation | Reports, blocks, verification videos | **Legitimate interest** (Art. 6(1)(f)) — platform safety | `internal/safety/`, `internal/verification/` | Balance test needed |
| Privacy policy & terms acceptance | Consent type, version, accepted_at, IP, user-agent | **Consent** (Art. 6(1)(a)) | `user_consents`, `internal/consent/`, consent gate middleware | Gate disabled until frontend ships (HAE-433) |
| Data export (subject access) | All user-linked data | **Legal obligation** (Art. 6(1)(c)) | `internal/dataexport/`, `GET /users/me/data-export` | Rate-limited 24h |
| Account deletion | All user data | **Legal obligation** (Art. 6(1)(c)) / erasure right | `internal/user/service.go` `DeleteAccount` | S3 + DB cascade |
| Admin audit log | Admin IP, token fingerprint, path, target | **Legitimate interest** (Art. 6(1)(f)) — security & breach reconstruction | `admin_audit_log`, `internal/auditlog/` | No raw admin token stored |
| Retention purge | Expired codes, events, tokens | **Legal obligation** (Art. 6(1)(c)) / storage limitation | `internal/retention/`, `pkg/commonlibrary/constants/retention.go` | Daily job |

## Items requiring legal sign-off

1. Whether push notifications are contract-necessary or require separate consent.
2. Legitimate-interest balance test for safety/moderation processing.
3. Whether analytics opt-out satisfies consent requirements or explicit opt-in is required at launch.
4. Alignment with published Privacy Policy and Terms at https://haerd.com/privacy and https://haerd.com/terms (content gaps tracked in HAE-436).

## Related controls

- Consent versions: `pkg/commonlibrary/constants/consent.go`
- Consent gate feature flag: `ENABLE_CONSENT_GATE` (default `false`)
- Analytics opt-out: `PATCH /users/me/preferences/analytics-opt-out`
