# Processor Register (Article 28 GDPR)

Sub-processors identified from the codebase. **DPA status and hosting region must be confirmed by legal/ops** — this register is a starting point derived from code, not a signed contract record.

| Processor | Service | Data shared | Purpose | Region (confirmed) | DPA status |
|---|---|---|---|---|---|
| **Amazon Web Services (S3)** | Object storage | Profile photos, voice prompts, message media, verification videos | Media upload/download | `eu-west-2` (London, UK) — confirmed from prod `AWS_REGION` | **TODO — confirm DPA/SCCs** |
| **Amazon Web Services (Rekognition)** | Face/video analysis | Verification video frames | Identity verification | `eu-west-1` (Ireland, EU/EEA) — confirmed from prod `AWS_REKOGNITION_REGION` | **TODO — confirm DPA/SCCs** |
| **OpenAI** | LLM API | Profile text (transcripts, prompts) for moderation/enrichment | Content processing | Not pinned in code (API endpoint; US-hosted) | **TODO — confirm DPA/SCCs** |
| **Twilio** | SMS | Phone numbers, OTP codes, notification SMS | Authentication & alerts | Twilio account region (not in code) | **TODO — confirm DPA/SCCs** |
| **Railway** | Hosting | All DB and app data at rest | Infrastructure | EU West (Amsterdam, Netherlands) — confirmed for `dating-api` service + PostGIS/Postgres | **TODO** |

## Step 0 — Hosting region confirmation (GDPR Art. 44–49)

Confirmed from the production environment on 2026-07-27:

| Service | Env var | Value | Location | EEA? |
|---|---|---|---|---|
| AWS S3 | `AWS_REGION` | `eu-west-2` | London, UK | No (UK) — covered by EU adequacy decision |
| AWS Rekognition | `AWS_REKOGNITION_REGION` | `eu-west-1` | Ireland, EU | Yes |
| Railway app | Railway dashboard | EU West | Amsterdam, Netherlands | Yes |
| Railway Postgres (PostGIS) | Railway dashboard | EU West | Amsterdam, Netherlands | Yes |

**Assessment:**

- **AWS S3 (`eu-west-2`, London):** Data at rest is in the UK, which is outside the EEA post-Brexit but benefits from the EU→UK adequacy decision, so EU user data can flow there without additional SCCs. UK GDPR applies. Transfer risk: **low**.
- **AWS Rekognition (`eu-west-1`, Ireland):** Fully within the EEA. Transfer risk: **none**.
- **Railway app + Postgres (EU West, Amsterdam):** Both the `dating-api` service and the PostGIS/Postgres database run in the Netherlands, fully within the EEA. All app and DB data at rest stays in the EEA. Transfer risk: **none**.
- **No AWS or hosting data is stored in the US.** The primary data stores (Postgres, S3) and compute are all in the EEA/UK, so the SCC/transfer piece is **not urgent** for infrastructure. The AWS GDPR DPA (Step 1) and Railway DPA (Step 4) still need to be confirmed/recorded.
- **OpenAI / Twilio:** US-hosted processors; cross-border transfers apply and rely on SCCs bundled into their DPAs (Steps 2–3).

The application reads AWS regions from environment variables (`AWS_REGION`, `AWS_REKOGNITION_REGION`), wired in `cmd/main.go` to the S3 and Rekognition clients. The codebase does not enforce a region — it follows whatever is configured in the deployment environment.

## Processors not found in code (verify separately)

- Push notification delivery (Expo/APNs/FCM) — device tokens stored locally; actual push delivery may involve additional processors.
- Email provider (if email OTP is enabled) — check `internal/communication/` for email channel.
- Error monitoring / APM (Sentry, Datadog, etc.) — not referenced in application code at time of writing.

## Cross-border transfers (Articles 44–49)

Any processor hosted outside the EEA requires appropriate safeguards (SCCs, adequacy decision, or explicit consent). **Legal must confirm** transfer mechanisms for each row above.

## Maintenance

Update this register when:
- A new external API/SDK is integrated
- AWS region or bucket configuration changes
- A DPA is signed or renewed

Related: [lawful-basis.md](./lawful-basis.md)
