package storage

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"

	"github.com/Haerd-Limited/dating-api/internal/broadcast/domain"
	"github.com/Haerd-Limited/dating-api/internal/entity"
)

type Repository interface {
	InsertBroadcastLog(ctx context.Context, broadcast *entity.SMSBroadcast) error
	GetContactedUserIDs(ctx context.Context) (map[string]bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) InsertBroadcastLog(ctx context.Context, broadcast *entity.SMSBroadcast) error {
	if err := broadcast.Insert(ctx, r.db, boil.Infer()); err != nil {
		return fmt.Errorf("insert sms broadcast log: %w", err)
	}

	return nil
}

func (r *repository) GetContactedUserIDs(ctx context.Context) (map[string]bool, error) {
	broadcasts, err := entity.SMSBroadcasts(
		entity.SMSBroadcastWhere.Status.EQ(domain.BroadcastStatusSent),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("get contacted user ids: %w", err)
	}

	contacted := make(map[string]bool, len(broadcasts))
	for _, broadcast := range broadcasts {
		contacted[broadcast.UserID] = true
	}

	return contacted, nil
}
