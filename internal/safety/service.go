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

func (s *service) BlockUser(ctx context.Context, req safetydomain.BlockRequest) error {
	if req.BlockerID == "" || req.BlockedID == "" {
		return ErrInvalidBlockRequest
	}

	if req.BlockerID == req.BlockedID {
		return ErrSelfBlock
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	blockEntity := safetymapper.BlockRequestToEntity(req)

	err = s.repo.CreateBlock(ctx, blockEntity, tx.Raw())
	if err != nil {
		return fmt.Errorf("create block: %w", err)
	}

	// Update match status to blocked
	if err := s.conversationRepo.SetMatchStatus(ctx, tx.Raw(), req.BlockerID, req.BlockedID, string(conversationdomain.MatchStatusBlocked)); err != nil {
		return fmt.Errorf("set match status: %w", err)
	}

	// Archive conversation if present
	convoID, err := s.conversationRepo.ArchiveConversationBetween(ctx, tx.Raw(), req.BlockerID, req.BlockedID)
	if err != nil {
		return fmt.Errorf("archive conversation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	s.broadcastBlockEvent(req.BlockerID, req.BlockedID, convoID)

	return nil
}

func (s *service) IsBlocked(ctx context.Context, userID, otherUserID string) (bool, error) {
	return s.repo.IsBlockedPair(ctx, userID, otherUserID)
}

func (s *service) GetBlockedUserIDs(ctx context.Context, userID string) ([]string, error) {
	blocks, err := s.repo.ListBlocksForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list blocks for user: %w", err)
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
		return "", fmt.Errorf("map report request: %w", err)
	}

	if err := s.repo.CreateReport(ctx, reportEntity, nil); err != nil {
		return "", fmt.Errorf("create report: %w", err)
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
		return nil, fmt.Errorf("list reports: %w", err)
	}

	return safetymapper.ReportEntitiesToDomain(reports)
}

func (s *service) GetReport(ctx context.Context, reportID string) (*safetydomain.Report, error) {
	reportEntity, err := s.repo.GetReportByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("get report by id: %w", err)
	}

	if reportEntity == nil {
		return nil, nil
	}

	reportDomain, err := safetymapper.ReportEntityToDomain(reportEntity)
	if err != nil {
		return nil, fmt.Errorf("map report entity to domain: %w", err)
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
		return fmt.Errorf("get report by id: %w", err)
	}

	if reportEntity == nil {
		return ErrReportNotFound
	}

	actionEntity, err := safetymapper.ResolveRequestToActionEntity(req)
	if err != nil {
		return fmt.Errorf("map resolve request: %w", err)
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	if err := s.repo.InsertReportAction(ctx, actionEntity, tx.Raw()); err != nil {
		return fmt.Errorf("insert report action: %w", err)
	}

	safetymapper.ApplyResolutionToReportEntity(reportEntity, req)

	if err := s.repo.UpdateReport(ctx, reportEntity, tx.Raw()); err != nil {
		return fmt.Errorf("update report: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
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
