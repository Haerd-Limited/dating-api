package storage

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type DiscoverRepository interface {
	GetCandidates(ctx context.Context, userID string, limit int, offset int) (entity.UserProfileSlice, error)
}

type discoverRepository struct {
	db *sqlx.DB
}

func NewDiscoverRepository(db *sqlx.DB) DiscoverRepository {
	return &discoverRepository{
		db: db,
	}
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

func (r *discoverRepository) GetCandidates(
	ctx context.Context,
	userID string,
	limit, offset int,
) (entity.UserProfileSlice, error) {
	oppositeGender, err := r.getOppositeGender(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get opposite gender: %w", err)
	}

	const stepComplete = "COMPLETE"

	users, err := entity.UserProfiles(
		// not me
		entity.UserProfileWhere.UserID.NEQ(userID),

		// opposite gender
		entity.UserProfileWhere.GenderID.EQ(null.Int16From(oppositeGender.ID)),

		// only fully onboarded users
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step = ?", stepComplete),

		// exclude anyone I've already swiped on (like / pass / superlike — any action)
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM swipes s
                 WHERE s.actor_id  = ?
                   AND s.target_id = user_profiles.user_id
            )`, userID),

		// exclude users who have already liked or superliked me
		qm.Where(`
            NOT EXISTS (
                SELECT 1
                  FROM swipes s
                 WHERE s.actor_id = user_profiles.user_id
                   AND s.target_id = ?
                   AND s.type IN ('like', 'superlike')
            )`, userID),

		qm.Limit(limit),
		qm.Offset(offset),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	return users, nil
}
