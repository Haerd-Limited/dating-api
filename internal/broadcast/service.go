package broadcast

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/broadcast/domain"
	broadcastmapper "github.com/Haerd-Limited/dating-api/internal/broadcast/mapper"
	broadcaststorage "github.com/Haerd-Limited/dating-api/internal/broadcast/storage"
	"github.com/Haerd-Limited/dating-api/internal/communication"
	onboardingdomain "github.com/Haerd-Limited/dating-api/internal/onboarding/domain"
	"github.com/Haerd-Limited/dating-api/internal/user"
)

const maxMessageLength = 1600

var (
	ErrEmptyMessage   = errors.New("message is required")
	ErrMessageTooLong = errors.New("message exceeds maximum length")
)

type Service interface {
	ListWaitlistUsers(ctx context.Context) ([]domain.WaitlistUser, error)
	SendBroadcast(ctx context.Context, userIDs []string, message string) (domain.BroadcastResult, error)
}

type service struct {
	logger               *zap.Logger
	repo                 broadcaststorage.Repository
	userService          user.Service
	communicationService communication.Service
}

func NewService(
	logger *zap.Logger,
	repo broadcaststorage.Repository,
	userService user.Service,
	communicationService communication.Service,
) Service {
	return &service{
		logger:               logger,
		repo:                 repo,
		userService:          userService,
		communicationService: communicationService,
	}
}

func (s *service) ListWaitlistUsers(ctx context.Context) ([]domain.WaitlistUser, error) {
	steps := []string{
		string(onboardingdomain.OnboardingStepsIntro),
		string(onboardingdomain.OnboardingStepsLocation),
	}

	users, err := s.userService.ListWaitlistUsers(ctx, steps)
	if err != nil {
		return nil, fmt.Errorf("list waitlist users: %w", err)
	}

	contactedUserIDs, err := s.repo.GetContactedUserIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("get contacted user ids: %w", err)
	}

	waitlistUsers := make([]domain.WaitlistUser, 0, len(users))

	for _, u := range users {
		if u.PhoneNumber == "" {
			continue
		}

		waitlistUsers = append(waitlistUsers, domain.WaitlistUser{
			ID:             u.ID,
			FirstName:      u.FirstName,
			Phone:          u.PhoneNumber,
			OnboardingStep: u.OnboardingStep,
			CreatedAt:      u.CreatedAt,
			Contacted:      contactedUserIDs[u.ID],
		})
	}

	return waitlistUsers, nil
}

func (s *service) SendBroadcast(ctx context.Context, userIDs []string, message string) (domain.BroadcastResult, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return domain.BroadcastResult{}, ErrEmptyMessage
	}

	if len(message) > maxMessageLength {
		return domain.BroadcastResult{}, ErrMessageTooLong
	}

	users, err := s.userService.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return domain.BroadcastResult{}, fmt.Errorf("get users by ids: %w", err)
	}

	result := domain.BroadcastResult{
		Total:      len(userIDs),
		Recipients: make([]domain.RecipientResult, 0, len(users)),
	}

	for _, u := range users {
		recipient := domain.RecipientResult{
			UserID: u.ID,
			Phone:  u.PhoneNumber,
		}

		if u.PhoneNumber == "" {
			errMsg := "user has no phone number"
			recipient.Status = domain.BroadcastStatusFailed
			recipient.Error = &errMsg
			result.Failed++
			result.Recipients = append(result.Recipients, recipient)

			s.logBroadcast(ctx, u.ID, "", message, domain.BroadcastStatusFailed, &errMsg)

			continue
		}

		sendErr := s.communicationService.SendSMS(u.PhoneNumber, message)
		if sendErr != nil {
			errMsg := sendErr.Error()
			recipient.Status = domain.BroadcastStatusFailed
			recipient.Error = &errMsg
			result.Failed++

			s.logBroadcast(ctx, u.ID, u.PhoneNumber, message, domain.BroadcastStatusFailed, &errMsg)
		} else {
			recipient.Status = domain.BroadcastStatusSent
			result.Sent++

			s.logBroadcast(ctx, u.ID, u.PhoneNumber, message, domain.BroadcastStatusSent, nil)
		}

		result.Recipients = append(result.Recipients, recipient)
	}

	return result, nil
}

func (s *service) logBroadcast(ctx context.Context, userID, phone, message, status string, errMsg *string) {
	logEntry := domain.BroadcastLog{
		UserID:  userID,
		Phone:   phone,
		Message: message,
		Status:  status,
		Error:   errMsg,
	}

	entityLog := broadcastmapper.DomainToEntity(logEntry)
	if err := s.repo.InsertBroadcastLog(ctx, entityLog); err != nil {
		s.logger.Sugar().Errorw("failed to insert sms broadcast log", "userID", userID, "error", err)
	}
}
