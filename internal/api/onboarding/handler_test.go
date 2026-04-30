package onboarding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/Haerd-Limited/dating-api/internal/api/onboarding/dto"
	"github.com/Haerd-Limited/dating-api/internal/onboarding"
	onboardingdomain "github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	commoncontext "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/context"
)

const testUserID = "user-handler-test"

func voicePromptDTO(promptType int16, position int16) dto.VoicePrompt {
	return dto.VoicePrompt{
		URL:          "https://example.com/audio.m4a",
		PromptType:   promptType,
		Position:     position,
		WaveformData: []float32{0.1, 0.2, 0.3},
	}
}

func validFiveCorePromptsRequest() dto.PromptsRequest {
	return dto.PromptsRequest{
		UploadedPrompts: []dto.VoicePrompt{
			voicePromptDTO(1, 1),
			voicePromptDTO(2, 2),
			voicePromptDTO(3, 3),
			voicePromptDTO(4, 4),
			voicePromptDTO(5, 5),
		},
	}
}

func TestPromptsHandler(t *testing.T) {
	mockLog := zaptest.NewLogger(t)

	cases := []struct {
		name         string
		input        dto.PromptsRequest
		setupMock    func(svc *onboarding.MockService)
		wantStatus   int
		wantContains string
	}{
		{
			name: "missing-core returns 400 with friendly message",
			input: dto.PromptsRequest{
				UploadedPrompts: []dto.VoicePrompt{
					voicePromptDTO(1, 1),
					voicePromptDTO(2, 2),
					voicePromptDTO(3, 3),
					voicePromptDTO(4, 4),
					voicePromptDTO(99, 5),
				},
			},
			setupMock: func(svc *onboarding.MockService) {
				svc.EXPECT().
					Prompts(gomock.Any(), gomock.Any()).
					Return(onboardingdomain.StepResult{}, fmt.Errorf("validate prompts: %w", profile.ErrMissingRequiredCorePrompts))
			},
			wantStatus:   http.StatusBadRequest,
			wantContains: "All Core prompts must be answered.",
		},
		{
			name:  "valid 5-prompt payload returns 200",
			input: validFiveCorePromptsRequest(),
			setupMock: func(svc *onboarding.MockService) {
				svc.EXPECT().
					Prompts(gomock.Any(), gomock.Any()).
					Return(onboardingdomain.StepResult{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "DTO validation: 4 prompts returns 400 before service is called",
			input: dto.PromptsRequest{
				UploadedPrompts: []dto.VoicePrompt{
					voicePromptDTO(1, 1),
					voicePromptDTO(2, 2),
					voicePromptDTO(3, 3),
					voicePromptDTO(4, 4),
				},
			},
			setupMock:    func(svc *onboarding.MockService) {}, // no service call expected
			wantStatus:   http.StatusBadRequest,
			wantContains: "at least 5 prompts",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := onboarding.NewMockService(ctrl)
			tc.setupMock(mockService)

			body, err := json.Marshal(tc.input)
			require.NoError(t, err)

			ctx := context.WithValue(context.Background(), commoncontext.UserIDKey, testUserID)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/onboarding/prompts", bytes.NewReader(body))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			h := NewOnboardingHandler(mockLog, mockService)
			h.Prompts().ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantStatus, recorder.Code, "body=%s", recorder.Body.String())

			if tc.wantContains != "" {
				assert.Contains(t, recorder.Body.String(), tc.wantContains)
			}
		})
	}
}
