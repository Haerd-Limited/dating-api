package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func ToOnboardingResponse(result domain.StepResult) dto.OnboardingResponse {
	switch v := result.Content.(type) {
	case domain.IntroContent:
		return dto.OnboardingResponse{
			AccessToken:     result.AccessToken,
			RefreshToken:    result.RefreshToken,
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
			Content:         MapIntroContentToDto(v),
		}
	case domain.LocationContent:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
			Content:         MapLocationContentToDto(v),
		}
	case domain.LifestyleContent:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
			Content:         MapLifestyleContentToDto(v),
		}
	case domain.BeliefsContent:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
			Content:         MapBeliefsContentToDto(v),
		}
	case domain.WorkAndEducationContent:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
			Content:         MapWorkAndEducationContentToDto(v),
		}
	case domain.LanguagesContent:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
			Content:         MapLanguagesContentToDto(v),
		}
	case domain.PhotosContent:
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
			Content:         MapPhotosContentToDto(v),
		}
	case nil: // for background,prompts and basics steps that don't populate content
		return dto.OnboardingResponse{
			OnboardingSteps: mapOnboardingStepsToDto(result.OnboardingSteps),
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

func MapPhotosContentToDto(content domain.PhotosContent) dto.PhotosContent {
	var urls []dto.UploadUrl
	for _, u := range content.VoicePromptsUploadUrls {
		urls = append(urls, dto.UploadUrl{
			Key:       u.Key,
			UploadUrl: u.UploadUrl,
			Headers:   u.Headers,
			MaxBytes:  u.MaxBytes,
		})
	}

	var prompts []dto.Prompt
	for _, prompt := range content.Prompts {
		prompts = append(prompts, dto.Prompt{
			ID:       prompt.ID,
			Label:    prompt.Label,
			Key:      prompt.Key,
			Category: prompt.Category,
		})
	}

	return dto.PhotosContent{
		VoicePromptsUploadUrls: urls,
		Prompts:                prompts,
	}
}

func MapLanguagesContentToDto(content domain.LanguagesContent) dto.LanguagesContent {
	var urls []dto.UploadUrl
	for _, u := range content.PhotoUploadUrls {
		urls = append(urls, dto.UploadUrl{
			Key:       u.Key,
			UploadUrl: u.UploadUrl,
			Headers:   u.Headers,
			MaxBytes:  u.MaxBytes,
		})
	}

	return dto.LanguagesContent{
		PhotoUploadUrls: urls,
	}
}

func MapWorkAndEducationContentToDto(content domain.WorkAndEducationContent) dto.WorkAndEducationContent {
	var languages []dto.Language
	for _, lang := range content.Languages {
		languages = append(languages, dto.Language{
			ID:    lang.ID,
			Label: lang.Label,
		})
	}

	return dto.WorkAndEducationContent{
		Languages: languages,
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

func MapIntroContentToDto(content domain.IntroContent) dto.IntroContent {
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

	return dto.IntroContent{
		DatingIntentions: datingIntentions,
		Genders:          genders,
	}
}
