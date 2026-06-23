package constants

const (
	ConsentTypePrivacyPolicy  = "privacy_policy"
	ConsentTypeTermsOfService = "terms_of_service"

	// Versions must match the "Last updated" date of the published documents at
	// https://haerd.com/privacy and https://haerd.com/terms.
	CurrentPrivacyPolicyVersion  = "2026-05-28"
	CurrentTermsOfServiceVersion = "2026-05-28"
)

var MandatoryConsentTypes = []string{ConsentTypePrivacyPolicy, ConsentTypeTermsOfService}

func CurrentConsentVersion(consentType string) string {
	switch consentType {
	case ConsentTypePrivacyPolicy:
		return CurrentPrivacyPolicyVersion
	case ConsentTypeTermsOfService:
		return CurrentTermsOfServiceVersion
	default:
		return ""
	}
}
