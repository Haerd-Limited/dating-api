package mapper

import (
	"github.com/Haerd-Limited/dating-api/internal/api/lookup/dto"
	"github.com/Haerd-Limited/dating-api/internal/lookup/domain"
)

func MapToGetPromptsResponse(domainPrompts []domain.Prompt) dto.GetPromptsResponse {
	var prompts []dto.Prompt
	for _, prompt := range domainPrompts {
		prompts = append(prompts, dto.Prompt{
			ID:       prompt.ID,
			Label:    prompt.Label,
			Key:      prompt.Key,
			Category: prompt.Category,
		})
	}

	return dto.GetPromptsResponse{
		Prompts: prompts,
	}
}

func MapToGetLanguagesResponse(domainLangs []domain.Language) dto.GetLanguagesResponse {
	var languages []dto.Language
	for _, lang := range domainLangs {
		languages = append(languages, dto.Language{
			ID:    lang.ID,
			Label: lang.Label,
		})
	}

	return dto.GetLanguagesResponse{
		Languages: languages,
	}
}

func MapToGetReligionsResponse(domainReligions []domain.Religion) dto.GetReligionsResponse {
	var religions []dto.Religion
	for _, religion := range domainReligions {
		religions = append(religions, dto.Religion{
			ID:    religion.ID,
			Label: religion.Label,
		})
	}

	return dto.GetReligionsResponse{
		Religions: religions,
	}
}

func MapToGetPoliticalBeliefsResponse(domainPoliticalBeliefs []domain.PoliticalBelief) dto.GetPoliticalBeliefsResponse {
	var politicalBeliefs []dto.PoliticalBelief
	for _, politicalBelief := range domainPoliticalBeliefs {
		politicalBeliefs = append(politicalBeliefs, dto.PoliticalBelief{
			ID:    politicalBelief.ID,
			Label: politicalBelief.Label,
		})
	}

	return dto.GetPoliticalBeliefsResponse{
		PoliticalBeliefs: politicalBeliefs,
	}
}

func MapToGetEthnicitiesResponse(domainEthnicities []domain.Ethnicity) dto.GetEthnicitiesResponse {
	var ethnicities []dto.Ethnicity
	for _, ethnicity := range domainEthnicities {
		ethnicities = append(ethnicities, dto.Ethnicity{
			ID:    ethnicity.ID,
			Label: ethnicity.Label,
		})
	}

	return dto.GetEthnicitiesResponse{
		Ethnicities: ethnicities,
	}
}

func MapToGetGendersResponse(domainGenders []domain.Gender) dto.GetGendersResponse {
	var genders []dto.Gender
	for _, gender := range domainGenders {
		genders = append(genders, dto.Gender{
			ID:    gender.ID,
			Label: gender.Label,
		})
	}

	return dto.GetGendersResponse{
		Genders: genders,
	}
}

func MapToGetDatingIntentionsResponse(domainDatingIntentions []domain.DatingIntention) dto.GetDatingIntentionsResponse {
	var datingIntentions []dto.DatingIntention
	for _, datingIntention := range domainDatingIntentions {
		datingIntentions = append(datingIntentions, dto.DatingIntention{
			ID:    datingIntention.ID,
			Label: datingIntention.Label,
		})
	}

	return dto.GetDatingIntentionsResponse{
		DatingIntentions: datingIntentions,
	}
}

func MapToGetEducationLevelsResponse(domainEducationLevels []domain.EducationLevel) dto.GetEducationLevelsResponse {
	var educationLevels []dto.EducationLevel
	for _, educationLevel := range domainEducationLevels {
		educationLevels = append(educationLevels, dto.EducationLevel{
			ID:    educationLevel.ID,
			Label: educationLevel.Label,
		})
	}

	return dto.GetEducationLevelsResponse{
		EducationLevels: educationLevels,
	}
}

func MapToGetHabitsResponse(domainHabits []domain.Habit) dto.GetHabitsResponse {
	var habits []dto.Habit
	for _, habit := range domainHabits {
		habits = append(habits, dto.Habit{
			ID:    habit.ID,
			Label: habit.Label,
		})
	}

	return dto.GetHabitsResponse{
		Habits: habits,
	}
}
