package discover

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"github.com/Haerd-Limited/dating-api/internal/discover/storage"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/matching"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

type Service interface {
	GetDiscoverFeed(ctx context.Context, userID string, limit int, offset int) (domain.DiscoverFeedResult, error)
	GetDiscoverFeedWithFilters(ctx context.Context, userID string, limit int, offset int, filters *domain.DiscoverFilters) (domain.DiscoverFeedResult, error)
	GetVoiceWorthHearing(ctx context.Context, userID string) ([]profilecard.ProfileCard, error)
	GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error)
	AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error)
}

type service struct {
	logger          *zap.Logger
	profileService  profile.Service
	matchingService matching.Service
	discoverRepo    storage.DiscoverRepository
}

func NewDiscoverService(
	logger *zap.Logger,
	profileService profile.Service,
	matchingService matching.Service,
	discoverRepo storage.DiscoverRepository,
) Service {
	return &service{
		logger:          logger,
		profileService:  profileService,
		matchingService: matchingService,
		discoverRepo:    discoverRepo,
	}
}

const (
	minOverlap                          = 5
	voiceWorthHearingSelectionLimit     = 3
	voiceWorthHearingPreferencePoolSize = 30
)

func (s *service) AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error) {
	return s.discoverRepo.AlreadyInteracted(ctx, userID, targetUserID)
}

// GetDiscoverFeed returns the discover feed without filters (backwards compatibility)
func (s *service) GetDiscoverFeed(ctx context.Context, userID string, limit int, offset int) (domain.DiscoverFeedResult, error) {
	return s.GetDiscoverFeedWithFilters(ctx, userID, limit, offset, nil)
}

// GetDiscoverFeedWithFilters returns the discover feed with optional filters
func (s *service) GetDiscoverFeedWithFilters(ctx context.Context, userID string, limit int, offset int, filters *domain.DiscoverFilters) (domain.DiscoverFeedResult, error) {
	if offset < 0 {
		offset = 0
	}

	now := time.Now().UTC()

	totalSwipes, gatingSwipeAt, err := s.discoverRepo.GetSwipeUsageStats(ctx, userID, domain.DiscoverQuotaWindow, domain.DiscoverQuotaLimit, now)
	if err != nil {
		return domain.DiscoverFeedResult{}, fmt.Errorf("failed to compute swipe usage userID=%s: %w", userID, err)
	}

	quota := domain.NewQuotaStatus(domain.DiscoverQuotaLimit, domain.DiscoverQuotaWindow, totalSwipes, gatingSwipeAt)

	remaining := quota.SwipesRemaining
	if remaining <= 0 {
		return domain.NewDiscoverFeedResult(nil, quota), nil
	}

	if offset >= remaining {
		return domain.NewDiscoverFeedResult(nil, quota), nil
	}

	if limit <= 0 || limit > remaining-offset {
		limit = remaining - offset
	}

	if filters != nil {
		if preferenceUpdate := domain.NewPreferenceUpdateFromFilters(filters); preferenceUpdate != nil {
			if err := s.discoverRepo.SaveUserDiscoverPreferences(ctx, userID, preferenceUpdate); err != nil {
				return domain.DiscoverFeedResult{}, fmt.Errorf("failed to persist discover preferences userID=%s: %w", userID, err)
			}
		}
	}

	candidates, err := s.discoverRepo.GetDiscoverFeedCandidatesWithFilters(ctx, userID, limit, offset, filters)
	if err != nil {
		return domain.DiscoverFeedResult{}, fmt.Errorf("failed to get candidate IDs userID=%s limit=%v offset=%v: %w", userID, limit, offset, err)
	}

	if len(candidates) == 0 {
		return domain.NewDiscoverFeedResult(nil, quota), nil
	}

	currentUserProfile, err := s.profileService.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return domain.DiscoverFeedResult{}, fmt.Errorf("failed to get current user profile userID=%s: %w", userID, err)
	}

	profiles := make([]profilecard.ProfileCard, 0, len(candidates))

	for _, candidate := range candidates {
		var profileErr error

		p, profileErr := s.profileService.GetProfileCardWithDistance(ctx, candidate.UserID, currentUserProfile.Latitude, currentUserProfile.Longitude)
		if profileErr != nil {
			return domain.DiscoverFeedResult{}, fmt.Errorf("failed to get profile card userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		// Apply post-query filters that can't be done efficiently in SQL
		if filters != nil && !s.passesPostQueryFilters(p, filters, &currentUserProfile) {
			continue
		}

		p.MatchSummary, profileErr = s.computeMatch(ctx, userID, candidate.UserID, minOverlap)
		if profileErr != nil {
			return domain.DiscoverFeedResult{}, fmt.Errorf("failed to compute match userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		profiles = append(profiles, p)
	}

	return domain.NewDiscoverFeedResult(profiles, quota), nil
}

func (s *service) GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error) {
	numberOfOppositeGenderProfiles, err := s.discoverRepo.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get number of complete profiles of opposite gender: %w", err)
	}

	if numberOfOppositeGenderProfiles < constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
		return nil, nil
	}

	profiles, err := s.GetVoiceWorthHearing(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get voice worth hearing userID=%s: %w", userID, err)
	}

	ids := make([]string, 0, len(profiles))
	for _, p := range profiles {
		ids = append(ids, p.UserID)
	}

	return ids, nil
}

// GetVoiceWorthHearing returns the curated "Voices Worth Hearing" cards for the current user.
// Results are cached per user for the calendar week starting on Sunday 00:00 UTC, so the set
// remains stable throughout the week and is recomputed automatically when a new week begins.
func (s *service) GetVoiceWorthHearing(ctx context.Context, userID string) ([]profilecard.ProfileCard, error) {
	weekStart := startOfWeek(time.Now().UTC(), time.Sunday)

	storedPreferences, err := s.discoverRepo.GetUserDiscoverPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get stored preferences userID=%s: %w", userID, err)
	}

	matcher := newPreferenceMatcher(storedPreferences)

	candidateLimit := voiceWorthHearingSelectionLimit
	if matcher != nil {
		candidateLimit = voiceWorthHearingPreferencePoolSize
	}

	cachedIDs, err := s.discoverRepo.GetWeeklyVoiceWorthHearingIDs(ctx, userID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("get cached voice worth hearing ids userID=%s: %w", userID, err)
	}

	var candidates []*entity.UserProfile

	switch {
	case len(cachedIDs) > 0:
		candidates, err = s.discoverRepo.GetVoiceWorthHearingByIDs(ctx, userID, cachedIDs)
		if err != nil {
			return nil, fmt.Errorf("hydrate cached candidates userID=%s: %w", userID, err)
		}

		candidates = orderCandidatesByIDs(candidates, cachedIDs)
	default:
		candidates, err = s.discoverRepo.GetVoiceWorthHearing(ctx, userID, candidateLimit)
		if err != nil {
			return nil, fmt.Errorf("get candidates userID=%s: %w", userID, err)
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	currentUserProfile, err := s.profileService.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user profile userID=%s: %w", userID, err)
	}

	var ethnicityByUser map[string][]int16

	if matcher != nil && matcher.requiresEthnicity() {
		userIDs := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			userIDs = append(userIDs, candidate.UserID)
		}

		ethnicityByUser, err = s.discoverRepo.GetUsersEthnicityIDs(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load candidate ethnicities userID=%s: %w", userID, err)
		}
	}

	profiles := make([]profilecard.ProfileCard, 0, voiceWorthHearingSelectionLimit)
	selectedIDs := make([]string, 0, voiceWorthHearingSelectionLimit)

	for _, candidate := range candidates {
		if len(profiles) == voiceWorthHearingSelectionLimit {
			break
		}

		alreadyInteracted, interactionErr := s.discoverRepo.AlreadyInteracted(ctx, userID, candidate.UserID)
		if interactionErr != nil {
			return nil, fmt.Errorf("already interacted userID=%s profileUserID=%s: %w", userID, candidate.UserID, interactionErr)
		}

		if alreadyInteracted {
			continue
		}

		card, profileErr := s.profileService.GetProfileCardWithDistance(ctx, candidate.UserID, currentUserProfile.Latitude, currentUserProfile.Longitude)
		if profileErr != nil {
			return nil, fmt.Errorf("get profile card userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		var datingIntentionID *int16

		if candidate.DatingIntentionID.Valid {
			value := candidate.DatingIntentionID.Int16
			datingIntentionID = &value
		}

		var religionID *int16

		if candidate.ReligionID.Valid {
			value := candidate.ReligionID.Int16
			religionID = &value
		}

		var candidateEthnicities []int16
		if matcher != nil && matcher.requiresEthnicity() {
			candidateEthnicities = ethnicityByUser[candidate.UserID]
		}

		if matcher != nil && !matcher.matches(card.Age, card.DistanceKm, datingIntentionID, religionID, candidateEthnicities) {
			continue
		}

		likeCount, likeErr := s.discoverRepo.GetLikeAndSuperlikeCount(ctx, candidate.UserID)
		if likeErr != nil {
			return nil, fmt.Errorf("get like and superlike count userID=%s profileUserID=%s: %w", userID, candidate.UserID, likeErr)
		}

		card.LikeCount = &likeCount

		card.MatchSummary, profileErr = s.computeMatch(ctx, userID, candidate.UserID, minOverlap)
		if profileErr != nil {
			return nil, fmt.Errorf("failed to compute match userID=%s profileUserID=%s: %w", userID, candidate.UserID, profileErr)
		}

		profiles = append(profiles, card)
		selectedIDs = append(selectedIDs, candidate.UserID)
	}

	if len(cachedIDs) == 0 && len(selectedIDs) > 0 {
		if err := s.discoverRepo.SaveWeeklyVoiceWorthHearingIDs(ctx, userID, weekStart, selectedIDs); err != nil {
			return nil, fmt.Errorf("persist vwh cache userID=%s: %w", userID, err)
		}
	}

	return profiles, nil
}

func orderCandidatesByIDs(candidates []*entity.UserProfile, ordering []string) []*entity.UserProfile {
	if len(ordering) == 0 || len(candidates) <= 1 {
		return candidates
	}

	index := make(map[string]int, len(ordering))
	for pos, id := range ordering {
		index[id] = pos
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a, aOK := index[candidates[i].UserID]
		b, bOK := index[candidates[j].UserID]

		switch {
		case aOK && bOK:
			return a < b
		case aOK:
			return true
		case bOK:
			return false
		default:
			return candidates[i].UserID < candidates[j].UserID
		}
	})

	return candidates
}

func startOfWeek(t time.Time, weekStart time.Weekday) time.Time {
	t = t.UTC()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	offset := (int(midnight.Weekday()) - int(weekStart) + 7) % 7

	return midnight.AddDate(0, 0, -offset)
}

func (s *service) computeMatch(ctx context.Context, userID string, candidateID string, minOverlap int) (*profilecard.MatchSummary, error) {
	matchSummary, err := s.matchingService.ComputeMatch(ctx, userID, candidateID, minOverlap)
	if err != nil {
		return nil, err
	}

	result := &profilecard.MatchSummary{
		MatchPercent: matchSummary.MatchPercent,
		OverlapCount: matchSummary.OverlapCount,
		Badges:       nil,
		HiddenReason: matchSummary.HiddenReason,
	}
	for _, badge := range matchSummary.Badges {
		result.Badges = append(result.Badges, profilecard.MatchBadge{
			QuestionID:    badge.QuestionID,
			QuestionText:  badge.QuestionText,
			PartnerAnswer: badge.PartnerAnswer,
			Weight:        badge.Weight,
		})
	}

	return result, nil
}

// passesPostQueryFilters applies filters that can't be efficiently done in SQL
func (s *service) passesPostQueryFilters(profile profilecard.ProfileCard, filters *domain.DiscoverFilters, currentUserProfile *profiledomain.EnrichedProfile) bool {
	if filters == nil {
		return true
	}

	// Apply distance filter if specified
	if filters.HasDistanceFilter() {
		if profile.DistanceKm > *filters.Distance.MaxDistanceKM {
			return false
		}
	}

	// Apply age filter if specified
	if filters.HasAgeFilter() {
		age := profile.Age
		if filters.AgeRange.MinAge != nil && age < *filters.AgeRange.MinAge {
			return false
		}

		if filters.AgeRange.MaxAge != nil && age > *filters.AgeRange.MaxAge {
			return false
		}
	}

	return true
}

type preferenceMatcher struct {
	prefs           *domain.StoredDiscoverPreferences
	datingIntentSet map[int16]struct{}
	religionSet     map[int16]struct{}
	ethnicitySet    map[int16]struct{}
}

func newPreferenceMatcher(prefs *domain.StoredDiscoverPreferences) *preferenceMatcher {
	if prefs == nil || !prefs.HasAnyPreference() {
		return nil
	}

	matcher := &preferenceMatcher{
		prefs: prefs,
	}

	if len(prefs.DatingIntentionIDs) > 0 {
		matcher.datingIntentSet = make(map[int16]struct{}, len(prefs.DatingIntentionIDs))
		for _, id := range prefs.DatingIntentionIDs {
			matcher.datingIntentSet[id] = struct{}{}
		}
	}

	if len(prefs.ReligionIDs) > 0 {
		matcher.religionSet = make(map[int16]struct{}, len(prefs.ReligionIDs))
		for _, id := range prefs.ReligionIDs {
			matcher.religionSet[id] = struct{}{}
		}
	}

	if len(prefs.EthnicityIDs) > 0 {
		matcher.ethnicitySet = make(map[int16]struct{}, len(prefs.EthnicityIDs))
		for _, id := range prefs.EthnicityIDs {
			matcher.ethnicitySet[id] = struct{}{}
		}
	}

	return matcher
}

func (m *preferenceMatcher) requiresEthnicity() bool {
	return m != nil && len(m.ethnicitySet) > 0
}

func (m *preferenceMatcher) matches(age int, distanceKM int, datingIntentionID *int16, religionID *int16, candidateEthnicities []int16) bool {
	if m == nil {
		return true
	}

	if m.prefs.DistanceKM != nil && distanceKM >= 0 {
		if distanceKM <= *m.prefs.DistanceKM {
			return true
		}
	}

	if m.prefs.MinAge != nil || m.prefs.MaxAge != nil {
		if m.ageWithinRange(age) {
			return true
		}
	}

	if len(m.datingIntentSet) > 0 && datingIntentionID != nil {
		if _, ok := m.datingIntentSet[*datingIntentionID]; ok {
			return true
		}
	}

	if len(m.religionSet) > 0 && religionID != nil {
		if _, ok := m.religionSet[*religionID]; ok {
			return true
		}
	}

	if len(m.ethnicitySet) > 0 && len(candidateEthnicities) > 0 {
		for _, id := range candidateEthnicities {
			if _, ok := m.ethnicitySet[id]; ok {
				return true
			}
		}
	}

	return false
}

func (m *preferenceMatcher) ageWithinRange(age int) bool {
	if m == nil || age <= 0 {
		return false
	}

	if m.prefs.MinAge != nil && age < *m.prefs.MinAge {
		return false
	}

	if m.prefs.MaxAge != nil && age > *m.prefs.MaxAge {
		return false
	}

	return true
}
