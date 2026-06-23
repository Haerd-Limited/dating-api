# Processor Register (Article 28 GDPR)

Sub-processors identified from the codebase. **DPA status and hosting region must be confirmed by legal/ops** — this register is a starting point derived from code, not a signed contract record.

| Processor | Service | Data shared | Purpose | Region (config) | DPA status |
|---|---|---|---|---|---|
| **Amazon Web Services (S3)** | Object storage | Profile photos, voice prompts, message media, verification videos | Media upload/download | `AWS_REGION` env (`internal/config/config.go`) | **TODO — confirm DPA/SCCs** |
| **Amazon Web Services (Rekognition)** | Face/video analysis | Verification video frames | Identity verification | `AWS_REKOGNITION_REGION` env | **TODO — confirm DPA/SCCs** |
| **OpenAI** | LLM API | Profile text (transcripts, prompts) for moderation/enrichment | Content processing | Not pinned in code (API endpoint) | **TODO — confirm DPA/SCCs** |
| **Twilio** | SMS | Phone numbers, OTP codes, notification SMS | Authentication & alerts | Twilio account region (not in code) | **TODO — confirm DPA/SCCs** |
| **Railway** (inferred) | Hosting | All DB and app data at rest | Infrastructure | **TODO — confirm region** | **TODO** |

## Region configuration

The application reads AWS regions from environment variables:

```
AWS_REGION              # S3 bucket region (required)
AWS_REKOGNITION_REGION  # Rekognition API region (required)
```

Wiring: `cmd/main.go` passes these to S3 and Rekognition clients.

**EU data residency:** If required, both `AWS_REGION` and `AWS_REKOGNITION_REGION` should be set to an EU region (e.g. `eu-west-1`) in production `.env` / deployment config. The codebase does not enforce this — it follows whatever is configured.

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
