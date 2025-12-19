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
	UpsertUserTheme(ctx context.Context, theme entity.UserTheme, tx *sql.Tx) error
	UpsertUserPrompts(ctx context.Context, userID string, prompts []entity.VoicePrompt, tx *sql.Tx) error
	UpsertUserPhotos(ctx context.Context, userID string, photos []entity.Photo, tx *sql.Tx) error
	UpsertUserSpokenLanguages(ctx context.Context, userID string, languages []int16, tx *sql.Tx) error
	UpsertUserEthnicities(ctx context.Context, userID string, ethnicities []int16, tx *sql.Tx) error
	UpdateUserProfile(ctx context.Context, userProfile *entity.UserProfile, whiteList []string, tx *sql.Tx) error
	GetUserTheme(ctx context.Context, userID string) (*entity.UserTheme, error)
	GetUserProfileByUserID(ctx context.Context, userID string) (*entity.UserProfile, error)
	GetUserSpokenLanguages(ctx context.Context, userID string) ([]int16, error)
	GetUserEthnicities(ctx context.Context, userID string) ([]int16, error)
	GetUserVoicePrompts(ctx context.Context, userID string) ([]*entity.VoicePrompt, error)
	GetUserPhotos(ctx context.Context, userID string) ([]*entity.Photo, error)
	GetVoicePromptByID(ctx context.Context, id int64) (*entity.VoicePrompt, error)
	IsVerified(ctx context.Context, userID string) (bool, error)
	GetVerificationStatus(ctx context.Context, userID string) (string, error)
	UpdateVoicePromptTranscript(ctx context.Context, id int64, transcript string) error
	// Stats helpers
	CountUsersBasicsCompletedByGender(ctx context.Context, genderID int16) (int64, error)
	CountUsersBasicsCompleted(ctx context.Context) (int64, error)
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
	status, err := pr.GetVerificationStatus(ctx, userID)
	if err != nil {
		return false, err
	}

	return status == "VERIFIED", nil
}

func (pr *profileRepository) GetVerificationStatus(ctx context.Context, userID string) (string, error) {
	profile, err := entity.UserProfiles(entity.UserProfileWhere.UserID.EQ(userID)).One(ctx, pr.db)
	if err != nil {
		return "", err
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

func (pr *profileRepository) UpdateVoicePromptTranscript(ctx context.Context, id int64, transcript string) error {
	vp, err := entity.VoicePrompts(entity.VoicePromptWhere.ID.EQ(id)).One(ctx, pr.db)
	if err != nil {
		return fmt.Errorf("voice prompt not found: %w", err)
	}

	vp.Transcript = null.StringFrom(transcript)

	_, err = vp.Update(ctx, pr.db, boil.Whitelist(entity.VoicePromptColumns.Transcript))
	if err != nil {
		return fmt.Errorf("failed to update transcript: %w", err)
	}

	return nil
}

// CountUsersBasicsCompletedByGender returns count of users who have completed BASICS (i.e., advanced to LOCATION or beyond) for a given gender.
func (pr *profileRepository) CountUsersBasicsCompletedByGender(ctx context.Context, genderID int16) (int64, error) {
	// Any step AFTER BASICS means BASICS is completed. We count users with step in the following list.
	const stepComplete = "COMPLETE"
	stepsAfterBasics := []string{
		"LOCATION",
		"LIFESTYLE",
		"BELIEFS",
		"BACKGROUND",
		"WORK_AND_EDUCATION",
		"LANGUAGES",
		"PHOTOS",
		"PROMPTS",
		"PROFILE",
		stepComplete,
	}

	// Build placeholders for IN clause
	args := make([]any, 0, len(stepsAfterBasics)+1)
	args = append(args, genderID)

	for _, s := range stepsAfterBasics {
		args = append(args, s)
	}

	queryMods := []qm.QueryMod{
		entity.UserProfileWhere.GenderID.EQ(null.Int16From(genderID)),
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step IN ("+strings.Repeat("?,", len(stepsAfterBasics)-1)+"?)", args[1:]...),
	}

	count, err := entity.UserProfiles(queryMods...).Count(ctx, pr.db)
	if err != nil {
		return 0, fmt.Errorf("count basics-completed by gender: %w", err)
	}

	return count, nil
}

// CountUsersBasicsCompleted returns total number of users that have completed BASICS.
func (pr *profileRepository) CountUsersBasicsCompleted(ctx context.Context) (int64, error) {
	stepsAfterBasics := []string{
		"LOCATION",
		"LIFESTYLE",
		"BELIEFS",
		"BACKGROUND",
		"WORK_AND_EDUCATION",
		"LANGUAGES",
		"PHOTOS",
		"PROMPTS",
		"PROFILE",
		"COMPLETE",
	}

	queryMods := []qm.QueryMod{
		qm.InnerJoin("users u ON u.id = user_profiles.user_id"),
		qm.Where("u.onboarding_step IN ("+strings.Repeat("?,", len(stepsAfterBasics)-1)+"?)", anySlice(stepsAfterBasics)...),
	}

	count, err := entity.UserProfiles(queryMods...).Count(ctx, pr.db)
	if err != nil {
		return 0, fmt.Errorf("count basics-completed: %w", err)
	}

	return count, nil
}

// helper to convert []string to []any
func anySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i := range ss {
		out[i] = ss[i]
	}

	return out
}

func (pr *profileRepository) InsertProfile(ctx context.Context, userProfile *entity.UserProfile, tx *sql.Tx) error {
	err := userProfile.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (pr *profileRepository) UpsertUserTheme(ctx context.Context, theme entity.UserTheme, tx *sql.Tx) error {
	exec := pr.executor(tx)

	err := theme.Upsert(ctx, exec, true, []string{"user_id"},
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
	tx *sql.Tx,
) (err error) {
	var execTx *sql.Tx

	var beginTx *sqlx.Tx

	if tx != nil {
		execTx = tx
	} else {
		beginTx, err = pr.db.BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		execTx = beginTx.Tx

		defer func() {
			if err != nil {
				_ = beginTx.Rollback()
			} else {
				_ = beginTx.Commit()
			}
		}()
	}

	// Mark existing prompts as inactive instead of deleting them
	// This preserves prompts that may be referenced in conversations/swipes
	if _, err = execTx.ExecContext(ctx, `UPDATE voice_prompts SET is_active = FALSE WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("mark existing voice_prompts as inactive: %w", err)
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

	// Insert all new prompts (they will be active by default due to database DEFAULT TRUE)
	// Note: Once SQLBoiler entities are regenerated with is_active field, we should explicitly
	// set prompts[i].IsActive = true here for clarity
	for i := range prompts {
		if err = prompts[i].Insert(ctx, execTx, boil.Infer()); err != nil {
			return fmt.Errorf("insert voice_prompt[%d]: %w", i, err)
		}
	}

	return nil
}

func (pr *profileRepository) UpsertUserPhotos(
	ctx context.Context,
	userID string,
	photos []entity.Photo,
	tx *sql.Tx,
) (err error) {
	var execTx *sql.Tx

	var beginTx *sqlx.Tx

	if tx != nil {
		execTx = tx
	} else {
		beginTx, err = pr.db.BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		execTx = beginTx.Tx

		defer func() {
			if err != nil {
				_ = beginTx.Rollback()
			} else {
				_ = beginTx.Commit()
			}
		}()
	}

	// Replace existing set for this user.
	if _, err = execTx.ExecContext(ctx, `DELETE FROM photos WHERE user_id = $1`, userID); err != nil {
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
		if err = photos[i].Insert(ctx, execTx, boil.Infer()); err != nil {
			return fmt.Errorf("insert photo[%d]: %w", i, err)
		}
	}

	return nil
}

func (pr *profileRepository) UpsertUserSpokenLanguages(
	ctx context.Context,
	userID string,
	languages []int16,
	tx *sql.Tx,
) (err error) {
	var execTx *sql.Tx

	var beginTx *sqlx.Tx

	if tx != nil {
		execTx = tx
	} else {
		beginTx, err = pr.db.BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		execTx = beginTx.Tx

		defer func() {
			if err != nil {
				_ = beginTx.Rollback()
			} else {
				_ = beginTx.Commit()
			}
		}()
	}

	// Clear existing selections
	_, err = execTx.ExecContext(ctx, `DELETE FROM user_languages WHERE user_id = $1`, userID)
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

	if _, err = execTx.ExecContext(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("insert user_languages: %w", err)
	}

	return nil
}

func (pr *profileRepository) UpsertUserEthnicities(
	ctx context.Context,
	userID string,
	ethnicities []int16,
	tx *sql.Tx,
) (err error) {
	var execTx *sql.Tx

	var beginTx *sqlx.Tx

	if tx != nil {
		execTx = tx
	} else {
		beginTx, err = pr.db.BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}

		execTx = beginTx.Tx

		defer func() {
			if err != nil {
				_ = beginTx.Rollback()
			} else {
				_ = beginTx.Commit()
			}
		}()
	}

	// Clear existing selections
	_, err = execTx.ExecContext(ctx, `DELETE FROM user_ethnicities WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user_ethnicities: %w", err)
	}

	// Nothing to insert? we're done after clearing.
	if len(ethnicities) == 0 {
		return nil
	}

	// De‑dupe in case the caller passed duplicates.
	uniq := make([]int16, 0, len(ethnicities))

	seen := make(map[int16]struct{}, len(ethnicities))
	for _, id := range ethnicities {
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

	sb.WriteString(`INSERT INTO user_ethnicities (user_id, ethnicity_id) VALUES `)

	args[0] = userID

	for i, eid := range uniq {
		if i > 0 {
			sb.WriteString(",")
		}
		// user_id is always $1; each ethnicity_id is $2, $3, ...
		sb.WriteString(fmt.Sprintf("($1,$%d)", i+2))

		args[i+1] = eid
	}

	// Ignore duplicates that could arise from race conditions.
	sb.WriteString(` ON CONFLICT (user_id, ethnicity_id) DO NOTHING`)

	if _, err = execTx.ExecContext(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("insert user_ethnicities: %w", err)
	}

	return nil
}

func (pr *profileRepository) GetUserEthnicities(ctx context.Context, userID string) ([]int16, error) {
	var ethnicityIDs []int16

	query := `SELECT ethnicity_id FROM user_ethnicities WHERE user_id = $1 ORDER BY ethnicity_id`

	rows, err := pr.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user_ethnicities: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var id int16
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan ethnicity_id: %w", err)
		}

		ethnicityIDs = append(ethnicityIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ethnicity rows: %w", err)
	}

	return ethnicityIDs, nil
}

func (pr *profileRepository) GetUserProfileByUserID(ctx context.Context, userID string) (*entity.UserProfile, error) {
	userProfile, err := entity.UserProfiles(entity.UserProfileWhere.UserID.EQ(userID)).One(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return userProfile, nil
}

func (pr *profileRepository) UpdateUserProfile(ctx context.Context, userProfile *entity.UserProfile, whiteList []string, tx *sql.Tx) error {
	userProfile.UpdatedAt = time.Now().UTC()

	whiteList = append(whiteList, "updated_at")

	exec := pr.executor(tx)

	_, err := userProfile.Update(ctx, exec, boil.Whitelist(whiteList...))
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
	// Only return active prompts for display purposes
	// Historical prompts (inactive) are preserved for conversations but not shown in profile
	vp, err := entity.VoicePrompts(
		entity.VoicePromptWhere.UserID.EQ(null.StringFrom(userID)),
		qm.Where("is_active = TRUE"),
	).All(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return vp, nil
}

func (pr *profileRepository) executor(tx *sql.Tx) boil.ContextExecutor {
	if tx != nil {
		return tx
	}

	return pr.db
}

func (pr *profileRepository) GetUserPhotos(ctx context.Context, userID string) ([]*entity.Photo, error) {
	photos, err := entity.Photos(entity.PhotoWhere.UserID.EQ(null.StringFrom(userID))).All(ctx, pr.db)
	if err != nil {
		return nil, err
	}

	return photos, nil
}
