package discover

import (
	"context"
	"errors"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/compatibility"
	"github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"github.com/Haerd-Limited/dating-api/internal/discover/storage"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/objects/profilecard"
)

type Service interface {
	GetDiscoverFeed(ctx context.Context, userID string, limit int, offset int) (domain.DiscoverFeedResult, error)
	GetDiscoverFeedWithFilters(ctx context.Context, userID string, limit int, offset int, filters *domain.DiscoverFilters) (domain.DiscoverFeedResult, error)
	GetVoiceWorthHearing(ctx context.Context, userID string) ([]profilecard.ProfileCard, error)
	GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error)
	AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error)
	GetUserPreferences(ctx context.Context, userID string) (*domain.StoredDiscoverPreferences, error)
	ComputeCompatibility(ctx context.Context, viewerID, targetID string) (*profilecard.CompatibilitySummary, error)
}

type service struct {
	logger               *zap.Logger
	profileService       profile.Service
	compatibilityService compatibility.Service
	discoverRepo         storage.DiscoverRepository
}

var ErrVoiceWorthHearingSearching = errors.New("voices worth hearing still searching")

func NewDiscoverService(
	logger *zap.Logger,
	profileService profile.Service,
	compatibilityService compatibility.Service,
	discoverRepo storage.DiscoverRepository,
) Service {
	return &service{
		logger:               logger,
		profileService:       profileService,
		compatibilityService: compatibilityService,
		discoverRepo:         discoverRepo,
	}
}

const (
	minOverlap                          = 5
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
		return domain.DiscoverFeedResult{}, commonlogger.LogError(s.logger, "failed to compute swipe usage", err, zap.String("userID", userID))
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
				return domain.DiscoverFeedResult{}, commonlogger.LogError(s.logger, "failed to persist discover preferences", err, zap.String("userID", userID), zap.Any("preferenceUpdate", preferenceUpdate))
			}
		}
	}

	candidates, err := s.discoverRepo.GetDiscoverFeedCandidatesWithFilters(ctx, userID, limit, offset, filters)
	if err != nil {
		return domain.DiscoverFeedResult{}, commonlogger.LogError(s.logger, "failed to get candidate IDs", err, zap.String("userID", userID), zap.Int("limit", limit), zap.Int("offset", offset))
	}

	if len(candidates) == 0 {
		return domain.NewDiscoverFeedResult(nil, quota), nil
	}

	currentUserProfile, err := s.profileService.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return domain.DiscoverFeedResult{}, commonlogger.LogError(s.logger, "failed to get current user profile", err, zap.String("userID", userID))
	}

	profiles := make([]profilecard.ProfileCard, 0, len(candidates))

	for _, candidate := range candidates {
		var profileErr error

		p, profileErr := s.profileService.GetProfileCardWithDistance(ctx, candidate.UserID, currentUserProfile.Latitude, currentUserProfile.Longitude)
		if profileErr != nil {
			return domain.DiscoverFeedResult{}, commonlogger.LogError(s.logger, "failed to get profile card", profileErr, zap.String("userID", userID), zap.String("profileUserID", candidate.UserID), zap.Float64("latitude", currentUserProfile.Latitude), zap.Float64("longitude", currentUserProfile.Longitude))
		}

		// Apply post-query filters that can't be done efficiently in SQL
		if filters != nil && !s.passesPostQueryFilters(p, filters, &currentUserProfile) {
			continue
		}

		p.CompatibilitySummary, profileErr = s.computeCompatibility(ctx, userID, candidate.UserID, minOverlap)
		if profileErr != nil {
			return domain.DiscoverFeedResult{}, commonlogger.LogError(s.logger, "failed to compute match", profileErr, zap.String("userID", userID), zap.String("profileUserID", candidate.UserID), zap.Int("minOverlap", minOverlap))
		}

		profiles = append(profiles, p)
	}

	return domain.NewDiscoverFeedResult(profiles, quota), nil
}

func (s *service) GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error) {
	profiles, err := s.GetVoiceWorthHearing(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrVoiceWorthHearingSearching) {
			return nil, nil
		}

		return nil, commonlogger.LogError(s.logger, "get voice worth hearing", err, zap.String("userID", userID))
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

	count, err := s.discoverRepo.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get number of complete profiles of opposite gender", err, zap.String("userID", userID))
	}

	if count <= constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
		return nil, ErrVoiceWorthHearingSearching
	}

	storedPreferences, err := s.discoverRepo.GetUserDiscoverPreferences(ctx, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get stored preferences", err, zap.String("userID", userID))
	}

	matcher := newPreferenceMatcher(storedPreferences)

	candidateLimit := constants.MaxNumberOfVWHUsersToSelect
	if matcher != nil {
		candidateLimit = voiceWorthHearingPreferencePoolSize
	}

	cachedIDs, err := s.discoverRepo.GetWeeklyVoiceWorthHearingIDs(ctx, userID, weekStart)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get cached voice worth hearing ids", err, zap.String("userID", userID), zap.Time("weekStart", weekStart))
	}

	var candidates []*entity.UserProfile

	switch {
	case len(cachedIDs) > 0:
		candidates, err = s.discoverRepo.GetVoiceWorthHearingByIDs(ctx, userID, cachedIDs)
		if err != nil {
			return nil, commonlogger.LogError(s.logger, "hydrate cached candidates", err, zap.String("userID", userID), zap.Strings("cachedIDs", cachedIDs))
		}

		candidates = orderCandidatesByIDs(candidates, cachedIDs)
	default:
		candidates, err = s.discoverRepo.GetVoiceWorthHearing(ctx, userID, candidateLimit)
		if err != nil {
			return nil, commonlogger.LogError(s.logger, "get candidates", err, zap.String("userID", userID), zap.Int("candidateLimit", candidateLimit))
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	currentUserProfile, err := s.profileService.GetEnrichedProfile(ctx, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "failed to get current user profile", err, zap.String("userID", userID), zap.Float64("latitude", currentUserProfile.Latitude), zap.Float64("longitude", currentUserProfile.Longitude))
	}

	var ethnicityByUser map[string][]int16

	if matcher != nil && matcher.requiresEthnicity() {
		userIDs := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			userIDs = append(userIDs, candidate.UserID)
		}

		ethnicityByUser, err = s.discoverRepo.GetUsersEthnicityIDs(ctx, userIDs)
		if err != nil {
			return nil, commonlogger.LogError(s.logger, "failed to load candidate ethnicities", err, zap.String("userID", userID), zap.Strings("userIDs", userIDs))
		}
	}

	profiles, selectedIDs, err := s.selectVoiceWorthHearingProfiles(ctx, matcher, candidates, userID, &currentUserProfile, ethnicityByUser)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "select voice worth hearing profiles", err, zap.String("userID", userID), zap.Any("matcher", matcher), zap.Int("len(candidates)", len(candidates)), zap.Float64("currentUserProfile.Latitude", currentUserProfile.Latitude), zap.Float64("currentUserProfile.Longitude", currentUserProfile.Longitude), zap.Any("ethnicityByUser", ethnicityByUser))
	}

	if len(profiles) == 0 {
		return nil, nil
	}

	if len(cachedIDs) == 0 {
		if err := s.discoverRepo.SaveWeeklyVoiceWorthHearingIDs(ctx, userID, weekStart, selectedIDs); err != nil {
			return nil, commonlogger.LogError(s.logger, "persist vwh cache", err, zap.String("userID", userID), zap.Time("weekStart", weekStart), zap.Strings("selectedIDs", selectedIDs))
		}
	}

	return profiles, nil
}

func (s *service) GetUserPreferences(ctx context.Context, userID string) (*domain.StoredDiscoverPreferences, error) {
	preferences, err := s.discoverRepo.GetUserDiscoverPreferences(ctx, userID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get user discover preferences", err, zap.String("userID", userID))
	}

	// Return empty preferences if none exist
	if preferences == nil {
		return &domain.StoredDiscoverPreferences{}, nil
	}

	return preferences, nil
}

func (s *service) computeCompatibility(ctx context.Context, userID string, candidateID string, minOverlap int) (*profilecard.CompatibilitySummary, error) {
	compatibilitySummary, err := s.compatibilityService.ComputeCompatibility(ctx, userID, candidateID, minOverlap)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "compute compatibility", err, zap.String("userID", userID), zap.String("candidateID", candidateID), zap.Int("minOverlap", minOverlap))
	}

	result := &profilecard.CompatibilitySummary{
		CompatibilityPercent: compatibilitySummary.CompatibilityPercent,
		OverlapCount:         compatibilitySummary.OverlapCount,
		Badges:               nil,
		HiddenReason:         compatibilitySummary.HiddenReason,
	}
	for _, badge := range compatibilitySummary.Badges {
		result.Badges = append(result.Badges, profilecard.CompatibilityBadge{
			QuestionID:    badge.QuestionID,
			QuestionText:  badge.QuestionText,
			PartnerAnswer: badge.PartnerAnswer,
			Weight:        badge.Weight,
			IsMismatch:    badge.IsMismatch,
			RequirementBy: badge.RequirementBy,
		})
	}

	return result, nil
}

func (s *service) ComputeCompatibility(ctx context.Context, viewerID, targetID string) (*profilecard.CompatibilitySummary, error) {
	return s.computeCompatibility(ctx, viewerID, targetID, minOverlap)
}

type preferenceMatcher struct {
	prefs           *domain.StoredDiscoverPreferences
	datingIntentSet map[int16]struct{}
	religionSet     map[int16]struct{}
	sexualitySet    map[int16]struct{}
	ethnicitySet    map[int16]struct{}
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

	if len(prefs.SexualityIDs) > 0 {
		matcher.sexualitySet = make(map[int16]struct{}, len(prefs.SexualityIDs))
		for _, id := range prefs.SexualityIDs {
			matcher.sexualitySet[id] = struct{}{}
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

func (m *preferenceMatcher) matchesAll(age int, distanceKm int, datingIntentionID *int16, religionID *int16, sexualityID *int16, candidateEthnicities []int16) bool {
	if m == nil {
		return true
	}

	if m.prefs.DistanceKM != nil {
		if distanceKm < 0 || distanceKm > *m.prefs.DistanceKM {
			return false
		}
	}

	if (m.prefs.MinAge != nil || m.prefs.MaxAge != nil) && !m.ageWithinRange(age) {
		return false
	}

	if len(m.datingIntentSet) > 0 {
		if datingIntentionID == nil {
			return false
		}

		if _, ok := m.datingIntentSet[*datingIntentionID]; !ok {
			return false
		}
	}

	if len(m.religionSet) > 0 {
		if religionID == nil {
			return false
		}

		if _, ok := m.religionSet[*religionID]; !ok {
			return false
		}
	}

	if len(m.sexualitySet) > 0 {
		if sexualityID == nil {
			return false
		}

		if _, ok := m.sexualitySet[*sexualityID]; !ok {
			return false
		}
	}

	if len(m.ethnicitySet) > 0 {
		if len(candidateEthnicities) == 0 {
			return false
		}

		for _, id := range candidateEthnicities {
			if _, ok := m.ethnicitySet[id]; ok {
				return true
			}
		}

		return false
	}

	return true
}

func (m *preferenceMatcher) matchesAny(age int, distanceKm int, datingIntentionID *int16, religionID *int16, sexualityID *int16, candidateEthnicities []int16) bool {
	if m == nil {
		return true
	}

	if m.prefs.DistanceKM != nil && distanceKm >= 0 && distanceKm <= *m.prefs.DistanceKM {
		return true
	}

	if (m.prefs.MinAge != nil || m.prefs.MaxAge != nil) && m.ageWithinRange(age) {
		return true
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

	if len(m.sexualitySet) > 0 && sexualityID != nil {
		if _, ok := m.sexualitySet[*sexualityID]; ok {
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

type candidateEvaluation struct {
	card              profilecard.ProfileCard
	matchesAll        bool
	matchesAny        bool
	alreadyInteracted bool
}

func (s *service) selectVoiceWorthHearingProfiles(
	ctx context.Context,
	matcher *preferenceMatcher,
	candidates []*entity.UserProfile,
	userID string,
	currentUserProfile *profiledomain.EnrichedProfile,
	ethnicityByUser map[string][]int16,
) ([]profilecard.ProfileCard, []string, error) {
	evaluationCache := make(map[string]*candidateEvaluation, len(candidates))

	selectionPasses := s.buildSelectionPasses(matcher)

	selectedProfiles := make([]profilecard.ProfileCard, 0, constants.MaxNumberOfVWHUsersToSelect)
	selectedIDs := make([]string, 0, constants.MaxNumberOfVWHUsersToSelect)

	for _, pass := range selectionPasses {
		if len(selectedProfiles) == constants.MaxNumberOfVWHUsersToSelect {
			break
		}

		for _, candidate := range candidates {
			if len(selectedProfiles) == constants.MaxNumberOfVWHUsersToSelect {
				break
			}

			if containsID(selectedIDs, candidate.UserID) {
				continue
			}

			evaluation, err := s.evaluateCandidate(ctx, matcher, candidate, currentUserProfile, ethnicityByUser, userID, evaluationCache)
			if err != nil {
				return nil, nil, err
			}

			if evaluation.alreadyInteracted {
				continue
			}

			if !pass(evaluation) {
				continue
			}

			selectedProfiles = append(selectedProfiles, evaluation.card)
			selectedIDs = append(selectedIDs, candidate.UserID)
		}
	}

	return selectedProfiles, selectedIDs, nil
}

func (s *service) buildSelectionPasses(matcher *preferenceMatcher) []func(*candidateEvaluation) bool {
	passes := make([]func(*candidateEvaluation) bool, 0, 3)

	if matcher != nil {
		passes = append(passes, func(eval *candidateEvaluation) bool { return eval.matchesAll })
		passes = append(passes, func(eval *candidateEvaluation) bool { return eval.matchesAny })
	}

	passes = append(passes, func(eval *candidateEvaluation) bool { return true })

	return passes
}

func (s *service) evaluateCandidate(
	ctx context.Context,
	matcher *preferenceMatcher,
	candidate *entity.UserProfile,
	currentUserProfile *profiledomain.EnrichedProfile,
	ethnicityByUser map[string][]int16,
	userID string,
	cache map[string]*candidateEvaluation,
) (*candidateEvaluation, error) {
	if evaluation, exists := cache[candidate.UserID]; exists {
		return evaluation, nil
	}

	evaluation := &candidateEvaluation{}

	alreadyInteracted, err := s.discoverRepo.AlreadyInteracted(ctx, userID, candidate.UserID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "already interacted", err, zap.String("userID", userID), zap.String("profileUserID", candidate.UserID))
	}

	evaluation.alreadyInteracted = alreadyInteracted
	if alreadyInteracted {
		cache[candidate.UserID] = evaluation
		return evaluation, nil
	}

	card, err := s.profileService.GetProfileCardWithDistance(ctx, candidate.UserID, currentUserProfile.Latitude, currentUserProfile.Longitude)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get profile card", err, zap.String("userID", userID), zap.String("profileUserID", candidate.UserID), zap.Float64("latitude", currentUserProfile.Latitude), zap.Float64("longitude", currentUserProfile.Longitude))
	}

	likeCount, err := s.discoverRepo.GetLikeAndSuperlikeCount(ctx, candidate.UserID)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "get like and superlike count", err, zap.String("userID", userID), zap.String("profileUserID", candidate.UserID), zap.Int("minOverlap", minOverlap))
	}

	card.LikeCount = &likeCount

	card.CompatibilitySummary, err = s.computeCompatibility(ctx, userID, candidate.UserID, minOverlap)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "failed to compute compatibility", err, zap.String("userID", userID), zap.String("profileUserID", candidate.UserID), zap.Int("minOverlap", minOverlap))
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

	var sexualityID *int16

	if candidate.SexualityID.Valid {
		value := candidate.SexualityID.Int16
		sexualityID = &value
	}

	var candidateEthnicities []int16
	if matcher != nil && matcher.requiresEthnicity() {
		candidateEthnicities = ethnicityByUser[candidate.UserID]
	}

	if matcher == nil {
		evaluation.matchesAll = true
		evaluation.matchesAny = true
	} else {
		evaluation.matchesAll = matcher.matchesAll(card.Age, card.DistanceKm, datingIntentionID, religionID, sexualityID, candidateEthnicities)
		evaluation.matchesAny = matcher.matchesAny(card.Age, card.DistanceKm, datingIntentionID, religionID, sexualityID, candidateEthnicities)
	}

	evaluation.card = card
	cache[candidate.UserID] = evaluation

	return evaluation, nil
}

func containsID(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}

	return false
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
