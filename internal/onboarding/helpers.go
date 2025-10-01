package onboarding

import (
	"context"
	"fmt"

	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
)

func (os *onboardingService) getLanguages(ctx context.Context) ([]domain.Language, error) {
	languageEntities, err := os.lookupRepo.GetLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapLanguagesToDomain(languageEntities), nil
}

func (os *onboardingService) getEducationLevels(ctx context.Context) ([]domain.EducationLevel, error) {
	educationLevelEntities, err := os.lookupRepo.GetEducationLevels(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEducationlevelsToDomain(educationLevelEntities), nil
}

func (os *onboardingService) getEthnicities(ctx context.Context) ([]domain.Ethnicity, error) {
	ethnicityEntities, err := os.lookupRepo.GetEthnicities(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapEthnicityToDomain(ethnicityEntities), nil
}

func (os *onboardingService) getPoliticalBeliefs(ctx context.Context) ([]domain.PoliticalBelief, error) {
	politicalBeliefsEntities, err := os.lookupRepo.GetPoliticalBeliefs(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPoliticalBeliefsToDomain(politicalBeliefsEntities), nil
}

func (os *onboardingService) getReligions(ctx context.Context) ([]domain.Religion, error) {
	religionsEntities, err := os.lookupRepo.GetReligions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapReligionsToDomain(religionsEntities), nil
}

func (os *onboardingService) getHabits(ctx context.Context) ([]domain.Habit, error) {
	habitEntities, err := os.lookupRepo.GetHabits(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapHabitsToDomain(habitEntities), nil
}

func (os *onboardingService) getGenders(ctx context.Context) ([]domain.Gender, error) {
	genderEntities, err := os.lookupRepo.GetGenders(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapGendersToDomain(genderEntities), nil
}

func (os *onboardingService) getPrompts(ctx context.Context) ([]domain.Prompt, error) {
	promptEntities, err := os.lookupRepo.GetPrompts(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapPromptsToDomain(promptEntities), nil
}

func (os *onboardingService) getDatingIntentions(ctx context.Context) ([]domain.DatingIntention, error) {
	datingIntentionsEntities, err := os.lookupRepo.GetDatingIntentions(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapDatingIntentionsToDomain(datingIntentionsEntities), nil
}

func (os *onboardingService) updateUserProfile(ctx context.Context, userProfile *domain.UserProfile) error {
	updatedUserProfileEntity, whiteList, err := mapper.MapProfileToEntityForUpdate(userProfile)
	if err != nil {
		return fmt.Errorf("failed to map user profile to entity: %w", err)
	}

	err = os.profileRepo.UpdateUserProfile(ctx, updatedUserProfileEntity, whiteList)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (os *onboardingService) getUserProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	userProfileEntity, err := os.profileRepo.GetUserProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapper.MapUserProfileToDomain(userProfileEntity), nil
}

// ensureStep makes sure that the step being called is the correct step to complete for the provided user. This prevents user's skipping steps in the onboarding process.
func (os *onboardingService) ensureStep(ctx context.Context, userID string, expected domain.Steps) error {
	u, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if domain.Steps(u.OnboardingStep) != expected {
		return ErrIncorrectStepCalled
	}

	return nil
}

func (os *onboardingService) bumpOnboardingStep(
	ctx context.Context,
	userID string,
	expected domain.Steps, // the step this endpoint owns
) (domain.Steps, error) {
	u, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		return domain.OnboardingStepsUnset, fmt.Errorf("get user: %w", err)
	}

	curr := domain.Steps(u.OnboardingStep)

	// 1) If we’re not exactly at the expected step, do NOT advance.
	//    (Already advanced or out of order => no-op)
	if curr != expected {
		return curr, nil
	}

	next := expected.NextStep()
	u.OnboardingStep = string(next)

	err = os.userService.UpdateUser(ctx, u)
	if err != nil {
		return domain.OnboardingStepsUnset, fmt.Errorf("update step: %w", err)
	}

	return next, nil
}
