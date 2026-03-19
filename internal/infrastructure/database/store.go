package database

import (
	"context"
	"fmt"

	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
	*sqlc.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *Store) WithTx(tx pgx.Tx) *Store {
	return &Store{
		pool:    s.pool,
		Queries: s.Queries.WithTx(tx),
	}
}

func (s *Store) InTx(ctx context.Context, fn func(store *Store) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := fn(s.WithTx(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}
