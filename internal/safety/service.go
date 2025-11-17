package safety

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	conversationdomain "github.com/Haerd-Limited/dating-api/internal/conversation/domain"
	conversationstorage "github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	realtimehub "github.com/Haerd-Limited/dating-api/internal/realtime"
	safetydomain "github.com/Haerd-Limited/dating-api/internal/safety/domain"
	safetymapper "github.com/Haerd-Limited/dating-api/internal/safety/mapper"
	safetystorage "github.com/Haerd-Limited/dating-api/internal/safety/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	commonanalytics "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/analytics"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	BlockUser(ctx context.Context, req safetydomain.BlockRequest) error
	IsBlocked(ctx context.Context, userID, otherUserID string) (bool, error)
	GetBlockedUserIDs(ctx context.Context, userID string) ([]string, error)

	CreateReport(ctx context.Context, req safetydomain.ReportRequest) (string, error)
	ListReports(ctx context.Context, filter safetydomain.ReportListFilter) ([]safetydomain.Report, error)
	GetReport(ctx context.Context, reportID string) (*safetydomain.Report, error)
	ResolveReport(ctx context.Context, req safetydomain.ResolveReportRequest) error
}

type service struct {
	logger           *zap.Logger
	repo             safetystorage.Repository
	conversationRepo conversationstorage.ConversationRepository
	uow              uow.UoW
	hub              realtimehub.Broadcaster
}

func NewService(
	logger *zap.Logger,
	repo safetystorage.Repository,
	conversationRepo conversationstorage.ConversationRepository,
	uow uow.UoW,
	hub realtimehub.Broadcaster,
) Service {
	return &service{
		logger:           logger,
		repo:             repo,
		conversationRepo: conversationRepo,
		uow:              uow,
		hub:              hub,
	}
}

var (
	ErrInvalidBlockRequest  = errors.New("blocker_id and blocked_id are required")
	ErrSelfBlock            = errors.New("you cannot block yourself")
	ErrInvalidReportRequest = errors.New("reporter_id and reported_user_id are required")
	ErrSelfReport           = errors.New("you cannot report yourself")
	ErrReportNotFound       = errors.New("report not found")
)

func (s *service) validateBlockRequest(req safetydomain.BlockRequest) error {
	if req.BlockerID == "" || req.BlockedID == "" {
		return ErrInvalidBlockRequest
	}

	if req.BlockerID == req.BlockedID {
		return ErrSelfBlock
	}

	return nil
}

func (s *service) BlockUser(ctx context.Context, req safetydomain.BlockRequest) error {
	if err := s.validateBlockRequest(req); err != nil {
		return commonlogger.LogError(s.logger, "validate block request", err, zap.String("blockerID", req.BlockerID), zap.String("blockedID", req.BlockedID))
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return commonlogger.LogError(s.logger, "begin tx", err)
	}

	defer func() { _ = tx.Rollback() }()

	blockEntity := safetymapper.BlockRequestToEntity(req)

	err = s.repo.CreateBlock(ctx, blockEntity, tx.Raw())
	if err != nil {
		return commonlogger.LogError(s.logger, "create block", err, zap.String("blockerID", req.BlockerID), zap.String("blockedID", req.BlockedID))
	}

	// Update match status to blocked
	if err := s.conversationRepo.SetMatchStatus(ctx, tx.Raw(), req.BlockerID, req.BlockedID, string(conversationdomain.MatchStatusBlocked)); err != nil {
		return commonlogger.LogError(s.logger, "set match status", err, zap.String("blockerID", req.BlockerID), zap.String("blockedID", req.BlockedID))
	}

	// Archive conversation if present
	convoID, err := s.conversationRepo.ArchiveConversationBetween(ctx, tx.Raw(), req.BlockerID, req.BlockedID)
	if err != nil {
		return commonlogger.LogError(s.logger, "archive conversation", err, zap.String("blockerID", req.BlockerID), zap.String("blockedID", req.BlockedID))
	}

	if err := tx.Commit(); err != nil {
		return commonlogger.LogError(s.logger, "commit tx", err)
	}

	s.broadcastBlockEvent(req.BlockerID, req.BlockedID, convoID)

	// analytics: user blocked
	commonanalytics.Track(ctx, "safety.user_blocked", &req.BlockerID, nil, map[string]any{
		"target_id": req.BlockedID,
	})

	return nil
}

func (s *service) IsBlocked(ctx context.Context, userID, otherUserID string) (bool, error) {
	return s.repo.IsBlockedPair(ctx, userID, otherUserID)
}

func (s *service) GetBlockedUserIDs(ctx context.Context, userID string) ([]string, error) {
	blocks, err := s.repo.ListBlocksForUser(ctx, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "list blocks for user", err, zap.String("userID", userID))
	}

	if len(blocks) == 0 {
		return []string{}, nil
	}

	unique := make(map[string]struct{})

	for _, block := range blocks {
		if block == nil {
			continue
		}

		other := block.BlockerUserID
		if other == userID {
			other = block.BlockedUserID
		}

		if other == "" {
			continue
		}

		unique[other] = struct{}{}
	}

	results := make([]string, 0, len(unique))
	for id := range unique {
		results = append(results, id)
	}

	return results, nil
}

func (s *service) CreateReport(ctx context.Context, req safetydomain.ReportRequest) (string, error) {
	// If the report is about another user, ensure users are not the same
	if req.ReporterUserID == "" || req.ReportedUserID == "" {
		return "", ErrInvalidReportRequest
	}

	if req.ReporterUserID == req.ReportedUserID {
		return "", ErrSelfReport
	}

	reportEntity, err := safetymapper.ReportRequestToEntity(req)
	if err != nil {
		return "", commonlogger.LogError(s.logger, "map report request", err, zap.String("reporterUserID", req.ReporterUserID), zap.String("reportedUserID", req.ReportedUserID))
	}

	if err := s.repo.CreateReport(ctx, reportEntity, nil); err != nil {
		return "", commonlogger.LogError(s.logger, "create report", err, zap.String("reporterUserID", req.ReporterUserID), zap.String("reportedUserID", req.ReportedUserID))
	}

	return reportEntity.ID, nil
}

func (s *service) ListReports(ctx context.Context, filter safetydomain.ReportListFilter) ([]safetydomain.Report, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	repoFilter := safetystorage.ReportListFilter{
		Statuses:   reportStatusesToStrings(filter.Status),
		Categories: filter.Category,
		ReporterID: filter.Reporter,
		ReportedID: filter.Reported,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	}

	reports, err := s.repo.ListReports(ctx, repoFilter)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "list reports", err)
	}

	return safetymapper.ReportEntitiesToDomain(reports)
}

func (s *service) GetReport(ctx context.Context, reportID string) (*safetydomain.Report, error) {
	reportEntity, err := s.repo.GetReportByID(ctx, reportID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get report by id", err, zap.String("reportID", reportID))
	}

	if reportEntity == nil {
		return nil, nil
	}

	reportDomain, err := safetymapper.ReportEntityToDomain(reportEntity)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "map report entity to domain", err, zap.String("reportID", reportID))
	}

	return &reportDomain, nil
}

func (s *service) ResolveReport(ctx context.Context, req safetydomain.ResolveReportRequest) error {
	if req.ReportID == "" {
		return fmt.Errorf("report_id is required")
	}

	if req.ActionType == "" {
		return fmt.Errorf("action_type is required")
	}

	if req.NewStatus == "" {
		return fmt.Errorf("new_status is required")
	}

	reportEntity, err := s.repo.GetReportByID(ctx, req.ReportID)
	if err != nil {
		return commonlogger.LogError(s.logger, "get report by id", err, zap.String("reportID", req.ReportID))
	}

	if reportEntity == nil {
		return ErrReportNotFound
	}

	actionEntity, err := safetymapper.ResolveRequestToActionEntity(req)
	if err != nil {
		return commonlogger.LogError(s.logger, "map resolve request", err, zap.String("reportID", req.ReportID))
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return commonlogger.LogError(s.logger, "begin tx", err)
	}

	defer func() { _ = tx.Rollback() }()

	if err := s.repo.InsertReportAction(ctx, actionEntity, tx.Raw()); err != nil {
		return commonlogger.LogError(s.logger, "insert report action", err, zap.String("reportID", req.ReportID))
	}

	safetymapper.ApplyResolutionToReportEntity(reportEntity, req)

	if err := s.repo.UpdateReport(ctx, reportEntity, tx.Raw()); err != nil {
		return commonlogger.LogError(s.logger, "update report", err, zap.String("reportID", req.ReportID))
	}

	if err := tx.Commit(); err != nil {
		return commonlogger.LogError(s.logger, "commit tx", err)
	}

	return nil
}

func (s *service) broadcastBlockEvent(blockerID, blockedID string, convoID *string) {
	if s.hub == nil {
		return
	}

	payload := map[string]any{
		"blocked_user_id": blockedID,
	}

	if convoID != nil {
		payload["conversation_id"] = *convoID
	}

	evt := dto.Event{
		ID:        realtimehub.NewEventID(),
		Type:      "match.blocked",
		ActorID:   blockerID,
		Ts:        time.Now(),
		ContextID: "",
		Data:      payload,
		Version:   1,
	}

	b, err := json.Marshal(evt)
	if err != nil {
		s.logger.Sugar().Warnw("failed to marshal block event", "error", err)
		return
	}

	s.hub.BroadcastToUser(blockerID, b)
	s.hub.BroadcastToUser(blockedID, b)
}

func reportStatusesToStrings(statuses []safetydomain.ReportStatus) []string {
	if len(statuses) == 0 {
		return nil
	}

	out := make([]string, len(statuses))
	for i, status := range statuses {
		out[i] = string(status)
	}

	return out
}
