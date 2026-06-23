package adminsession

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/adminsession/domain"
	adminsessionstorage "github.com/Haerd-Limited/dating-api/internal/adminsession/storage"
)

type Service interface {
	Roster() []string
	CreateSession(ctx context.Context, req domain.CreateSessionRequest) (domain.SessionResult, error)
	ValidateToken(ctx context.Context, token string) (*domain.Session, error)
	TouchSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, token string) error
}

type service struct {
	logger *zap.Logger
	repo   adminsessionstorage.Repository
	roster []string
}

func NewService(logger *zap.Logger, repo adminsessionstorage.Repository, roster []string) Service {
	return &service{
		logger: logger,
		repo:   repo,
		roster: roster,
	}
}

var (
	ErrInvalidDisplayName = errors.New("display name is not in the admin roster")
	ErrSessionNotFound    = errors.New("admin session not found")
	ErrSessionExpired     = errors.New("admin session expired")
)

func (s *service) Roster() []string {
	out := make([]string, len(s.roster))
	copy(out, s.roster)

	return out
}

func (s *service) CreateSession(ctx context.Context, req domain.CreateSessionRequest) (domain.SessionResult, error) {
	name := strings.TrimSpace(req.DisplayName)
	if name == "" || !s.isInRoster(name) {
		return domain.SessionResult{}, ErrInvalidDisplayName
	}

	token, err := generateSessionToken()
	if err != nil {
		return domain.SessionResult{}, fmt.Errorf("generate session token: %w", err)
	}

	now := time.Now().UTC()
	session := domain.Session{
		ID:          uuid.NewString(),
		DisplayName: name,
		TokenHash:   HashToken(token),
		APIKeyFP:    req.APIKeyFP,
		IP:          req.IP,
		CreatedAt:   now,
		LastSeenAt:  now,
		ExpiresAt:   now.Add(domain.SessionTTL),
	}

	if err := s.repo.Insert(ctx, session); err != nil {
		return domain.SessionResult{}, fmt.Errorf("insert session: %w", err)
	}

	return domain.SessionResult{
		SessionToken: token,
		DisplayName:  name,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

func (s *service) ValidateToken(ctx context.Context, token string) (*domain.Session, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrSessionNotFound
	}

	session, err := s.repo.GetByTokenHash(ctx, HashToken(token))
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	if session == nil {
		return nil, ErrSessionNotFound
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

func (s *service) TouchSession(ctx context.Context, sessionID string) error {
	now := time.Now().UTC()
	return s.repo.Touch(ctx, sessionID, now, now.Add(domain.SessionTTL))
}

func (s *service) DeleteSession(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrSessionNotFound
	}

	return s.repo.DeleteByTokenHash(ctx, HashToken(token))
}

func (s *service) isInRoster(name string) bool {
	return slices.Contains(s.roster, name)
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func TokenFingerprint(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:8])
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func ParseRoster(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			out = append(out, name)
		}
	}

	return out
}
