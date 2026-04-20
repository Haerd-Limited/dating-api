package dataexport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	compatibilitystorage "github.com/Haerd-Limited/dating-api/internal/compatibility/storage"
	conversationstorage "github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/dataexport/domain"
	"github.com/Haerd-Limited/dating-api/internal/dataexport/storage"
	"github.com/Haerd-Limited/dating-api/internal/discover"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	feedbackstorage "github.com/Haerd-Limited/dating-api/internal/feedback/storage"
	insightstorage "github.com/Haerd-Limited/dating-api/internal/insights/storage"
	interactionstorage "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	profilestorage "github.com/Haerd-Limited/dating-api/internal/profile/storage"
	safetystorage "github.com/Haerd-Limited/dating-api/internal/safety/storage"
	userstorage "github.com/Haerd-Limited/dating-api/internal/user/storage"
	verificationstorage "github.com/Haerd-Limited/dating-api/internal/verification/storage"
)

const rateLimitWindow = 24 * time.Hour

var ErrExportRateLimited = errors.New("data export can only be requested once every 24 hours")

// Service exports all user data for GDPR Right to Access / Data Portability.
type Service interface {
	ExportUserData(ctx context.Context, userID string) (*domain.ExportPayload, error)
}

type service struct {
	logger            *zap.Logger
	exportRepo        storage.Repository
	userRepo          userstorage.UserRepository
	profileRepo       profilestorage.ProfileRepository
	discoverService   discover.Service
	conversationRepo  conversationstorage.ConversationRepository
	interactionRepo   interactionstorage.InteractionRepository
	feedbackRepo      feedbackstorage.Repository
	insightsRepo      insightstorage.Repository
	verificationRepo  verificationstorage.VerificationRepository
	safetyRepo        safetystorage.Repository
	compatibilityRepo compatibilitystorage.CompatibilityRepository
}

// NewService returns a new data export service.
func NewService(
	logger *zap.Logger,
	exportRepo storage.Repository,
	userRepo userstorage.UserRepository,
	profileRepo profilestorage.ProfileRepository,
	discoverService discover.Service,
	conversationRepo conversationstorage.ConversationRepository,
	interactionRepo interactionstorage.InteractionRepository,
	feedbackRepo feedbackstorage.Repository,
	insightsRepo insightstorage.Repository,
	verificationRepo verificationstorage.VerificationRepository,
	safetyRepo safetystorage.Repository,
	compatibilityRepo compatibilitystorage.CompatibilityRepository,
) Service {
	return &service{
		logger:            logger,
		exportRepo:        exportRepo,
		userRepo:          userRepo,
		profileRepo:       profileRepo,
		discoverService:   discoverService,
		conversationRepo:  conversationRepo,
		interactionRepo:   interactionRepo,
		feedbackRepo:      feedbackRepo,
		insightsRepo:      insightsRepo,
		verificationRepo:  verificationRepo,
		safetyRepo:        safetyRepo,
		compatibilityRepo: compatibilityRepo,
	}
}

// ExportUserData assembles and returns all personal data for the user.
func (s *service) ExportUserData(ctx context.Context, userID string) (*domain.ExportPayload, error) {
	lastAt, err := s.exportRepo.GetLastRequestedAt(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get last export time: %w", err)
	}
	if !lastAt.IsZero() && time.Since(lastAt) < rateLimitWindow {
		return nil, ErrExportRateLimited
	}

	payload, err := s.assembleExport(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.exportRepo.InsertRequest(ctx, userID); err != nil {
		return nil, fmt.Errorf("record export request: %w", err)
	}

	s.logger.Info("Data export completed", zap.String("userID", userID), zap.Time("exportedAt", payload.ExportedAt))
	return payload, nil
}

func (s *service) assembleExport(ctx context.Context, userID string) (*domain.ExportPayload, error) {
	payload := &domain.ExportPayload{
		ExportedAt:           time.Now().UTC(),
		Account:              domain.AccountExport{},
		Profile:              domain.ProfileExport{},
		Preferences:          nil,
		Photos:               nil,
		VoicePrompts:         nil,
		Swipes:               nil,
		Matches:              nil,
		Conversations:        nil,
		Feedback:             nil,
		Events:               nil,
		InsightSnapshots:     nil,
		VerificationAttempts: nil,
		Blocks:               nil,
		Reports:              nil,
		MatchingAnswers:      nil,
	}

	// Account
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	payload.Account = userToAccountExport(user)

	// Profile
	profile, err := s.profileRepo.GetUserProfileByUserID(ctx, userID)
	if err == nil {
		payload.Profile = userProfileToExport(profile)
	}

	// Preferences (discover service)
	prefs, err := s.discoverService.GetUserPreferences(ctx, userID)
	if err == nil && prefs != nil {
		payload.Preferences = prefs
	}

	// Photos
	photos, err := s.profileRepo.GetUserPhotos(ctx, userID)
	if err == nil {
		payload.Photos = photosToExport(photos)
	}

	// Voice prompts (all, including inactive)
	voicePrompts, err := s.profileRepo.GetAllVoicePromptsByUserID(ctx, userID)
	if err == nil {
		payload.VoicePrompts = voicePromptsToExport(voicePrompts)
	}

	// Swipes
	swipes, err := s.interactionRepo.ListSwipesByUserID(ctx, userID)
	if err == nil {
		payload.Swipes = swipesToExport(swipes)
	}

	// Matches
	matches, err := s.conversationRepo.GetMatches(ctx, userID)
	if err == nil {
		payload.Matches = matchesToExport(matches)
	}

	// Conversations with messages
	conversations, err := s.gatherConversationsWithMessages(ctx, userID)
	if err == nil {
		payload.Conversations = conversations
	}

	// Feedback
	feedbackList, err := s.feedbackRepo.ListByUserID(ctx, userID)
	if err == nil {
		payload.Feedback = feedbackToExport(feedbackList)
	}

	// Events
	events, err := s.insightsRepo.ListEventsByUserID(ctx, userID)
	if err == nil {
		payload.Events = eventsToExport(events)
	}

	// Insight snapshots
	snapshots, err := s.insightsRepo.ListInsightSnapshotsByUserID(ctx, userID)
	if err == nil {
		payload.InsightSnapshots = insightSnapshotsToExport(snapshots)
	}

	// Verification attempts
	attempts, err := s.verificationRepo.ListVerificationAttemptsByUserID(ctx, userID)
	if err == nil {
		payload.VerificationAttempts = verificationAttemptsToExport(attempts)
	}

	// Blocks
	blocks, err := s.safetyRepo.ListBlocksForUser(ctx, userID)
	if err == nil {
		payload.Blocks = blocksToExport(blocks)
	}

	// Reports (where user is reporter or reported)
	reports, err := s.gatherUserReports(ctx, userID)
	if err == nil {
		payload.Reports = reportsToExport(reports)
	}

	// Matching answers
	answers, err := s.compatibilityRepo.GetUserAnswers(ctx, userID)
	if err == nil {
		payload.MatchingAnswers = userAnswersToExport(answers)
	}

	return payload, nil
}

func (s *service) gatherConversationsWithMessages(ctx context.Context, userID string) ([]domain.ConversationExport, error) {
	matches, err := s.conversationRepo.GetMatches(ctx, userID)
	if err != nil || len(matches) == 0 {
		return nil, err
	}
	var out []domain.ConversationExport
	for _, m := range matches {
		otherID := m.UserB
		if m.UserB == userID {
			otherID = m.UserA
		}
		convo, err := s.conversationRepo.GetConversationByUserIDs(ctx, userID, otherID)
		if err != nil {
			continue
		}
		msgs, err := s.conversationRepo.GetMessagesByConversationID(ctx, convo.ID, userID)
		if err != nil {
			msgs = nil
		}
		out = append(out, domain.ConversationExport{
			ID:       convo.ID,
			UserA:    convo.UserA,
			UserB:    convo.UserB,
			Messages: messagesToExport(msgs),
		})
	}
	return out, nil
}

func (s *service) gatherUserReports(ctx context.Context, userID string) ([]*entity.UserReport, error) {
	reporterFilter := safetystorage.ReportListFilter{Limit: 1000}
	reporterFilter.ReporterID = &userID
	asReporter, err := s.safetyRepo.ListReports(ctx, reporterFilter)
	if err != nil {
		return nil, err
	}
	reportedFilter := safetystorage.ReportListFilter{Limit: 1000}
	reportedFilter.ReportedID = &userID
	asReported, err := s.safetyRepo.ListReports(ctx, reportedFilter)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var combined []*entity.UserReport
	for _, r := range asReporter {
		if _, ok := seen[r.ID]; !ok {
			seen[r.ID] = struct{}{}
			combined = append(combined, r)
		}
	}
	for _, r := range asReported {
		if _, ok := seen[r.ID]; !ok {
			seen[r.ID] = struct{}{}
			combined = append(combined, r)
		}
	}
	return combined, nil
}

func userToAccountExport(u *entity.User) domain.AccountExport {
	out := domain.AccountExport{
		ID:             u.ID,
		FirstName:      u.FirstName,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		OnboardingStep: u.OnboardingStep,
	}
	if u.Email.Valid {
		out.Email = &u.Email.String
	}
	if u.LastName.Valid {
		out.LastName = &u.LastName.String
	}
	if u.Phone.Valid {
		out.Phone = &u.Phone.String
	}
	if u.HowDidYouHearAboutUs.Valid {
		out.HowDidYouHearAboutUs = &u.HowDidYouHearAboutUs.String
	}
	return out
}

func userProfileToExport(p *entity.UserProfile) domain.ProfileExport {
	out := domain.ProfileExport{
		UserID:      p.UserID,
		DisplayName: p.DisplayName,
		Geo:         p.Geo,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Verified:    p.Verified,
	}
	if p.Birthdate.Valid {
		out.Birthdate = &p.Birthdate.Time
	}
	if p.HeightCM.Valid {
		v := int(p.HeightCM.Int16)
		out.HeightCM = &v
	}
	if p.City.Valid {
		out.City = &p.City.String
	}
	if p.Country.Valid {
		out.Country = &p.Country.String
	}
	if p.GenderID.Valid {
		v := int(p.GenderID.Int16)
		out.GenderID = &v
	}
	if p.DatingIntentionID.Valid {
		v := int(p.DatingIntentionID.Int16)
		out.DatingIntentionID = &v
	}
	if p.ReligionID.Valid {
		v := int(p.ReligionID.Int16)
		out.ReligionID = &v
	}
	if p.EducationLevelID.Valid {
		v := int(p.EducationLevelID.Int16)
		out.EducationLevelID = &v
	}
	if p.PoliticalBeliefID.Valid {
		v := int(p.PoliticalBeliefID.Int16)
		out.PoliticalBeliefID = &v
	}
	if p.Work.Valid {
		out.Work = &p.Work.String
	}
	if p.JobTitle.Valid {
		out.JobTitle = &p.JobTitle.String
	}
	if p.University.Valid {
		out.University = &p.University.String
	}
	if p.Emoji.Valid {
		out.Emoji = &p.Emoji.String
	}
	if p.SexualityID.Valid {
		v := int(p.SexualityID.Int16)
		out.SexualityID = &v
	}
	return out
}

func photosToExport(photos []*entity.Photo) []domain.PhotoExport {
	out := make([]domain.PhotoExport, 0, len(photos))
	for _, p := range photos {
		var pos *int
		if p.Position.Valid {
			v := int(p.Position.Int16)
			pos = &v
		}
		out = append(out, domain.PhotoExport{
			ID:        p.ID,
			URL:       p.URL,
			Position:  pos,
			IsPrimary: p.IsPrimary,
			CreatedAt: p.CreatedAt,
		})
	}
	return out
}

func voicePromptsToExport(vps []*entity.VoicePrompt) []domain.VoicePromptExport {
	out := make([]domain.VoicePromptExport, 0, len(vps))
	for _, v := range vps {
		var pt *int
		if v.PromptType.Valid {
			p := int(v.PromptType.Int16)
			pt = &p
		}
		out = append(out, domain.VoicePromptExport{
			ID:         v.ID,
			PromptType: pt,
			AudioURL:   v.AudioURL,
			IsActive:   v.IsActive,
			CreatedAt:  v.CreatedAt,
		})
	}
	return out
}

func swipesToExport(swipes []*entity.Swipe) []domain.SwipeExport {
	out := make([]domain.SwipeExport, 0, len(swipes))
	for _, s := range swipes {
		var msgType, msg *string
		var promptID *int64
		if s.MessageType.Valid {
			msgType = &s.MessageType.String
		}
		if s.Message.Valid {
			msg = &s.Message.String
		}
		if s.PromptID.Valid {
			promptID = &s.PromptID.Int64
		}
		out = append(out, domain.SwipeExport{
			ID:          s.ID,
			ActorID:     s.ActorID,
			TargetID:    s.TargetID,
			Action:      s.Action,
			MessageType: msgType,
			Message:     msg,
			PromptID:    promptID,
			CreatedAt:   s.CreatedAt,
		})
	}
	return out
}

func matchesToExport(matches []*entity.Match) []domain.MatchExport {
	out := make([]domain.MatchExport, 0, len(matches))
	for _, m := range matches {
		var revealedAt *time.Time
		if m.RevealedAt.Valid {
			revealedAt = &m.RevealedAt.Time
		}
		out = append(out, domain.MatchExport{
			ID:         m.ID,
			UserA:      m.UserA,
			UserB:      m.UserB,
			CreatedAt:  m.CreatedAt,
			RevealedAt: revealedAt,
			Status:     m.Status,
			DateMode:   m.DateMode,
		})
	}
	return out
}

func messagesToExport(msgs []*entity.Message) []domain.MessageExport {
	out := make([]domain.MessageExport, 0, len(msgs))
	for _, m := range msgs {
		var textBody, mediaKey *string
		var editedAt, deletedAt *time.Time
		if m.TextBody.Valid {
			textBody = &m.TextBody.String
		}
		if m.MediaKey.Valid {
			mediaKey = &m.MediaKey.String
		}
		if m.EditedAt.Valid {
			editedAt = &m.EditedAt.Time
		}
		if m.DeletedAt.Valid {
			deletedAt = &m.DeletedAt.Time
		}
		out = append(out, domain.MessageExport{
			ID:        m.ID,
			SenderID:  m.SenderID,
			Type:      m.Type,
			TextBody:  textBody,
			MediaKey:  mediaKey,
			CreatedAt: m.CreatedAt,
			EditedAt:  editedAt,
			DeletedAt: deletedAt,
		})
	}
	return out
}

func feedbackToExport(list []*entity.Feedback) []domain.FeedbackExport {
	out := make([]domain.FeedbackExport, 0, len(list))
	for _, f := range list {
		var title *string
		if f.Title.Valid {
			title = &f.Title.String
		}
		out = append(out, domain.FeedbackExport{
			ID:        f.ID,
			Type:      f.Type,
			Title:     title,
			Text:      f.Text,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		})
	}
	return out
}

func eventsToExport(events []*entity.Event) []domain.EventExport {
	out := make([]domain.EventExport, 0, len(events))
	for _, e := range events {
		var props map[string]interface{}
		if len(e.Props) > 0 {
			_ = json.Unmarshal(e.Props, &props)
		}
		out = append(out, domain.EventExport{
			ID:         e.ID,
			OccurredAt: e.OccurredAt,
			Name:       e.Name,
			Props:      props,
			Version:    e.Version,
		})
	}
	return out
}

func insightSnapshotsToExport(snapshots []*entity.InsightSnapshot) []domain.InsightSnapshotExport {
	out := make([]domain.InsightSnapshotExport, 0, len(snapshots))
	for _, sn := range snapshots {
		var payload map[string]interface{}
		if len(sn.Payload) > 0 {
			_ = json.Unmarshal(sn.Payload, &payload)
		}
		out = append(out, domain.InsightSnapshotExport{
			ID:          sn.ID,
			Key:         sn.Key,
			PeriodStart: sn.PeriodStart,
			PeriodEnd:   sn.PeriodEnd,
			Scope:       sn.Scope,
			Payload:     payload,
			CreatedAt:   sn.CreatedAt,
		})
	}
	return out
}

func verificationAttemptsToExport(attempts []*entity.VerificationAttempt) []domain.VerificationAttemptExport {
	out := make([]domain.VerificationAttemptExport, 0, len(attempts))
	for _, a := range attempts {
		var sid *string
		if a.SessionID.Valid {
			sid = &a.SessionID.String
		}
		out = append(out, domain.VerificationAttemptExport{
			ID:        a.ID,
			Type:      a.Type,
			Status:    a.Status,
			SessionID: sid,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		})
	}
	return out
}

func blocksToExport(blocks entity.UserBlockSlice) []domain.BlockExport {
	out := make([]domain.BlockExport, 0, len(blocks))
	for _, b := range blocks {
		var reason *string
		if b.Reason.Valid {
			reason = &b.Reason.String
		}
		out = append(out, domain.BlockExport{
			BlockerUserID: b.BlockerUserID,
			BlockedUserID: b.BlockedUserID,
			Reason:        reason,
			CreatedAt:     b.CreatedAt,
		})
	}
	return out
}

func reportsToExport(reports []*entity.UserReport) []domain.ReportExport {
	out := make([]domain.ReportExport, 0, len(reports))
	for _, r := range reports {
		out = append(out, domain.ReportExport{
			ID:             r.ID,
			ReporterUserID: r.ReporterUserID,
			ReportedUserID: r.ReportedUserID,
			Category:       r.Category,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
		})
	}
	return out
}

func userAnswersToExport(answers entity.UserAnswerSlice) []domain.UserAnswerExport {
	out := make([]domain.UserAnswerExport, 0, len(answers))
	for _, a := range answers {
		out = append(out, domain.UserAnswerExport{
			UserID:     a.UserID,
			QuestionID: a.QuestionID,
			AnswerID:   a.AnswerID,
			Importance: a.Importance,
			UpdatedAt:  a.UpdatedAt,
		})
	}
	return out
}
