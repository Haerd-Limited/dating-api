package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type DiscoverRepository interface {
	GetDiscoverFeedCandidates(ctx context.Context, userID string, limit int, offset int) (entity.UserProfileSlice, error)
	GetDiscoverFeedCandidatesWithFilters(ctx context.Context, userID string, limit int, offset int, filters *domain.DiscoverFilters) (entity.UserProfileSlice, error)
	GetVoiceWorthHearing(ctx context.Context, userID string, limit int) ([]*entity.UserProfile, error)
	GetVoiceWorthHearingByIDs(ctx context.Context, userID string, candidateIDs []string) ([]*entity.UserProfile, error)
	GetLikeAndSuperlikeCount(ctx context.Context, userID string) (int64, error)
	AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error)
	GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error)
	GetNumberOfCompleteProfilesOfOppositeGender(ctx context.Context, userID string) (int64, error)
	GetSwipeUsageStats(ctx context.Context, userID string, window time.Duration, limit int, now time.Time) (int, *time.Time, error)
	SaveUserDiscoverPreferences(ctx context.Context, userID string, preferences *domain.DiscoverPreferenceUpdate) error
	GetUserDiscoverPreferences(ctx context.Context, userID string) (*domain.StoredDiscoverPreferences, error)
	GetUsersEthnicityIDs(ctx context.Context, userIDs []string) (map[string][]int16, error)
	GetWeeklyVoiceWorthHearingIDs(ctx context.Context, userID string, weekStart time.Time) ([]string, error)
	SaveWeeklyVoiceWorthHearingIDs(ctx context.Context, userID string, weekStart time.Time, candidateIDs []string) error
}

type discoverRepository struct {
	db *sqlx.DB
}

func NewDiscoverRepository(db *sqlx.DB) DiscoverRepository {
	return &discoverRepository{
		db: db,
	}
}

const stepComplete = "COMPLETE"

func (r *discoverRepository) GetLikeAndSuperlikeCount(ctx context.Context, userID string) (int64, error) {
	count, err := entity.Swipes(
		entity.SwipeWhere.TargetID.EQ(userID),
		qm.Where("action IN (?, ?)", constants.ActionLike, constants.ActionSuperlike),
	).Count(ctx, r.db)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *discoverRepository) AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error) {
	exists, err := entity.Swipes(
		qm.Where(
			"(actor_id = ? AND target_id = ?) OR (actor_id = ? AND target_id = ? AND action IN (?, ?))",
			userID, targetUserID, // you → them (any action)
			targetUserID, userID, // them → you
			constants.ActionLike, constants.ActionSuperlike,
		),
	).Exists(ctx, r.db)
	if err != nil {
		return false, fmt.Errorf("check interactions (%s <-> %s): %w", userID, targetUserID, err)
	}

	return exists, nil
}

func (r *discoverRepository) GetVoiceWorthHearing(ctx context.Context, userID string, limit int) ([]*entity.UserProfile, error) {
	numberOfOppositeGenderProfiles, err := r.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get number of complete profiles of opposite gender: %w", err)
	}

	if numberOfOppositeGenderProfiles <= constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
		return nil, nil
	}

	oppositeGender, err := r.getOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get opposite gender: %w", err)
	}

	if limit <= 0 {
		limit = 3
	}

	users, err := entity.UserProfiles(
		// not me
		entity.UserProfileWhere.UserID.NEQ(userID),

		// opposite gender
		entity.UserProfileWhere.GenderID.EQ(null.Int16From(oppositeGender.ID)),

		// only fully onboarded users
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step = ?", stepComplete),

		// Order by number of likes and superlikes
		qm.OrderBy(`(
			SELECT COUNT(*)
			FROM swipes s
			WHERE s.target_id = user_profiles.user_id
			AND s.action IN (?, ?)
		) DESC`, constants.ActionLike, constants.ActionSuperlike),

		// Limit to top N
		qm.Limit(limit),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get user profiles: %w", err)
	}

	return users, nil
}

func (r *discoverRepository) GetVoiceWorthHearingByIDs(ctx context.Context, userID string, candidateIDs []string) ([]*entity.UserProfile, error) {
	if len(candidateIDs) == 0 {
		return nil, nil
	}

	oppositeGender, err := r.getOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get opposite gender: %w", err)
	}

	placeholders := make([]string, len(candidateIDs))
	args := make([]interface{}, len(candidateIDs))

	for i, id := range candidateIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	users, err := entity.UserProfiles(
		qm.Where("user_profiles.user_id IN ("+strings.Join(placeholders, ",")+")", args...),
		entity.UserProfileWhere.UserID.NEQ(userID),
		entity.UserProfileWhere.GenderID.EQ(null.Int16From(oppositeGender.ID)),
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step = ?", stepComplete),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get user profiles by ids: %w", err)
	}

	return users, nil
}

func (r *discoverRepository) GetNumberOfCompleteProfilesOfOppositeGender(ctx context.Context, userID string) (int64, error) {
	oppositeGender, err := r.getOppositeGender(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get opposite gender: %w", err)
	}

	count, err := entity.UserProfiles(
		entity.UserProfileWhere.UserID.NEQ(userID),
		entity.UserProfileWhere.GenderID.EQ(null.Int16From(oppositeGender.ID)),
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step = ?", stepComplete),
	).Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("get user profiles: %w", err)
	}

	return count, nil
}

func (r *discoverRepository) GetSwipeUsageStats(ctx context.Context, userID string, window time.Duration, limit int, now time.Time) (int, *time.Time, error) {
	since := now.Add(-window)

	count64, err := entity.Swipes(
		entity.SwipeWhere.ActorID.EQ(userID),
		entity.SwipeWhere.CreatedAt.GTE(since),
	).Count(ctx, r.db)
	if err != nil {
		return 0, nil, fmt.Errorf("count swipes within window userID=%s: %w", userID, err)
	}

	count := int(count64)
	if count < limit {
		return count, nil, nil
	}

	k := count - limit + 1
	if k < 1 {
		k = 1
	}

	thresholdSwipe, err := entity.Swipes(
		entity.SwipeWhere.ActorID.EQ(userID),
		entity.SwipeWhere.CreatedAt.GTE(since),
		qm.OrderBy("created_at ASC"),
		qm.Offset(k-1),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return count, nil, fmt.Errorf("get gating swipe userID=%s: %w", userID, err)
	}

	t := thresholdSwipe.CreatedAt.UTC()

	return count, &t, nil
}

func (r *discoverRepository) GetDiscoverFeedCandidates(
	ctx context.Context,
	userID string,
	limit, offset int,
) (entity.UserProfileSlice, error) {
	oppositeGender, err := r.getOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get opposite gender: %w", err)
	}

	mods := []qm.QueryMod{
		entity.UserProfileWhere.UserID.NEQ(userID),
		entity.UserProfileWhere.GenderID.EQ(null.Int16From(oppositeGender.ID)),
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step = ?", stepComplete),

		// exclude anyone I've already swiped on (any action)
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM swipes s
                 WHERE s.actor_id  = ?
                   AND s.target_id = user_profiles.user_id
            )`, userID),

		// exclude users who already liked/superliked me
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM swipes s
                 WHERE s.actor_id = user_profiles.user_id
                   AND s.target_id = ?
                   AND s.action IN (?, ?)
            )`, userID, constants.ActionLike, constants.ActionSuperlike),

		// exclude blocked relationships (either direction)
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM user_blocks b
                 WHERE (b.blocker_user_id = ? AND b.blocked_user_id = user_profiles.user_id)
                    OR (b.blocker_user_id = user_profiles.user_id AND b.blocked_user_id = ?)
            )`, userID, userID),
	}

	excludeIDs, err := r.GetVoiceWorthHearingIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get VWH ids userID=%s: %w", userID, err)
	}

	// exclude Voice Worth Hearing candidates
	if len(excludeIDs) > 0 {
		ph := make([]string, len(excludeIDs))
		args := make([]interface{}, len(excludeIDs))

		for i, id := range excludeIDs {
			ph[i] = "?"
			args[i] = id
		}

		mods = append(mods, qm.Where(
			"user_profiles.user_id NOT IN ("+strings.Join(ph, ",")+")",
			args...,
		))
	}

	mods = append(mods, qm.Limit(limit), qm.Offset(offset))

	users, err := entity.UserProfiles(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get user profiles: %w", err)
	}

	return users, nil
}

func (r *discoverRepository) GetDiscoverFeedCandidatesWithFilters(
	ctx context.Context,
	userID string,
	limit, offset int,
	filters *domain.DiscoverFilters,
) (entity.UserProfileSlice, error) {
	oppositeGender, err := r.getOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get opposite gender: %w", err)
	}

	mods := []qm.QueryMod{
		entity.UserProfileWhere.UserID.NEQ(userID),
		entity.UserProfileWhere.GenderID.EQ(null.Int16From(oppositeGender.ID)),
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step = ?", stepComplete),

		// exclude anyone I've already swiped on (any action)
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM swipes s
                 WHERE s.actor_id  = ?
                   AND s.target_id = user_profiles.user_id
            )`, userID),

		// exclude users who already liked/superliked me
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM swipes s
                 WHERE s.actor_id = user_profiles.user_id
                   AND s.target_id = ?
                   AND s.action IN (?, ?)
            )`, userID, constants.ActionLike, constants.ActionSuperlike),

		// exclude blocked relationships (either direction)
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM user_blocks b
                 WHERE (b.blocker_user_id = ? AND b.blocked_user_id = user_profiles.user_id)
                    OR (b.blocker_user_id = user_profiles.user_id AND b.blocked_user_id = ?)
            )`, userID, userID),
	}

	// Apply filters if provided
	if filters != nil && !filters.IsEmpty() {
		filterMods, err := r.buildFilterMods(filters)
		if err != nil {
			return nil, fmt.Errorf("build filter mods: %w", err)
		}

		mods = append(mods, filterMods...)
	}

	excludeIDs, err := r.GetVoiceWorthHearingIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get VWH ids userID=%s: %w", userID, err)
	}

	// exclude Voice Worth Hearing candidates
	if len(excludeIDs) > 0 {
		ph := make([]string, len(excludeIDs))
		args := make([]interface{}, len(excludeIDs))

		for i, id := range excludeIDs {
			ph[i] = "?"
			args[i] = id
		}

		mods = append(mods, qm.Where(
			"user_profiles.user_id NOT IN ("+strings.Join(ph, ",")+")",
			args...,
		))
	}

	mods = append(mods, qm.Limit(limit), qm.Offset(offset))

	users, err := entity.UserProfiles(mods...).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get user profiles: %w", err)
	}

	return users, nil
}

func (r *discoverRepository) buildFilterMods(filters *domain.DiscoverFilters) ([]qm.QueryMod, error) {
	var mods []qm.QueryMod

	// Age range filter
	if filters.HasAgeFilter() {
		ageMods, err := r.buildAgeFilterMods(filters.AgeRange)
		if err != nil {
			return nil, fmt.Errorf("build age filter mods: %w", err)
		}

		mods = append(mods, ageMods...)
	}

	// Dating intentions filter
	if filters.HasDatingIntentionsFilter() {
		mods = append(mods, entity.UserProfileWhere.DatingIntentionID.IN(filters.DatingIntentions.IntentionIDs))
	}

	// Religion filter
	if filters.HasReligionsFilter() {
		mods = append(mods, entity.UserProfileWhere.ReligionID.IN(filters.Religions.ReligionIDs))
	}

	// Sexuality filter
	if filters.HasSexualitiesFilter() {
		mods = append(mods, entity.UserProfileWhere.SexualityID.IN(filters.Sexualities.SexualityIDs))
	}

	// Ethnicity filter (requires join with user_ethnicities table)
	if filters.HasEthnicitiesFilter() {
		mods = append(mods, qm.InnerJoin("user_ethnicities ue ON ue.user_id = user_profiles.user_id"))
		mods = append(mods, qm.WhereIn("ue.ethnicity_id IN ?", convertToInterfaceSlice(filters.Ethnicities.EthnicityIDs)...))
	}

	return mods, nil
}

func (r *discoverRepository) buildAgeFilterMods(ageFilter *domain.AgeRangeFilter) ([]qm.QueryMod, error) {
	var mods []qm.QueryMod

	if ageFilter.MinAge != nil {
		// Calculate the maximum birthdate for minimum age
		maxBirthdate := r.calculateMaxBirthdateForAge(*ageFilter.MinAge)
		mods = append(mods, qm.Where("user_profiles.birthdate <= ?", maxBirthdate))
	}

	if ageFilter.MaxAge != nil {
		// Calculate the minimum birthdate for maximum age
		minBirthdate := r.calculateMinBirthdateForAge(*ageFilter.MaxAge)
		mods = append(mods, qm.Where("user_profiles.birthdate >= ?", minBirthdate))
	}

	return mods, nil
}

func (r *discoverRepository) calculateMaxBirthdateForAge(minAge int) string {
	// Latest birthdate that still means "at least minAge today" (inclusive).
	// Someone born on this date is exactly minAge today.
	now := time.Now().UTC()
	cutoff := now.AddDate(-minAge, 0, 0)
	return cutoff.Format("2006-01-02")
}

func (r *discoverRepository) calculateMinBirthdateForAge(maxAge int) string {
	// Earliest birthdate that still means "at most maxAge today" (inclusive).
	// Someone born on this date is exactly maxAge today.
	now := time.Now().UTC()
	cutoff := now.AddDate(-maxAge, 0, 0)
	return cutoff.Format("2006-01-02")
}

func (r *discoverRepository) SaveUserDiscoverPreferences(ctx context.Context, userID string, preferences *domain.DiscoverPreferenceUpdate) error {
	if preferences == nil {
		return nil
	}

	up := &entity.UserPreference{
		UserID: userID,
	}

	// ClearAll: persist cleared state by setting all discover preference columns to null/empty
	if preferences.ClearAll {
		up.DistanceKM = null.Int16{}
		up.AgeMin = null.Int16{}
		up.AgeMax = null.Int16{}
		up.SeekIntentionIds = nil
		up.SeekReligionIds = nil
		up.SeekSexualityIds = nil
		up.SeekEthnicityIds = nil
		clearCols := []string{
			entity.UserPreferenceColumns.UpdatedAt,
			entity.UserPreferenceColumns.DistanceKM,
			entity.UserPreferenceColumns.AgeMin,
			entity.UserPreferenceColumns.AgeMax,
			entity.UserPreferenceColumns.SeekIntentionIds,
			entity.UserPreferenceColumns.SeekReligionIds,
			entity.UserPreferenceColumns.SeekSexualityIds,
			entity.UserPreferenceColumns.SeekEthnicityIds,
		}
		updateColumns := boil.Whitelist(clearCols...)
		insertColumns := boil.Whitelist(append([]string{entity.UserPreferenceColumns.UserID}, clearCols[1:]...)...)
		err := up.Upsert(ctx, r.db, true, []string{entity.UserPreferenceColumns.UserID}, updateColumns, insertColumns)
		if err != nil {
			return fmt.Errorf("save user discover preferences (clear all): %w", err)
		}
		return nil
	}

	// Set nullable int16 fields
	if preferences.DistanceKM != nil {
		up.DistanceKM = null.Int16From(int16(*preferences.DistanceKM))
	}

	if preferences.MinAge != nil {
		up.AgeMin = null.Int16From(int16(*preferences.MinAge))
	}

	if preferences.MaxAge != nil {
		up.AgeMax = null.Int16From(int16(*preferences.MaxAge))
	}

	// Set array fields (convert int16 to int64)
	if len(preferences.DatingIntentionIDs) > 0 {
		up.SeekIntentionIds = convertInt16SliceToInt64Array(preferences.DatingIntentionIDs)
	}

	if len(preferences.ReligionIDs) > 0 {
		up.SeekReligionIds = convertInt16SliceToInt64Array(preferences.ReligionIDs)
	}

	if len(preferences.SexualityIDs) > 0 {
		up.SeekSexualityIds = convertInt16SliceToInt64Array(preferences.SexualityIDs)
	}

	if len(preferences.EthnicityIDs) > 0 {
		up.SeekEthnicityIds = convertInt16SliceToInt64Array(preferences.EthnicityIDs)
	}

	// Build column whitelist for upsert - only include fields that are set
	updateCols := []string{entity.UserPreferenceColumns.UpdatedAt}
	insertCols := []string{entity.UserPreferenceColumns.UserID}

	if preferences.DistanceKM != nil {
		updateCols = append(updateCols, entity.UserPreferenceColumns.DistanceKM)
		insertCols = append(insertCols, entity.UserPreferenceColumns.DistanceKM)
	}

	if preferences.MinAge != nil {
		updateCols = append(updateCols, entity.UserPreferenceColumns.AgeMin)
		insertCols = append(insertCols, entity.UserPreferenceColumns.AgeMin)
	}

	if preferences.MaxAge != nil {
		updateCols = append(updateCols, entity.UserPreferenceColumns.AgeMax)
		insertCols = append(insertCols, entity.UserPreferenceColumns.AgeMax)
	}

	if len(preferences.DatingIntentionIDs) > 0 {
		updateCols = append(updateCols, entity.UserPreferenceColumns.SeekIntentionIds)
		insertCols = append(insertCols, entity.UserPreferenceColumns.SeekIntentionIds)
	}

	if len(preferences.ReligionIDs) > 0 {
		updateCols = append(updateCols, entity.UserPreferenceColumns.SeekReligionIds)
		insertCols = append(insertCols, entity.UserPreferenceColumns.SeekReligionIds)
	}

	if len(preferences.SexualityIDs) > 0 {
		updateCols = append(updateCols, entity.UserPreferenceColumns.SeekSexualityIds)
		insertCols = append(insertCols, entity.UserPreferenceColumns.SeekSexualityIds)
	}

	if len(preferences.EthnicityIDs) > 0 {
		updateCols = append(updateCols, entity.UserPreferenceColumns.SeekEthnicityIds)
		insertCols = append(insertCols, entity.UserPreferenceColumns.SeekEthnicityIds)
	}

	updateColumns := boil.Whitelist(updateCols...)
	insertColumns := boil.Whitelist(insertCols...)

	err := up.Upsert(ctx, r.db, true, []string{entity.UserPreferenceColumns.UserID}, updateColumns, insertColumns)
	if err != nil {
		return fmt.Errorf("save user discover preferences: %w", err)
	}

	return nil
}

func (r *discoverRepository) GetUserDiscoverPreferences(ctx context.Context, userID string) (*domain.StoredDiscoverPreferences, error) {
	up, err := entity.FindUserPreference(ctx, r.db, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get user discover preferences: %w", err)
	}

	return &domain.StoredDiscoverPreferences{
		DistanceKM:         fromNullInt16(up.DistanceKM),
		MinAge:             fromNullInt16(up.AgeMin),
		MaxAge:             fromNullInt16(up.AgeMax),
		DatingIntentionIDs: convertInt64ArrayToInt16Slice(up.SeekIntentionIds),
		ReligionIDs:        convertInt64ArrayToInt16Slice(up.SeekReligionIds),
		SexualityIDs:       convertInt64ArrayToInt16Slice(up.SeekSexualityIds),
		EthnicityIDs:       convertInt64ArrayToInt16Slice(up.SeekEthnicityIds),
	}, nil
}

func (r *discoverRepository) GetUsersEthnicityIDs(ctx context.Context, userIDs []string) (result map[string][]int16, err error) {
	if len(userIDs) == 0 {
		return map[string][]int16{}, nil
	}

	const query = `
		SELECT user_id, ethnicity_id
		FROM user_ethnicities
		WHERE user_id = ANY($1)
	`

	rows, err := r.db.QueryxContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, fmt.Errorf("get user ethnicity ids: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close user ethnicity rows: %w", closeErr)
		}
	}()

	result = make(map[string][]int16, len(userIDs))

	for rows.Next() {
		var userID string

		var ethnicityID int16

		if scanErr := rows.Scan(&userID, &ethnicityID); scanErr != nil {
			return nil, fmt.Errorf("scan user ethnicity row: %w", scanErr)
		}

		result[userID] = append(result[userID], ethnicityID)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, fmt.Errorf("iterate user ethnicity rows: %w", iterErr)
	}

	return result, nil
}

func convertToInterfaceSlice(int16Slice []int16) []interface{} {
	result := make([]interface{}, len(int16Slice))
	for i, v := range int16Slice {
		result[i] = v
	}

	return result
}

func convertInt16SliceToInt64Array(values []int16) types.Int64Array {
	if len(values) == 0 {
		return nil
	}

	result := make([]int64, len(values))
	for i, v := range values {
		result[i] = int64(v)
	}

	return result
}

func convertInt64ArrayToInt16Slice(values types.Int64Array) []int16 {
	if len(values) == 0 {
		return nil
	}

	result := make([]int16, len(values))
	for i, v := range values {
		result[i] = int16(v)
	}

	return result
}

func fromNullInt16(value null.Int16) *int {
	if !value.Valid {
		return nil
	}

	v := int(value.Int16)

	return &v
}

func (r *discoverRepository) getOppositeGender(ctx context.Context, userID string) (*entity.Gender, error) {
	// Step 1: Load the user profile with gender
	userProfile, err := entity.UserProfiles(
		entity.UserProfileWhere.UserID.EQ(userID),
	).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	if !userProfile.GenderID.Valid {
		return nil, fmt.Errorf("user profile gender is not set")
	}

	// Step 2: Load the current gender
	currentGender, err := entity.Genders(
		entity.GenderWhere.ID.EQ(userProfile.GenderID.Int16),
	).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current gender: %w", err)
	}

	// Step 3: Assume genders are stored as "male" / "female" keys in DB
	var oppositeKey string

	switch currentGender.Label {
	case "Male", "male":
		oppositeKey = "female"
	case "Female", "female":
		oppositeKey = "male"
	default:
		return nil, fmt.Errorf("unsupported gender label: %s", currentGender.Label)
	}

	// Step 4: Fetch opposite gender
	opposite, err := entity.Genders(
		entity.GenderWhere.Label.ILIKE(oppositeKey),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch opposite gender: %w", err)
	}

	return opposite, nil
}

func (r *discoverRepository) GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error) {
	count, err := r.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get number of complete profiles of opposite gender: %w", err)
	}

	if count <= constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
		return nil, nil
	}

	weekStart := startOfWeek(time.Now().UTC(), time.Sunday)

	cached, err := r.GetWeeklyVoiceWorthHearingIDs(ctx, userID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("get weekly cached vwh ids userID=%s: %w", userID, err)
	}

	if len(cached) > 0 {
		return cached, nil
	}

	profiles, err := r.GetVoiceWorthHearing(ctx, userID, constants.MaxNumberOfVWHUsersToSelect)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(profiles))
	for _, p := range profiles {
		ids = append(ids, p.UserID)
	}

	return ids, nil
}

func (r *discoverRepository) GetWeeklyVoiceWorthHearingIDs(ctx context.Context, userID string, weekStart time.Time) ([]string, error) {
	weekStart = normalizeWeekStart(weekStart)

	const query = `
SELECT candidate_ids
FROM voice_worth_hearing_weekly
WHERE user_id = $1
  AND week_start = $2
`

	var ids pq.StringArray

	err := r.db.QueryRowxContext(ctx, query, userID, weekStart).Scan(&ids)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("select weekly vwh ids userID=%s: %w", userID, err)
	}

	return append([]string(nil), ids...), nil
}

func (r *discoverRepository) SaveWeeklyVoiceWorthHearingIDs(ctx context.Context, userID string, weekStart time.Time, candidateIDs []string) error {
	weekStart = normalizeWeekStart(weekStart)

	const stmt = `
INSERT INTO voice_worth_hearing_weekly (user_id, week_start, candidate_ids)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, week_start)
DO UPDATE SET candidate_ids = EXCLUDED.candidate_ids,
              updated_at    = now()
`

	_, err := r.db.ExecContext(ctx, stmt, userID, weekStart, pq.StringArray(candidateIDs))
	if err != nil {
		return fmt.Errorf("upsert weekly vwh ids userID=%s: %w", userID, err)
	}

	return nil
}

func startOfWeek(t time.Time, weekStart time.Weekday) time.Time {
	t = t.UTC()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	offset := (int(midnight.Weekday()) - int(weekStart) + 7) % 7

	return midnight.AddDate(0, 0, -offset)
}

func normalizeWeekStart(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
