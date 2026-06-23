package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type ReportListFilter struct {
	Statuses   []string
	Categories []string
	ReporterID *string
	ReportedID *string
	Limit      int
	Offset     int
}

//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
type Repository interface {
	CreateBlock(ctx context.Context, block *entity.UserBlock, tx *sql.Tx) error
	IsBlockedPair(ctx context.Context, userID, otherUserID string) (bool, error)
	ListBlocksForUser(ctx context.Context, userID string) (entity.UserBlockSlice, error)

	CreateReport(ctx context.Context, report *entity.UserReport, tx *sql.Tx) error
	GetReportByID(ctx context.Context, reportID string) (*entity.UserReport, error)
	ListReports(ctx context.Context, filter ReportListFilter) (entity.UserReportSlice, error)
	UpdateReport(ctx context.Context, report *entity.UserReport, tx *sql.Tx) error
	InsertReportAction(ctx context.Context, action *entity.ReportAction, reviewerName string, tx *sql.Tx) error
	LoadReviewerNamesForReport(ctx context.Context, reportID string) (map[string]string, error)

	InsertModerationWarning(ctx context.Context, warning *entity.UserModerationWarning, tx *sql.Tx) error
	ListUnacknowledgedWarnings(ctx context.Context, userID string) (entity.UserModerationWarningSlice, error)
	GetWarningByID(ctx context.Context, warningID string) (*entity.UserModerationWarning, error)
	AcknowledgeWarning(ctx context.Context, warningID, userID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateBlock(ctx context.Context, block *entity.UserBlock, tx *sql.Tx) error {
	exec := r.executor(tx)

	return block.Upsert(
		ctx,
		exec,
		true,
		[]string{entity.UserBlockColumns.BlockerUserID, entity.UserBlockColumns.BlockedUserID},
		boil.Whitelist(entity.UserBlockColumns.Reason),
		boil.Infer(),
	)
}

func (r *repository) IsBlockedPair(ctx context.Context, userID, otherUserID string) (bool, error) {
	return entity.UserBlocks(
		qm.Where("(blocker_user_id = ? AND blocked_user_id = ?) OR (blocker_user_id = ? AND blocked_user_id = ?)",
			userID, otherUserID, otherUserID, userID),
	).Exists(ctx, r.db)
}

func (r *repository) ListBlocksForUser(ctx context.Context, userID string) (entity.UserBlockSlice, error) {
	return entity.UserBlocks(
		qm.Where("(blocker_user_id = ? OR blocked_user_id = ?)", userID, userID),
	).All(ctx, r.db)
}

func (r *repository) CreateReport(ctx context.Context, report *entity.UserReport, tx *sql.Tx) error {
	exec := r.executor(tx)
	return report.Insert(ctx, exec, boil.Infer())
}

func (r *repository) GetReportByID(ctx context.Context, reportID string) (*entity.UserReport, error) {
	return entity.UserReports(
		qm.Where("user_reports.id = ?", reportID),
		qm.Load(
			entity.UserReportRels.ReportReportActions,
			qm.OrderBy(entity.ReportActionColumns.CreatedAt+" ASC"),
		),
	).One(ctx, r.db)
}

func (r *repository) ListReports(ctx context.Context, filter ReportListFilter) (entity.UserReportSlice, error) {
	mods := []qm.QueryMod{
		qm.Load(
			entity.UserReportRels.ReportReportActions,
			qm.OrderBy(entity.ReportActionColumns.CreatedAt+" ASC"),
		),
		qm.OrderBy(entity.UserReportColumns.CreatedAt + " DESC"),
	}

	if len(filter.Statuses) > 0 {
		mods = append(mods, qm.WhereIn(
			"user_reports.status IN ?",
			stringSliceToInterface(filter.Statuses)...,
		))
	}

	if len(filter.Categories) > 0 {
		mods = append(mods, qm.WhereIn(
			"user_reports.category IN ?",
			stringSliceToInterface(filter.Categories)...,
		))
	}

	if filter.ReporterID != nil {
		mods = append(mods, qm.Where("user_reports.reporter_user_id = ?", *filter.ReporterID))
	}

	if filter.ReportedID != nil {
		mods = append(mods, qm.Where("user_reports.reported_user_id = ?", *filter.ReportedID))
	}

	if filter.Limit > 0 {
		mods = append(mods, qm.Limit(filter.Limit))
	}

	if filter.Offset > 0 {
		mods = append(mods, qm.Offset(filter.Offset))
	}

	return entity.UserReports(mods...).All(ctx, r.db)
}

func (r *repository) UpdateReport(ctx context.Context, report *entity.UserReport, tx *sql.Tx) error {
	exec := r.executor(tx)
	_, err := report.Update(ctx, exec, boil.Whitelist(
		entity.UserReportColumns.Status,
		entity.UserReportColumns.AutoAction,
		entity.UserReportColumns.ResolvedAt,
		entity.UserReportColumns.UpdatedAt,
	))

	return err
}

func (r *repository) InsertReportAction(ctx context.Context, action *entity.ReportAction, reviewerName string, tx *sql.Tx) error {
	exec := r.executor(tx)

	var reviewerID interface{}
	if action.ReviewerID.Valid {
		reviewerID = action.ReviewerID.String
	}

	var payload interface{}
	if action.ActionPayload.Valid {
		payload = action.ActionPayload.JSON
	}

	var notes interface{}
	if action.Notes.Valid {
		notes = action.Notes.String
	}

	err := exec.QueryRowContext(ctx, `
		INSERT INTO report_actions (report_id, reviewer_id, reviewer_name, action_type, action_payload, notes, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7)
		RETURNING id
	`, action.ReportID, reviewerID, reviewerName, action.ActionType, payload, notes, action.CreatedAt).Scan(&action.ID)
	if err != nil {
		return fmt.Errorf("insert report action: %w", err)
	}

	return nil
}

func (r *repository) LoadReviewerNamesForReport(ctx context.Context, reportID string) (map[string]string, error) {
	rows, err := r.db.QueryxContext(ctx, `
		SELECT id, reviewer_name FROM report_actions WHERE report_id = $1 AND reviewer_name IS NOT NULL
	`, reportID)
	if err != nil {
		return nil, fmt.Errorf("load reviewer names: %w", err)
	}

	defer func() { _ = rows.Close() }()

	out := make(map[string]string)

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		out[id] = name
	}

	return out, rows.Err()
}

func (r *repository) executor(tx *sql.Tx) boil.ContextExecutor {
	if tx != nil {
		return tx
	}

	return r.db
}

func stringSliceToInterface(values []string) []interface{} {
	args := make([]interface{}, len(values))
	for i, value := range values {
		args[i] = value
	}

	return args
}
