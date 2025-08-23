package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func ToOnboardingResponse(result any) dto.OnboardingResponse {
	switch v := result.(type) {
	case domain.RegisterResult:
		return dto.OnboardingResponse{
			AccessToken:     &v.AccessToken,
			RefreshToken:    &v.RefreshToken,
			OnboardingSteps: mapOnboardingStepsToDto(v.OnboardingSteps),
			Content:         MapRegisterContentToDto(v.Content),
		}
	case domain.BasicsResult:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(v.OnboardingSteps),
		}
	default:
		return dto.OnboardingResponse{}
	}
}

func mapOnboardingStepsToDto(steps domain.OnboardingSteps) dto.OnboardingSteps {
	var stringSteps []string

	for _, step := range steps.Steps {
		stringSteps = append(stringSteps, string(step))
	}

	return dto.OnboardingSteps{
		PreviousStep: string(steps.PreviousStep),
		CurrentStep:  string(steps.CurrentStep),
		NextStep:     string(steps.NextStep),
		Progress:     steps.Progress,
		Steps:        stringSteps,
		TotalSteps:   steps.TotalSteps,
	}
}

func MapRegisterContentToDto(content domain.RegisterContent) dto.RegisterContent {
	var datingIntentions []dto.DatingIntention
	for _, intention := range content.DatingIntentions {
		datingIntentions = append(datingIntentions, dto.DatingIntention{
			ID:    intention.ID,
			Label: intention.Label,
		})
	}

	var genders []dto.Gender
	for _, gender := range content.Genders {
		genders = append(genders, dto.Gender{
			ID:    gender.ID,
			Label: gender.Label,
		})
	}

	return dto.RegisterContent{
		DatingIntentions: datingIntentions,
		Genders:          genders,
	}
}
