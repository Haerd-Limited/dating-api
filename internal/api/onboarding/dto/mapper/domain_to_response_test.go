package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	onboardingdomain "github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
)

func TestToOnboardingResponseMapsQuestionPacksOverview(t *testing.T) {
	steps := onboardingdomain.OnboardingStepsQuestionPacks.GenerateOnboardingSteps()
	result := ToOnboardingResponse(onboardingdomain.StepResult{
		OnboardingSteps: steps,
		Content: onboardingdomain.QuestionPacksContent{
			QuestionPacks: []onboardingdomain.QuestionPack{
				{
					CategoryKey:                "communication",
					CategoryName:               "Communication",
					NumberOfCompletedQuestions: 4,
					TotalQuestions:             10,
					ProgressPercent:            40,
				},
			},
		},
	})

	assert.Equal(t, "QUESTION_PACKS", result.CurrentStep)
	assert.Equal(t, "VIDEO_VERIFICATION", result.NextStep)
	assert.Equal(t, []string{
		"INTRO",
		"BASICS",
		"LOCATION",
		"LIFESTYLE",
		"BELIEFS",
		"BACKGROUND",
		"WORK_AND_EDUCATION",
		"LANGUAGES",
		"PHOTOS",
		"PROMPTS",
		"QUESTION_PACKS",
		"VIDEO_VERIFICATION",
		"COMPLETE",
	}, result.Steps)
	assert.Equal(t, len(onboardingdomain.OrderedSteps), result.TotalSteps)

	overview, ok := result.Content.(dto.QuestionPacksContent)
	require.True(t, ok)
	require.Len(t, overview.QuestionPacks, 1)
	assert.Equal(t, dto.QuestionPack{
		CategoryKey:                "communication",
		CategoryName:               "Communication",
		NumberOfCompletedQuestions: 4,
		TotalQuestions:             10,
		ProgressPercent:            40,
	}, overview.QuestionPacks[0])
}

func TestToOnboardingResponsePreservesMetadataForUnsupportedContent(t *testing.T) {
	result := ToOnboardingResponse(onboardingdomain.StepResult{
		OnboardingSteps: onboardingdomain.OnboardingStepsQuestionPacks.GenerateOnboardingSteps(),
		Content:         struct{ Unsupported bool }{Unsupported: true},
	})

	assert.Equal(t, "QUESTION_PACKS", result.CurrentStep)
	assert.Equal(t, "VIDEO_VERIFICATION", result.NextStep)
	assert.Len(t, result.Steps, len(onboardingdomain.OrderedSteps))
	assert.Nil(t, result.Content)
}
