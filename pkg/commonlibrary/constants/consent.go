package constants

const (
	ConsentTypePrivacyPolicy  = "privacy_policy"
	ConsentTypeTermsOfService = "terms_of_service"

	CurrentPrivacyPolicyVersion  = "2026-04-30"
	CurrentTermsOfServiceVersion = "2026-04-30"
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
