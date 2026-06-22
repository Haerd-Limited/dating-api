//go:generate mockgen -source=repository.go -destination=repository_mock.go -package=storage
package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/consent/domain"
	consentmapper "github.com/Haerd-Limited/dating-api/internal/consent/mapper"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type Repository interface {
	Insert(ctx context.Context, req domain.RecordRequest) error
	Revoke(ctx context.Context, userID, consentType string) error
	ListForUser(ctx context.Context, userID string) ([]*entity.UserConsent, error)
	GetMissingMandatory(ctx context.Context, userID string, types []string, versions map[string]string) ([]string, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Insert(ctx context.Context, req domain.RecordRequest) error {
	consent := consentmapper.RecordRequestToEntity(req)

	err := consent.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("insert user consent: %w", err)
	}

	return nil
}

func (r *repository) Revoke(ctx context.Context, userID, consentType string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_consents
		   SET revoked_at = $3
		 WHERE user_id = $1
		   AND consent_type = $2
		   AND revoked_at IS NULL
	`, userID, consentType, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("revoke user consent: %w", err)
	}

	return nil
}

func (r *repository) ListForUser(ctx context.Context, userID string) ([]*entity.UserConsent, error) {
	consents, err := entity.UserConsents(
		entity.UserConsentWhere.UserID.EQ(userID),
		qm.OrderBy("accepted_at DESC"),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("list user consents: %w", err)
	}

	return consents, nil
}

func (r *repository) GetMissingMandatory(ctx context.Context, userID string, types []string, versions map[string]string) ([]string, error) {
	if len(types) == 0 {
		return nil, nil
	}

	var valuesClause strings.Builder

	args := []any{userID}
	argIdx := 2

	for i, consentType := range types {
		version, ok := versions[consentType]
		if !ok {
			continue
		}

		if i > 0 {
			valuesClause.WriteString(", ")
		}

		valuesClause.WriteString(fmt.Sprintf("($%d::text, $%d::text)", argIdx, argIdx+1))

		args = append(args, consentType, version)
		argIdx += 2
	}

	query := fmt.Sprintf(`
		WITH required(consent_type, version) AS (
			VALUES %s
		)
		SELECT r.consent_type
		  FROM required r
		 WHERE NOT EXISTS (
			SELECT 1
			  FROM user_consents uc
			 WHERE uc.user_id = $1
			   AND uc.consent_type = r.consent_type
			   AND uc.version = r.version
			   AND uc.accepted = true
			   AND uc.revoked_at IS NULL
		 )
	`, valuesClause.String())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get missing mandatory consents: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var missing []string

	for rows.Next() {
		var consentType string
		if err := rows.Scan(&consentType); err != nil {
			return nil, fmt.Errorf("scan missing consent type: %w", err)
		}

		missing = append(missing, consentType)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate missing consents: %w", err)
	}

	return missing, nil
}
