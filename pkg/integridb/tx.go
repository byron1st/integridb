package integridb

import (
	"context"
	"database/sql"
)

// Tx wraps a standard database/sql.Tx transaction with event buffering.
// Events generated during the transaction are buffered and persisted atomically on Commit.
type Tx struct {
	// tx is the underlying database transaction
	tx *sql.Tx

	// db is the parent IntegriDB connection
	db *DB

	// ctx is the context for this transaction (optional)
	ctx context.Context

	// pendingEvents buffers events to be persisted on Commit
	pendingEvents []*Event

	// committed tracks whether the transaction has been committed
	committed bool

	// rolledBack tracks whether the transaction has been rolled back
	rolledBack bool
}

// Exec executes a query without returning any rows.
// For tracked tables, mutations are captured as events and buffered until Commit.
// For Stage 1, this is a pass-through to the underlying transaction.
func (t *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	// TODO: Stage 3 will add event capture here
	return t.tx.Exec(query, args...)
}

// ExecContext executes a query without returning any rows, with context support.
// For tracked tables, mutations are captured as events and buffered until Commit.
// For Stage 1, this is a pass-through to the underlying transaction.
func (t *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// TODO: Stage 3 will add event capture here
	return t.tx.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows.
// SELECT queries are passed through directly to the underlying transaction.
func (t *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryContext executes a query that returns rows, with context support.
// SELECT queries are passed through directly to the underlying transaction.
func (t *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (t *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row, with context support.
func (t *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// Prepare creates a prepared statement within this transaction.
func (t *Tx) Prepare(query string) (*sql.Stmt, error) {
	// TODO: Stage 3 will wrap sql.Stmt to capture events
	return t.tx.Prepare(query)
}

// PrepareContext creates a prepared statement within this transaction, with context support.
func (t *Tx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// TODO: Stage 3 will wrap sql.Stmt to capture events
	return t.tx.PrepareContext(ctx, query)
}

// Stmt returns a transaction-specific prepared statement from an existing statement.
func (t *Tx) Stmt(stmt *sql.Stmt) *sql.Stmt {
	return t.tx.Stmt(stmt)
}

// StmtContext returns a transaction-specific prepared statement from an existing statement, with context support.
func (t *Tx) StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt {
	return t.tx.StmtContext(ctx, stmt)
}

// Commit commits the transaction.
// Pending events are persisted to the event store atomically with the user data.
func (t *Tx) Commit() error {
	if t.rolledBack {
		return sql.ErrTxDone
	}
	if t.committed {
		return sql.ErrTxDone
	}

	// TODO: Stage 4 will persist pending events here
	// For now, just commit the underlying transaction

	if err := t.tx.Commit(); err != nil {
		return err
	}

	t.committed = true
	return nil
}

// Rollback aborts the transaction.
// All pending events are discarded.
func (t *Tx) Rollback() error {
	if t.committed {
		return sql.ErrTxDone
	}
	if t.rolledBack {
		return sql.ErrTxDone
	}

	// Discard pending events
	t.pendingEvents = nil

	if err := t.tx.Rollback(); err != nil {
		return err
	}

	t.rolledBack = true
	return nil
}

// PendingEventCount returns the number of events buffered in this transaction.
// This is useful for debugging and testing.
func (t *Tx) PendingEventCount() int {
	return len(t.pendingEvents)
}
