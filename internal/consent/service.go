//go:generate mockgen -source=service.go -destination=service_mock.go -package=consent
package consent

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
	consentmapper "github.com/Haerd-Limited/dating-api/internal/consent/mapper"
	consentstorage "github.com/Haerd-Limited/dating-api/internal/consent/storage"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type Service interface {
	Record(ctx context.Context, req domain.RecordRequest) error
	Revoke(ctx context.Context, userID, consentType string) error
	ListForUser(ctx context.Context, userID string) ([]domain.Consent, error)
	GetMissingMandatory(ctx context.Context, userID string) ([]string, error)
}

type service struct {
	logger *zap.Logger
	repo   consentstorage.Repository
	cache  sync.Map
}

type cacheEntry struct {
	missing   []string
	expiresAt time.Time
}

func NewService(logger *zap.Logger, repo consentstorage.Repository) Service {
	return &service{
		logger: logger,
		repo:   repo,
	}
}

var (
	ErrInvalidConsentType = errors.New("invalid consent type")
	ErrConsentVersion     = errors.New("invalid consent version")
)

func (s *service) Record(ctx context.Context, req domain.RecordRequest) error {
	if err := s.validateConsentType(req.Type); err != nil {
		return err
	}

	expectedVersion := constants.CurrentConsentVersion(req.Type)
	if req.Version != expectedVersion {
		return fmt.Errorf("%w: expected %s", ErrConsentVersion, expectedVersion)
	}

	err := s.repo.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("record consent: %w", err)
	}

	s.cache.Delete(req.UserID)

	return nil
}

func (s *service) Revoke(ctx context.Context, userID, consentType string) error {
	if err := s.validateConsentType(consentType); err != nil {
		return err
	}

	err := s.repo.Revoke(ctx, userID, consentType)
	if err != nil {
		return fmt.Errorf("revoke consent: %w", err)
	}

	s.cache.Delete(userID)

	return nil
}

func (s *service) ListForUser(ctx context.Context, userID string) ([]domain.Consent, error) {
	entities, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list consents: %w", err)
	}

	return consentmapper.EntitiesToDomain(entities), nil
}

func (s *service) GetMissingMandatory(ctx context.Context, userID string) ([]string, error) {
	if entry, ok := s.cache.Load(userID); ok {
		cached := entry.(cacheEntry)
		if time.Now().Before(cached.expiresAt) {
			return slices.Clone(cached.missing), nil
		}

		s.cache.Delete(userID)
	}

	versions := map[string]string{
		constants.ConsentTypePrivacyPolicy:  constants.CurrentPrivacyPolicyVersion,
		constants.ConsentTypeTermsOfService: constants.CurrentTermsOfServiceVersion,
	}

	missing, err := s.repo.GetMissingMandatory(ctx, userID, constants.MandatoryConsentTypes, versions)
	if err != nil {
		return nil, fmt.Errorf("get missing mandatory consents: %w", err)
	}

	s.cache.Store(userID, cacheEntry{
		missing:   slices.Clone(missing),
		expiresAt: time.Now().Add(30 * time.Second),
	})

	return missing, nil
}

func (s *service) validateConsentType(consentType string) error {
	if !slices.Contains(constants.MandatoryConsentTypes, consentType) {
		return ErrInvalidConsentType
	}

	return nil
}
