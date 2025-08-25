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
	case domain.LocationResult:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(v.OnboardingSteps),
			Content:         MapLocationContentToDto(v.Content),
		}
	case domain.LifestyleResult:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(v.OnboardingSteps),
			Content:         MapLifestyleContentToDto(v.Content),
		}
	case domain.BeliefsResult:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(v.OnboardingSteps),
			Content:         MapBeliefsContentToDto(v.Content),
		}
	case domain.BackgroundResult:
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

func MapBeliefsContentToDto(content domain.BeliefsContent) dto.BeliefsContent {
	var educationLevels []dto.EducationLevel
	for _, level := range content.EducationLevels {
		educationLevels = append(educationLevels, dto.EducationLevel{
			ID:    level.ID,
			Label: level.Label,
		})
	}

	var ethnicities []dto.Ethnicity
	for _, ethnicity := range content.Ethnicities {
		ethnicities = append(ethnicities, dto.Ethnicity{
			ID:    ethnicity.ID,
			Label: ethnicity.Label,
		})
	}

	return dto.BeliefsContent{
		Ethnicities:     ethnicities,
		EducationLevels: educationLevels,
	}
}

func MapLifestyleContentToDto(content domain.LifestyleContent) dto.LifestyleContent {
	var religions []dto.Religion
	for _, religion := range content.Religions {
		religions = append(religions, dto.Religion{
			ID:    religion.ID,
			Label: religion.Label,
		})
	}

	var politicalBeliefs []dto.PoliticalBelief
	for _, politicalBelief := range content.PoliticalBeliefs {
		politicalBeliefs = append(politicalBeliefs, dto.PoliticalBelief{
			ID:    politicalBelief.ID,
			Label: politicalBelief.Label,
		})
	}

	return dto.LifestyleContent{
		Religions:        religions,
		PoliticalBeliefs: politicalBeliefs,
	}
}

func MapLocationContentToDto(content domain.LocationContent) dto.LocationContent {
	var habits []dto.Habit
	for _, habit := range content.Habits {
		habits = append(habits, dto.Habit{
			ID:    habit.ID,
			Label: habit.Label,
		})
	}

	return dto.LocationContent{
		Habits: habits,
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
