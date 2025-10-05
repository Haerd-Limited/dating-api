package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/constants"
)

type DiscoverRepository interface {
	GetDiscoverFeedCandidates(ctx context.Context, userID string, limit int, offset int) (entity.UserProfileSlice, error)
	GetVoiceWorthHearing(ctx context.Context, userID string) ([]*entity.UserProfile, error)
	GetLikeAndSuperlikeCount(ctx context.Context, userID string) (int64, error)
	AlreadyInteracted(ctx context.Context, userID string, targetUserID string) (bool, error)
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

	excludeIDs, err := r.getVoiceWorthHearingIDs(ctx, userID)
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

func (r *discoverRepository) getVoiceWorthHearingIDs(ctx context.Context, userID string) ([]string, error) {
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
