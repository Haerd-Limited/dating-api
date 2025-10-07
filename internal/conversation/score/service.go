package score

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Haerd-Limited/dating-api/internal/conversation/score/domain"
	"github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"go.uber.org/zap"
	"math"
)

type Service interface {
	Apply(ctx context.Context, conversationID, userID string, c domain.Contribution) (domain.ScoreSnapshot, error)
	GetSnapshot(ctx context.Context, conversationID, userID string) (domain.ScoreSnapshot, error)
	ApplyViaTx(ctx context.Context, tx *sql.Tx, convoID, userID string, c domain.Contribution) (domain.ScoreSnapshot, error)
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
		return domain.ScoreSnapshot{}, fmt.Errorf("get score config: %w", err)
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
		return domain.ScoreSnapshot{}, fmt.Errorf("update user conversation score userID=%s convoID=%s earned=%v: %w", userID, convoID, earned, err)
	}

	// 4) read both scores
	var me, them int
	me, err = s.conversationRepo.GetUserConversationScore(ctx, userID, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get user conversation score userID=%s convoID=%s: %w", userID, convoID, err)
	}
	them, err = s.conversationRepo.GetOtherParticipantConversationScore(ctx, userID, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get other participant conversation score userID=%s convoID=%s: %w", userID, convoID, err)
	}

	shared := me
	if them < shared {
		shared = them
	}
	canReveal := shared >= cfg.Threshold

	// The reveal happens via the handshake endpoint when both confirm.

	if err := tx.Commit(); err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("commit tx: %w", err)
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

func (s *service) Apply(ctx context.Context, convoID, userID string, c domain.Contribution) (domain.ScoreSnapshot, error) {
	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback() }()
	cfg, err := s.getScoreConfig(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get score config: %w", err)
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
		return domain.ScoreSnapshot{}, fmt.Errorf("update user conversation score userID=%s convoID=%s earned=%v: %w", userID, convoID, earned, err)
	}

	// 4) read both scores
	var me, them int
	me, err = s.conversationRepo.GetUserConversationScore(ctx, userID, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get user conversation score userID=%s convoID=%s: %w", userID, convoID, err)
	}
	them, err = s.conversationRepo.GetOtherParticipantConversationScore(ctx, userID, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get other participant conversation score userID=%s convoID=%s: %w", userID, convoID, err)
	}

	shared := me
	if them < shared {
		shared = them
	}
	canReveal := shared >= cfg.Threshold

	// The reveal happens via the handshake endpoint when both confirm.

	if err := tx.Commit(); err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("commit tx: %w", err)
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

func (s *service) GetSnapshot(ctx context.Context, convoID, userID string) (domain.ScoreSnapshot, error) {
	cfg, err := s.getScoreConfig(ctx)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get score config: %w", err)
	}

	convo, err := s.conversationRepo.GetConversationByID(ctx, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get conversation by ID convoID=%s: %w", convoID, err)
	}

	var me, them int
	me, err = s.conversationRepo.GetUserConversationScore(ctx, userID, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get user conversation score userID=%s convoID=%s: %w", userID, convoID, err)
	}
	them, err = s.conversationRepo.GetOtherParticipantConversationScore(ctx, userID, convoID)
	if err != nil {
		return domain.ScoreSnapshot{}, fmt.Errorf("get other participant conversation score userID=%s convoID=%s: %w", userID, convoID, err)
	}

	var revealed bool
	if convo.VisibilityState == constants.VisibilityStateVisible {
		revealed = true
	}

	return domain.ScoreSnapshot{Threshold: cfg.Threshold, Me: me, Them: them, Revealed: revealed}, nil
}

func (s *service) getScoreConfig(ctx context.Context) (domain.ScoreCfg, error) {
	settings, err := s.conversationRepo.GetScoreSettings(ctx)
	if err != nil {
		return domain.ScoreCfg{}, fmt.Errorf("get score settings: %w", err)
	}
	text, err := s.conversationRepo.GetScoringText(ctx)
	if err != nil {
		return domain.ScoreCfg{}, fmt.Errorf("get scoring text: %w", err)
	}
	bonuses, err := s.conversationRepo.GetScoringBonuses(ctx)
	if err != nil {
		return domain.ScoreCfg{}, fmt.Errorf("get scoring bonuses: %w", err)
	}
	call, err := s.conversationRepo.GetScoringCall(ctx)
	if err != nil {
		return domain.ScoreCfg{}, fmt.Errorf("get scoring call: %w", err)
	}
	voice, err := s.conversationRepo.GetScoringVoice(ctx)
	if err != nil {
		return domain.ScoreCfg{}, fmt.Errorf("get scoring voice: %w", err)
	}

	textBase, ok := text.Base.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert text base to float64: %w", err)
	}

	perChar, ok := text.PerChar.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert text per char to float64: %w", err)
	}

	textMax, ok := text.MaxPerMessage.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert text max to float64: %w", err)
	}
	voicePerSec, ok := voice.PerSecond.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert voice per second to float64: %w", err)
	}

	voiceMax, ok := voice.MaxPerNote.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert voice max to float64: %w", err)
	}

	callPerMin, ok := call.PerMinute.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert call per minute to float64: %w", err)
	}

	callMax, ok := call.MaxPerCall.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert call max to float64: %w", err)
	}

	firstMsgOfDay, ok := bonuses.FirstMessageOfDay.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert first message of day to float64: %w", err)
	}

	replyBonus, ok := bonuses.ReplyBonus.Float64()
	if !ok {
		return domain.ScoreCfg{}, fmt.Errorf("convert reply bonus to float64: %w", err)
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
