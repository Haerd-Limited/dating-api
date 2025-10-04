package uow

import (
	"context"
	"database/sql"
)

type UoW interface {
	Begin(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit() error
	Rollback() error

	// Return the underlying *sql.Tx so repos can use sqlboiler/sqlx against it.
	Raw() *sql.Tx
}

type sqlUoW struct {
	db *sql.DB
}

func New(db *sql.DB) UoW {
	return &sqlUoW{db: db}
}

func (u *sqlUoW) Begin(ctx context.Context) (Tx, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	return &sqlTx{tx: tx}, nil
}

type sqlTx struct {
	tx *sql.Tx
}

func (t *sqlTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqlTx) Rollback() error {
	_ = t.tx.Rollback()
	return nil
}

func (t *sqlTx) Raw() *sql.Tx {
	return t.tx
}
