package onboarding

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/mapper"
	commonanalytics "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/analytics"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

// todo: call of below through lookup service if logic gets involved
func (os *onboardingService) getLanguages(ctx context.Context) ([]domain.Language, error) {
	languageEntities, err := os.lookupRepo.GetLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapLanguagesToDomain(languageEntities), nil
}

func (os *onboardingService) validatePrompts(uploadedPrompts domain.Prompts) error {
	if len(uploadedPrompts.UploadedPrompts) == 0 {
		return ErrMissingPrompts
	}

	if len(uploadedPrompts.UploadedPrompts) < MinimumNumberOfPrompts {
		return fmt.Errorf("%w. please provide atleast %v", ErrNotEnoughPromptsProvided, MinimumNumberOfPrompts)
	}

	if len(uploadedPrompts.UploadedPrompts) > constants.MaximumNumberOfPrompts {
		return fmt.Errorf("%w. please provide atmost %v", ErrTooManyPromptsProvided, constants.MaximumNumberOfPrompts)
	}

	return nil
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

func (os *onboardingService) getSexualities(ctx context.Context) ([]domain.Sexuality, error) {
	sexualityEntities, err := os.lookupRepo.GetSexualities(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.MapSexualitiesToDomain(sexualityEntities), nil
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

	// analytics: onboarding step completed
	userIDCopy := userID // capture address
	commonanalytics.Track(ctx, "onboarding.completed_step", &userIDCopy, nil, map[string]any{
		"step": string(expected),
	})

	return next, nil
}

func (os *onboardingService) sendPreregistrationNotification(ctx context.Context, userID string, genderID, sexualityID int16) {
	// Get user details
	userDetails, err := os.userService.GetUser(ctx, userID)
	if err != nil {
		commonlogger.LogError(os.logger, "failed to get user for notification", err, zap.String("userID", userID))
		return
	}

	// Get gender label
	genderEntity, err := os.lookupRepo.GetGenderByID(ctx, genderID)
	if err != nil {
		commonlogger.LogError(os.logger, "failed to get gender for notification", err, zap.String("userID", userID), zap.Int16("genderID", genderID))
		return
	}

	// Get sexuality label
	sexualityEntity, err := os.lookupRepo.GetSexualityByID(ctx, sexualityID)
	if err != nil {
		commonlogger.LogError(os.logger, "failed to get sexuality for notification", err, zap.String("userID", userID), zap.Int16("sexualityID", sexualityID))
		return
	}

	// Build message
	lastName := ""
	if userDetails.LastName != nil {
		lastName = *userDetails.LastName
	}

	message := fmt.Sprintf("New pre-registration completed!\nName: %s %s\nGender: %s\nSexuality: %s",
		userDetails.FirstName, lastName, genderEntity.Label, sexualityEntity.Label)

	// Send SMS notification to all configured phone numbers
	for _, phoneNumber := range os.notificationPhoneNumbers {
		if phoneNumber == "" {
			continue
		}

		err = os.communicationService.SendSMS(phoneNumber, message)
		if err != nil {
			commonlogger.LogError(os.logger, "failed to send preregistration notification SMS", err,
				zap.String("userID", userID), zap.String("phoneNumber", phoneNumber))
			// Continue sending to other numbers even if one fails
		}
	}
}
