package pgx_transactor

import (
	"context"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type pgxTxKey struct{}

// injects pgx.Tx into context
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, pgxTxKey{}, tx)
}

// retrieves pgx.Tx from context
func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(pgxTxKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

// PgxTransactor represents pgx transactor behavior
type PgxTransactor interface {
	WithinTransaction(ctx context.Context, txFn func(context.Context) error) error
	WithinTransactionWithOptions(ctx context.Context, txFn func(context.Context) error, opts pgx.TxOptions) error
}

type pgxTransactor struct {
	pool *pgxpool.Pool
}

// NewPgxTransactor builds new PgxTransactor
func NewPgxTransactor(p *pgxpool.Pool) PgxTransactor {
	return &pgxTransactor{pool: p}
}

// WithinTransaction runs WithinTransactionWithOptions with default tx options
func (t *pgxTransactor) WithinTransaction(ctx context.Context, txFunc func(context.Context) error) error {
	return t.WithinTransactionWithOptions(ctx, txFunc, pgx.TxOptions{})
}

// WithinTransactionWithOptions runs logic within transaction passing context with pgx.Tx injected into it,
// so you can retrieve it via PgxWithinTransactionRunner function Runner
func (t *pgxTransactor) WithinTransactionWithOptions(ctx context.Context, txFunc func(context.Context) error, opts pgx.TxOptions) error {
	return t.pool.BeginTxFunc(ctx, opts, func(tx pgx.Tx) error {
		return txFunc(injectTx(ctx, tx))
	})
}

// PgxQueryRunner represents query runner behavior
type PgxQueryRunner interface {
	pgxtype.Querier
	Begin(context.Context) (pgx.Tx, error)
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults
	CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
}

// PgxWithinTransactionRunner represents query runner retriever for pgx
type PgxWithinTransactionRunner interface {
	Runner(ctx context.Context) PgxQueryRunner
}

type pgxWithinTransactionRunner struct {
	pool *pgxpool.Pool
}

// NewPgxWithinTransactionRunner builds new PgxWithinTransactionRunner
func NewPgxWithinTransactionRunner(p *pgxpool.Pool) PgxWithinTransactionRunner {
	return &pgxWithinTransactionRunner{pool: p}
}

// Runner extracts query runner from context, if pgx.Tx is injected into context it is returned and pgxpool.Pool otherwise
func (e *pgxWithinTransactionRunner) Runner(ctx context.Context) PgxQueryRunner {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}
	return e.pool
}
