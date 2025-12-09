package integridb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/byron1st/integridb/pkg/adapter/postgres"
)

// DB wraps a standard database/sql.DB connection with event sourcing capabilities.
// It provides the same API as sql.DB while transparently capturing mutations as events.
type DB struct {
	// db is the underlying database connection
	db *sql.DB

	// config holds the IntegriDB configuration
	config *Config

	// adapter handles event storage operations
	adapter *postgres.PostgresAdapter
}

// Open creates a new IntegriDB database connection.
// The config must specify the driver, DSN, and tracked tables.
func Open(config *Config) (*DB, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Open the underlying database connection
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	// Create the adapter (currently only PostgreSQL is supported)
	adapter := postgres.NewPostgresAdapter(db)

	integriDB := &DB{
		db:      db,
		config:  config,
		adapter: adapter,
	}

	return integriDB, nil
}

// Migrate creates the IntegriDB schema (events table, tracked tables registry).
// This should be called once during application initialization.
// The operation is idempotent and safe to run multiple times.
func (d *DB) Migrate(ctx context.Context) error {
	if err := d.adapter.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	// Register tracked tables in the registry
	for _, table := range d.config.TrackedTables {
		if err := d.adapter.RegisterTrackedTable(ctx, table.Name, table.PrimaryKey); err != nil {
			return fmt.Errorf("register tracked table %q: %w", table.Name, err)
		}
	}

	return nil
}

// Ping verifies the database connection is alive.
func (d *DB) Ping() error {
	return d.db.Ping()
}

// PingContext verifies the database connection is alive with context.
func (d *DB) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// Exec executes a query without returning any rows.
// For tracked tables, this will capture INSERT/UPDATE/DELETE operations as events.
// For Stage 1, this is a pass-through to the underlying database.
func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// ExecContext executes a query without returning any rows, with context support.
// For tracked tables, this will capture INSERT/UPDATE/DELETE operations as events.
// For Stage 1, this is a pass-through to the underlying database.
func (d *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// TODO: Stage 2 will add SQL parsing and event capture here
	return d.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows.
// SELECT queries are passed through directly to the underlying database.
func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryContext executes a query that returns rows, with context support.
// SELECT queries are passed through directly to the underlying database.
func (d *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row, with context support.
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// Prepare creates a prepared statement for later queries or executions.
func (d *DB) Prepare(query string) (*sql.Stmt, error) {
	// TODO: Stage 3 will wrap sql.Stmt to capture events
	return d.db.Prepare(query)
}

// PrepareContext creates a prepared statement for later queries or executions, with context support.
func (d *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// TODO: Stage 3 will wrap sql.Stmt to capture events
	return d.db.PrepareContext(ctx, query)
}

// Begin starts a new transaction.
// Events will be buffered and persisted atomically on Commit.
func (d *DB) Begin() (*Tx, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:            tx,
		db:            d,
		pendingEvents: make([]*Event, 0),
	}, nil
}

// BeginTx starts a new transaction with context and options.
// Events will be buffered and persisted atomically on Commit.
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:            tx,
		db:            d,
		ctx:           ctx,
		pendingEvents: make([]*Event, 0),
	}, nil
}

// Stats returns database statistics.
func (d *DB) Stats() sql.DBStats {
	return d.db.Stats()
}

// SetMaxIdleConns sets the maximum number of idle connections in the pool.
func (d *DB) SetMaxIdleConns(n int) {
	d.db.SetMaxIdleConns(n)
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
func (d *DB) SetMaxOpenConns(n int) {
	d.db.SetMaxOpenConns(n)
}

// UnderlyingDB returns the underlying *sql.DB for advanced use cases.
// Use with caution: operations on the underlying DB bypass event capture.
func (d *DB) UnderlyingDB() *sql.DB {
	return d.db
}

// Config returns the IntegriDB configuration.
func (d *DB) Config() *Config {
	return d.config
}
