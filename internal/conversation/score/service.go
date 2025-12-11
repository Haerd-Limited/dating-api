package score

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/conversation/score/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	Apply(ctx context.Context, conversationID, userID string, c domain.Contribution) (domain.ScoreSnapshot, error)
	GetSnapshot(ctx context.Context, conversationID, userID string) (domain.ScoreSnapshot, error)
	ApplyViaTx(ctx context.Context, tx *sql.Tx, convoID, userID string, c domain.Contribution) (domain.ScoreSnapshot, error)
	GetScores(ctx context.Context, convoID, userID string, tx *sql.Tx) (me int, them int, shared int, err error)
}

type service struct {
	logger           *zap.Logger
	conversationRepo storage.ConversationRepository
	uow              uow.UoW
}

func NewScoreService(
	logger *zap.Logger,
	conversationRepo storage.ConversationRepository,
	uow uow.UoW,
) Service {
	return &service{
		logger:           logger,
		conversationRepo: conversationRepo,
		uow:              uow,
	}
}

func (s *service) ApplyViaTx(ctx context.Context, tx *sql.Tx, convoID, userID string, c domain.Contribution) (domain.ScoreSnapshot, error) {
	cfg, err := s.getScoreConfig(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "get score config", err)
	}

	// 1) compute raw points
	var pts float64

	switch c.Type {
	case domain.ContribText:
		p := cfg.TextBase + cfg.TextPerChar*float64(c.TextLen)
		if p > cfg.TextMax {
			p = cfg.TextMax
		}

		pts = p
	case domain.ContribVoice:
		if c.Seconds >= cfg.VoiceMinSec {
			p := cfg.VoicePerSec * float64(c.Seconds)
			if p > cfg.VoiceMax {
				p = cfg.VoiceMax
			}

			pts = p
		}
	case domain.ContribCall:
		if c.Seconds >= cfg.CallMinSec {
			mins := float64(c.Seconds) / 60.0

			p := cfg.CallPerMin * mins
			if p > cfg.CallMax {
				p = cfg.CallMax
			}

			pts = p
		}
	}

	// 2) light anti-spam (optional now): cooldown / dup / rate — can add later

	earned := int(math.Floor(pts))
	if earned < 0 {
		earned = 0
	}

	// 3) persist my score
	err = s.conversationRepo.UpdateUserConversationScore(ctx, tx, convoID, userID, earned)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "update user conversation score", err, zap.String("userID", userID), zap.String("convoID", convoID), zap.Int("earned", earned))
	}

	// 4) read both scores
	me, them, shared, err := s.GetScores(ctx, convoID, userID, tx)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "get scores", err, zap.String("userID", userID), zap.String("convoID", convoID))
	}

	canReveal := shared >= cfg.Threshold

	// The reveal happens via the handshake endpoint when both confirm.

	// Extend your snapshot to include shared/min & canReveal.
	// If you can’t change the struct yet, repurpose 'Revealed' to carry 'canReveal' temporarily.
	return domain.ScoreSnapshot{
		Threshold: cfg.Threshold,
		Me:        me,
		Them:      them,
		// Add these fields to your struct:
		Shared:    shared,    // <-- int points (min of both)
		CanReveal: canReveal, // <-- bool: UI shows Reveal button
		// Keep this false; actual reveal flips later via /reveal
		Revealed: false,
	}, nil
}

func (s *service) Apply(ctx context.Context, convoID, userID string, c domain.Contribution) (domain.ScoreSnapshot, error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "begin tx", err)
	}

	defer func() { _ = tx.Rollback() }()

	cfg, err := s.getScoreConfig(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "get score config", err)
	}

	// 1) compute raw points
	var pts float64

	switch c.Type {
	case domain.ContribText:
		p := cfg.TextBase + cfg.TextPerChar*float64(c.TextLen)
		if p > cfg.TextMax {
			p = cfg.TextMax
		}

		pts = p
		s.logger.Info("text contribution", zap.Any("TextPerChar", cfg.TextPerChar), zap.Any("textMAx", cfg.TextMax), zap.Any("text length", c.TextLen), zap.Float64("pts", pts))
	case domain.ContribVoice:
		if c.Seconds >= cfg.VoiceMinSec {
			p := cfg.VoicePerSec * float64(c.Seconds)
			if p > cfg.VoiceMax {
				p = cfg.VoiceMax
			}

			pts = p
		}
	case domain.ContribCall:
		if c.Seconds >= cfg.CallMinSec {
			mins := float64(c.Seconds) / 60.0

			p := cfg.CallPerMin * mins
			if p > cfg.CallMax {
				p = cfg.CallMax
			}

			pts = p
		}
	}

	// 2) light anti-spam (optional now): cooldown / dup / rate — can add later

	earned := int(math.Floor(pts))
	if earned < 0 {
		earned = 0
	}

	// 3) persist my score
	err = s.conversationRepo.UpdateUserConversationScore(ctx, tx.Raw(), convoID, userID, earned)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "update user conversation score", err, zap.String("userID", userID), zap.String("convoID", convoID), zap.Int("earned", earned))
	}

	// 4) read scores
	me, them, shared, err := s.GetScores(ctx, convoID, userID, nil)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "get scores", err, zap.String("userID", userID), zap.String("convoID", convoID))
	}

	canReveal := shared >= cfg.Threshold

	// The reveal happens via the handshake endpoint when both confirm.

	err = tx.Commit()
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "commit tx", err)
	}

	// Extend your snapshot to include shared/min & canReveal.
	// If you can’t change the struct yet, repurpose 'Revealed' to carry 'canReveal' temporarily.
	return domain.ScoreSnapshot{
		Threshold: cfg.Threshold,
		Me:        me,
		Them:      them,
		// Add these fields to your struct:
		Shared:    shared,    // <-- int points (min of both)
		CanReveal: canReveal, // <-- bool: UI shows Reveal button
		// Keep this false; actual reveal flips later via /reveal
		Revealed: false,
	}, nil
}

func (s *service) GetScores(ctx context.Context, convoID, userID string, tx *sql.Tx) (me int, them int, shared int, err error) {
	// 4) read both scores
	me, err = s.conversationRepo.GetUserConversationScore(ctx, userID, convoID, tx)
	if err != nil {
		return 0, 0, 0, commonlogger.LogError(s.logger, "get user conversation score", err, zap.String("userID", userID), zap.String("convoID", convoID))
	}

	them, err = s.conversationRepo.GetOtherParticipantConversationScore(ctx, userID, convoID, tx)
	if err != nil {
		return 0, 0, 0, commonlogger.LogError(s.logger, "get other participant conversation score", err, zap.String("userID", userID), zap.String("convoID", convoID))
	}

	shared = me
	if them < shared {
		shared = them
	}

	return me, them, shared, nil
}

func (s *service) GetSnapshot(ctx context.Context, convoID, userID string) (domain.ScoreSnapshot, error) {
	cfg, err := s.getScoreConfig(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "get score config", err)
	}

	convo, err := s.conversationRepo.GetConversationByID(ctx, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "get conversation by ID", err, zap.String("convoID", convoID))
	}

	me, them, shared, err := s.GetScores(ctx, convoID, userID, nil)
	if err != nil {
		return domain.ScoreSnapshot{}, commonlogger.LogError(s.logger, "get scores", err, zap.String("userID", userID), zap.String("convoID", convoID))
	}

	canReveal := shared >= cfg.Threshold

	var revealed bool
	if convo.VisibilityState == constants.VisibilityStateRevealed {
		revealed = true
	}

	return domain.ScoreSnapshot{Threshold: cfg.Threshold, Me: me, Them: them, Revealed: revealed, Shared: shared, CanReveal: canReveal}, nil
}

func (s *service) getScoreConfig(ctx context.Context) (domain.ScoreCfg, error) {
	settings, err := s.conversationRepo.GetScoreSettings(ctx)
	if err != nil {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "get score settings", err)
	}

	text, err := s.conversationRepo.GetScoringText(ctx)
	if err != nil {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "get scoring text", err)
	}

	bonuses, err := s.conversationRepo.GetScoringBonuses(ctx)
	if err != nil {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "get scoring bonuses", err)
	}

	call, err := s.conversationRepo.GetScoringCall(ctx)
	if err != nil {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "get scoring call", err)
	}

	voice, err := s.conversationRepo.GetScoringVoice(ctx)
	if err != nil {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "get scoring voice", err)
	}

	textBase, ok := text.Base.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert text base to float64", fmt.Errorf("conversion failed"))
	}

	perChar, ok := text.PerChar.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert text per char to float64", fmt.Errorf("conversion failed"))
	}

	textMax, ok := text.MaxPerMessage.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert text max to float64", fmt.Errorf("conversion failed"))
	}

	voicePerSec, ok := voice.PerSecond.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert voice per second to float64", fmt.Errorf("conversion failed"))
	}

	voiceMax, ok := voice.MaxPerNote.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert voice max to float64", fmt.Errorf("conversion failed"))
	}

	callPerMin, ok := call.PerMinute.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert call per minute to float64", fmt.Errorf("conversion failed"))
	}

	callMax, ok := call.MaxPerCall.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert call max to float64", fmt.Errorf("conversion failed"))
	}

	firstMsgOfDay, ok := bonuses.FirstMessageOfDay.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert first message of day to float64", fmt.Errorf("conversion failed"))
	}

	replyBonus, ok := bonuses.ReplyBonus.Float64()
	if !ok {
		return domain.ScoreCfg{}, commonlogger.LogError(s.logger, "convert reply bonus to float64", fmt.Errorf("conversion failed"))
	}

	return domain.ScoreCfg{
		Threshold:       settings.Threshold,
		TextBase:        textBase,
		TextPerChar:     perChar,
		TextMax:         textMax,
		TextCooldownSec: text.CooldownSeconds,
		VoicePerSec:     voicePerSec,
		VoiceMax:        voiceMax,
		VoiceMinSec:     voice.MinDurationSeconds,
		CallPerMin:      callPerMin,
		CallMax:         callMax,
		CallMinSec:      call.MinDurationSeconds,
		FirstMsgOfDay:   firstMsgOfDay,
		ReplyWithinSec:  bonuses.ReplyWithinSeconds,
		ReplyBonus:      replyBonus,
		DupWindowSec:    0,
		MaxMsgsPerMin:   0,
	}, nil
}
