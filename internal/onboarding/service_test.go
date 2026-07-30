package onboarding

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/compatibility"
	compatibilitydomain "github.com/Haerd-Limited/dating-api/internal/compatibility/domain"
	"github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/user"
	userdomain "github.com/Haerd-Limited/dating-api/internal/user/domain"
)

const questionPacksTestUserID = "user-question-packs-test"

func newQuestionPacksOnboardingService(t *testing.T, userSvc user.Service, compatSvc compatibility.Service) *onboardingService {
	t.Helper()

	return &onboardingService{
		logger:               zaptest.NewLogger(t),
		userService:          userSvc,
		compatibilityService: compatSvc,
	}
}

func userAtOnboardingStep(step domain.Steps) *userdomain.User {
	return &userdomain.User{
		ID:             questionPacksTestUserID,
		OnboardingStep: string(step),
	}
}

func TestOrderedStepsQuestionPacksPlacement(t *testing.T) {
	idx := func(step domain.Steps) int {
		for i, s := range domain.OrderedSteps {
			if s == step {
				return i
			}
		}

		return -1
	}

	promptsIdx := idx(domain.OnboardingStepsPrompts)
	questionPacksIdx := idx(domain.OnboardingStepsQuestionPacks)
	videoIdx := idx(domain.OnboardingStepsVideoVerification)

	require.NotEqual(t, -1, promptsIdx)
	require.NotEqual(t, -1, questionPacksIdx)
	require.NotEqual(t, -1, videoIdx)
	assert.Less(t, promptsIdx, questionPacksIdx)
	assert.Less(t, questionPacksIdx, videoIdx)
	assert.Equal(t, domain.OnboardingStepsVideoVerification, domain.OnboardingStepsQuestionPacks.NextStep())
}

func TestGetUserCurrentStepMapsQuestionPacksOverview(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	userSvc := user.NewMockService(ctrl)
	compatSvc := compatibility.NewMockService(ctrl)

	userSvc.EXPECT().
		GetUser(ctx, questionPacksTestUserID).
		Return(userAtOnboardingStep(domain.OnboardingStepsQuestionPacks), nil)
	compatSvc.EXPECT().
		GetOverview(ctx, questionPacksTestUserID).
		Return(compatibilitydomain.Overview{
			QuestionPacks: []compatibilitydomain.Pack{
				{
					CategoryKey:                "communication",
					CategoryName:               "Communication",
					NumberOfCompletedQuestions: 4,
					TotalQuestions:             10,
					ProgressPercent:            40,
				},
			},
		}, nil)

	svc := newQuestionPacksOnboardingService(t, userSvc, compatSvc)
	result, err := svc.GetUserCurrentStep(ctx, questionPacksTestUserID)

	require.NoError(t, err)
	assert.Equal(t, domain.OnboardingStepsQuestionPacks, result.CurrentStep)
	assert.Equal(t, domain.QuestionPacksContent{
		QuestionPacks: []domain.QuestionPack{
			{
				CategoryKey:                "communication",
				CategoryName:               "Communication",
				NumberOfCompletedQuestions: 4,
				TotalQuestions:             10,
				ProgressPercent:            40,
			},
		},
	}, result.Content)
}

func TestQuestionPacks(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name      string
		setupMock func(userSvc *user.MockService, compatSvc *compatibility.MockService)
		wantErr   error
		wantNext  domain.Steps
	}{
		{
			name: "wrong step returns ErrIncorrectStepCalled",
			setupMock: func(userSvc *user.MockService, compatSvc *compatibility.MockService) {
				userSvc.EXPECT().
					GetUser(ctx, questionPacksTestUserID).
					Return(userAtOnboardingStep(domain.OnboardingStepsPrompts), nil)
			},
			wantErr: ErrIncorrectStepCalled,
		},
		{
			name: "incomplete question packs returns ErrQuestionPacksIncomplete",
			setupMock: func(userSvc *user.MockService, compatSvc *compatibility.MockService) {
				userSvc.EXPECT().
					GetUser(ctx, questionPacksTestUserID).
					Return(userAtOnboardingStep(domain.OnboardingStepsQuestionPacks), nil)
				compatSvc.EXPECT().
					IsQuestionPacksComplete(ctx, questionPacksTestUserID).
					Return(false, nil)
			},
			wantErr: ErrQuestionPacksIncomplete,
		},
		{
			name: "complete advances to video verification",
			setupMock: func(userSvc *user.MockService, compatSvc *compatibility.MockService) {
				gomock.InOrder(
					userSvc.EXPECT().
						GetUser(ctx, questionPacksTestUserID).
						Return(userAtOnboardingStep(domain.OnboardingStepsQuestionPacks), nil),
					compatSvc.EXPECT().
						IsQuestionPacksComplete(ctx, questionPacksTestUserID).
						Return(true, nil),
					userSvc.EXPECT().
						GetUser(ctx, questionPacksTestUserID).
						Return(userAtOnboardingStep(domain.OnboardingStepsQuestionPacks), nil),
					userSvc.EXPECT().
						UpdateUser(ctx, gomock.Any()).
						DoAndReturn(func(_ context.Context, u *userdomain.User) error {
							assert.Equal(t, string(domain.OnboardingStepsVideoVerification), u.OnboardingStep)
							return nil
						}),
				)
			},
			wantNext: domain.OnboardingStepsVideoVerification,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userSvc := user.NewMockService(ctrl)
			compatSvc := compatibility.NewMockService(ctrl)
			tc.setupMock(userSvc, compatSvc)

			svc := newQuestionPacksOnboardingService(t, userSvc, compatSvc)

			result, err := svc.QuestionPacks(ctx, domain.QuestionPacks{UserID: questionPacksTestUserID})

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr), "expected %v, got %v", tc.wantErr, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantNext, result.CurrentStep)
			assert.Equal(t, domain.OnboardingStepsQuestionPacks, result.PreviousStep)
		})
	}
}
