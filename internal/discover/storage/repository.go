package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/discover/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type DiscoverRepository interface {
	GetDiscoverFeedCandidates(ctx context.Context, userID string, limit int, offset int) (entity.UserProfileSlice, error)
	GetDiscoverFeedCandidatesWithFilters(ctx context.Context, userID string, limit int, offset int, filters *domain.DiscoverFilters) (entity.UserProfileSlice, error)
	GetVoiceWorthHearing(ctx context.Context, userID string) ([]*entity.UserProfile, error)
	GetLikeAndSuperlikeCount(ctx context.Context, userID string) (int64, error)
	AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error)
	GetVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error)
	GetNumberOfCompleteProfilesOfOppositeGender(ctx context.Context, userID string) (int64, error)
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

func (r *discoverRepository) GetVoiceWorthHearing(ctx context.Context, userID string) ([]*entity.UserProfile, error) {
	numberOfOppositeGenderProfiles, err := r.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get number of complete profiles of opposite gender: %w", err)
	}

	if numberOfOppositeGenderProfiles < constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
		return nil, nil
	}

	oppositeGender, err := r.getOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get opposite gender: %w", err)
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

		// Limit to top 3
		qm.Limit(3),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get user profiles: %w", err)
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
	}

	count, err := r.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get number of complete profiles of opposite gender: %w", err)
	}

	if count > constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
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
	}

	// Apply filters if provided
	if filters != nil && !filters.IsEmpty() {
		filterMods, err := r.buildFilterMods(filters)
		if err != nil {
			return nil, fmt.Errorf("build filter mods: %w", err)
		}

		mods = append(mods, filterMods...)
	}

	count, err := r.GetNumberOfCompleteProfilesOfOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get number of complete profiles of opposite gender: %w", err)
	}

	if count > constants.MinimumNumberOfUsersRequiredToBuildVwhUsers {
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
	// Calculate the latest birthdate that would result in the minimum age
	// This is a simplified calculation - in production you might want to use time.Time
	year := 2024 - minAge // Assuming current year is 2024
	return fmt.Sprintf("%d-12-31", year)
}

func (r *discoverRepository) calculateMinBirthdateForAge(maxAge int) string {
	// Calculate the earliest birthdate that would result in the maximum age
	year := 2024 - maxAge // Assuming current year is 2024
	return fmt.Sprintf("%d-01-01", year)
}

func convertToInterfaceSlice(int16Slice []int16) []interface{} {
	result := make([]interface{}, len(int16Slice))
	for i, v := range int16Slice {
		result[i] = v
	}

	return result
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
	profiles, err := r.GetVoiceWorthHearing(ctx, userID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(profiles))
	for _, p := range profiles {
		ids = append(ids, p.UserID)
	}

	return ids, nil
}
