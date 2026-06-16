package domain

import "time"

type (
	User struct {
		ID                   string
		Email                string
		PhoneNumber          string
		FirstName            string
		LastName             *string
		OnboardingStep       string
		HowDidYouHearAboutUs *string
		CreatedAt            time.Time
	}
)
