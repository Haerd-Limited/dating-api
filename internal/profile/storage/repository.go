package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/friendsofgo/errors"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type ProfileRepository interface {
	InsertProfile(ctx context.Context, userProfile *entity.UserProfile, tx *sql.Tx) error
	UpsertUserTheme(ctx context.Context, theme entity.UserTheme) error
	UpsertUserPrompts(ctx context.Context, userID string, prompts []entity.VoicePrompt) error
	UpsertUserPhotos(ctx context.Context, userID string, photos []entity.Photo) error
	UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16) error
	UpdateUserProfile(ctx context.Context, userProfile *entity.UserProfile, whiteList []string) error
	GetUserTheme(ctx context.Context, userID string) (*entity.UserTheme, error)
	GetUserProfileByUserID(ctx context.Context, userID string) (*entity.UserProfile, error)
	GetUserSpokenLanguages(ctx context.Context, userID string) ([]int16, error)
	GetUserVoicePrompts(ctx context.Context, userID string) ([]*entity.VoicePrompt, error)
	GetUserPhotos(ctx context.Context, userID string) ([]*entity.Photo, error)
	GetVoicePromptByID(ctx context.Context, id int64) (*entity.VoicePrompt, error)
	IsVerified(ctx context.Context, userID string) (bool, error)
}

type profileRepository struct {
	db *sqlx.DB
}

func NewProfileRepository(db *sqlx.DB) ProfileRepository {
	return &profileRepository{
		db: db,
	}
}

func (pr *profileRepository) IsVerified(ctx context.Context, userID string) (bool, error) {
	profile, err := entity.UserProfiles(entity.UserWhere.ID.EQ(userID)).One(ctx, pr.db)
	if err != nil {
		return false, err
	}

	return profile.Verified, nil
}

func (pr *profileRepository) GetVoicePromptByID(ctx context.Context, id int64) (*entity.VoicePrompt, error) {
	vp, err := entity.VoicePrompts(entity.VoicePromptWhere.ID.EQ(id)).One(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return vp, nil
}

func (pr *profileRepository) InsertProfile(ctx context.Context, userProfile *entity.UserProfile, tx *sql.Tx) error {
	err := userProfile.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (pr *profileRepository) UpsertUserTheme(ctx context.Context, theme entity.UserTheme) error {
	err := theme.Upsert(ctx, pr.db, true, []string{"user_id"},
		boil.Whitelist("base_hex", "palette", "updated_at"),
		boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (pr *profileRepository) GetUserTheme(ctx context.Context, userID string) (*entity.UserTheme, error) {
	ut, err := entity.UserThemes(
		entity.UserThemeWhere.UserID.EQ(userID),
		qm.Limit(1),
	).One(ctx, pr.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch user theme: %w", err)
	}

	return ut, nil
}

func (pr *profileRepository) GetPrompts(ctx context.Context) (entity.PromptTypeSlice, error) {
	prompts, err := entity.PromptTypes().All(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return prompts, nil
}

func (pr *profileRepository) UpsertUserPrompts(
	ctx context.Context,
	userID string,
	prompts []entity.VoicePrompt,
) (err error) {
	tx, err := pr.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// Replace existing set for this user.
	if _, err = tx.ExecContext(ctx, `DELETE FROM voice_prompts WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete existing voice_prompts: %w", err)
	}

	if len(prompts) == 0 {
		return nil
	}

	// Normalize: userID, default positions, exactly one primary, validate basics.
	primarySeen := false

	for i := range prompts {
		// user_id
		prompts[i].UserID = null.StringFrom(userID)

		// required: audio_url
		if strings.TrimSpace(prompts[i].AudioURL) == "" {
			return fmt.Errorf("prompt[%d]: audio_url is required", i)
		}

		// required: prompt_type (FK)
		if !prompts[i].PromptType.Valid {
			return fmt.Errorf("prompt[%d]: prompt_type is required", i)
		}

		// optional but sensible: non-negative duration
		if prompts[i].DurationMS < 0 {
			return fmt.Errorf("prompt[%d]: duration_ms cannot be negative", i)
		}

		// position defaults to 1-based index if not provided
		if !prompts[i].Position.Valid {
			prompts[i].Position = null.Int16From(int16(i + 1))
		}

		// ensure only one primary; keep the first, demote the rest
		if prompts[i].IsPrimary {
			if primarySeen {
				prompts[i].IsPrimary = false
			} else {
				primarySeen = true
			}
		}
	}

	// If none were marked primary, promote the first one.
	if !primarySeen {
		prompts[0].IsPrimary = true
	}

	// Insert all
	for i := range prompts {
		if err = prompts[i].Insert(ctx, tx, boil.Infer()); err != nil {
			return fmt.Errorf("insert voice_prompt[%d]: %w", i, err)
		}
	}

	return nil
}

func (pr *profileRepository) UpsertUserPhotos(
	ctx context.Context,
	userID string,
	photos []entity.Photo,
) (err error) {
	tx, err := pr.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// Replace existing set for this user.
	if _, err = tx.ExecContext(ctx, `DELETE FROM photos WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete existing photos: %w", err)
	}

	if len(photos) == 0 {
		return nil // nothing else to do
	}

	// Normalize: userID, positions, and exactly one primary.
	primarySeen := false

	for i := range photos {
		// Required URL check
		if strings.TrimSpace(photos[i].URL) == "" {
			return fmt.Errorf("photo[%d]: url is required", i)
		}

		// Assign user
		photos[i].UserID = null.StringFrom(userID)

		// Position: if not provided, make it 1-based index order
		if !photos[i].Position.Valid {
			photos[i].Position = null.Int16From(int16(i + 1))
		}

		// Primary handling: keep the first primary=true, demote subsequent ones
		if photos[i].IsPrimary {
			if primarySeen {
				photos[i].IsPrimary = false
			} else {
				primarySeen = true
			}
		}
	}

	// If none were marked primary, promote the first one.
	if !primarySeen {
		photos[0].IsPrimary = true
	}

	// Insert all
	for i := range photos {
		if err = photos[i].Insert(ctx, tx, boil.Infer()); err != nil {
			return fmt.Errorf("insert photo[%d]: %w", i, err)
		}
	}

	return nil
}

func (pr *profileRepository) UpsertUserSpokenLanguages(
	ctx context.Context,
	userID string,
	languages []int16,
) (err error) {
	tx, err := pr.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// Clear existing selections
	_, err = tx.ExecContext(ctx, `DELETE FROM user_languages WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user_languages: %w", err)
	}

	// Nothing to insert? we're done after clearing.
	if len(languages) == 0 {
		return nil
	}

	// De‑dupe in case the caller passed duplicates.
	uniq := make([]int16, 0, len(languages))

	seen := make(map[int16]struct{}, len(languages))
	for _, id := range languages {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}

		uniq = append(uniq, id)
	}

	// Build a single INSERT ... VALUES (...) ... statement.
	var (
		sb   strings.Builder
		args = make([]any, 1+len(uniq))
	)

	sb.WriteString(`INSERT INTO user_languages (user_id, language_id) VALUES `)

	args[0] = userID

	for i, lid := range uniq {
		if i > 0 {
			sb.WriteString(",")
		}
		// user_id is always $1; each language_id is $2, $3, ...
		sb.WriteString(fmt.Sprintf("($1,$%d)", i+2))

		args[i+1] = lid
	}

	// Ignore duplicates that could arise from race conditions.
	sb.WriteString(` ON CONFLICT (user_id, language_id) DO NOTHING`)

	if _, err = tx.ExecContext(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("insert user_languages: %w", err)
	}

	return nil
}

func (pr *profileRepository) GetUserProfileByUserID(ctx context.Context, userID string) (*entity.UserProfile, error) {
	userProfile, err := entity.UserProfiles(entity.UserProfileWhere.UserID.EQ(userID)).One(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return userProfile, nil
}

func (pr *profileRepository) UpdateUserProfile(ctx context.Context, userProfile *entity.UserProfile, whiteList []string) error {
	userProfile.UpdatedAt = time.Now().UTC()

	whiteList = append(whiteList, "updated_at")

	_, err := userProfile.Update(ctx, pr.db, boil.Whitelist(whiteList...))
	if err != nil {
		return err
	}

	return nil
}

func (pr *profileRepository) GetUserSpokenLanguages(ctx context.Context, userID string) ([]int16, error) {
	// Load the user
	u, err := entity.Users(
		entity.UserWhere.ID.EQ(userID),
		qm.Load(entity.UserRels.Languages),
	).One(ctx, pr.db)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	// Extract the IDs
	if u.R == nil || u.R.Languages == nil {
		return []int16{}, nil
	}

	ids := make([]int16, len(u.R.Languages))
	for i, lang := range u.R.Languages {
		ids[i] = lang.ID
	}

	return ids, nil
}

func (pr *profileRepository) GetUserVoicePrompts(ctx context.Context, userID string) ([]*entity.VoicePrompt, error) {
	vp, err := entity.VoicePrompts(entity.VoicePromptWhere.UserID.EQ(null.StringFrom(userID))).All(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return vp, nil
}

func (pr *profileRepository) GetUserPhotos(ctx context.Context, userID string) ([]*entity.Photo, error) {
	photos, err := entity.Photos(entity.PhotoWhere.UserID.EQ(null.StringFrom(userID))).All(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return photos, nil
}
