package onboarding

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
	commonErrors "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/errors"
)

func (os *onboardingService) validateBeliefDetails(beliefDetails domain.Beliefs) error {
	if beliefDetails.PoliticalBeliefsID == 0 {
		return fmt.Errorf("invalid political belief id: %w", ErrInvalidID)
	}

	if beliefDetails.ReligionID == 0 {
		return fmt.Errorf("invalid religion id: %w", ErrInvalidID)
	}

	return nil
}

func (os *onboardingService) validateBackgroundDetails(backgroundDetails domain.Background) error {
	if backgroundDetails.EducationLevelID == 0 {
		return fmt.Errorf("invalid education level id: %w", ErrInvalidID)
	}

	if backgroundDetails.EthnicityID == 0 {
		return fmt.Errorf("invalid ethnicity id: %w", ErrInvalidID)
	}

	return nil
}

func (os *onboardingService) validateLifestyleDetails(lifestyleDetails domain.Lifestyle) error {
	if lifestyleDetails.DrinkingID == 0 {
		return fmt.Errorf("invaid drinking id: %w", ErrInvalidID)
	}

	if lifestyleDetails.SmokingID == 0 {
		return fmt.Errorf("invaid smoking id: %w", ErrInvalidID)
	}

	if lifestyleDetails.MarijuanaID == 0 {
		return fmt.Errorf("invalid marijuanna id: %w", ErrInvalidID)
	}

	if lifestyleDetails.DrugsID == 0 {
		return fmt.Errorf("invalid drugs id: %w", ErrInvalidID)
	}

	return nil
}

func (os *onboardingService) validateBasicDetails(basicDetails domain.Basics) error {
	if basicDetails.GenderID == 0 {
		return fmt.Errorf("invaid gender id: %w", ErrInvalidID)
	}

	if basicDetails.DatingIntentionID == 0 {
		return fmt.Errorf("invaid dating intention id: %w", ErrInvalidID)
	}

	return nil
}

func (os *onboardingService) validateAndSanitiseIntroDetails(intro *domain.Intro) error {
	intro.FirstName = strings.TrimSpace(intro.FirstName)
	if hasAnySpace(intro.FirstName) {
		return fmt.Errorf("first%w", ErrNameContainsSpaces)
	}
	// first name length check
	if l := len(intro.FirstName); l < minNameLen || l > maxNameLen {
		return ErrInvalidNameLength
	}

	if intro.LastName != nil {
		temp := strings.TrimSpace(*intro.LastName)
		intro.LastName = &temp

		if hasAnySpace(*intro.LastName) {
			return fmt.Errorf("last%w", ErrNameContainsSpaces)
		}

		// last name length check
		if l := len(*intro.LastName); l < minNameLen || l > maxNameLen {
			return ErrInvalidNameLength
		}
	}

	if !looksLikeEmail(strings.TrimSpace(intro.Email)) {
		return commonErrors.ErrInvalidEmail
	}

	return nil
}

// hasAnySpace returns true if s contains any Unicode whitespace character.
func hasAnySpace(s string) bool {
	return strings.IndexFunc(s, unicode.IsSpace) >= 0
}

func looksLikeEmail(s string) bool { return strings.Contains(s, "@") && strings.Contains(s, ".") }

// todo: call of below through lookup service if logic gets involved
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
