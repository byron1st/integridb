package integridb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/byron1st/integridb/pkg/adapter/postgres"
	"github.com/byron1st/integridb/pkg/event"
	"github.com/byron1st/integridb/pkg/query"
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

	// store manages event persistence and retrieval
	store *event.Store

	// colCache caches column metadata for tables
	colCache *columnCache
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

	// Create the event store
	store := event.NewStore(adapter)

	integriDB := &DB{
		db:       db,
		config:   config,
		adapter:  adapter,
		store:    store,
		colCache: newColumnCache(),
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
func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.ExecContext(context.Background(), query, args...)
}

// ExecContext executes a query without returning any rows, with context support.
// For tracked tables, this will capture INSERT/UPDATE/DELETE operations as events.
func (d *DB) ExecContext(ctx context.Context, queryStr string, args ...any) (sql.Result, error) {
	// Parse the SQL query
	parsedQuery, err := query.Parse(queryStr)
	if err != nil {
		// If parsing fails, fall back to direct execution
		return d.db.ExecContext(ctx, queryStr, args...)
	}

	// Check if this is a mutation on a tracked table
	if !parsedQuery.IsMutation() {
		return d.db.ExecContext(ctx, queryStr, args...)
	}

	tableConfig := d.config.FindTableConfig(parsedQuery.TableName)
	if tableConfig == nil {
		// Table is not tracked, execute directly
		return d.db.ExecContext(ctx, queryStr, args...)
	}

	// Route to appropriate handler based on query type
	switch parsedQuery.Type {
	case query.QueryTypeInsert:
		return d.execInsert(ctx, parsedQuery, tableConfig, queryStr, args...)
	case query.QueryTypeUpdate:
		return d.execUpdate(ctx, parsedQuery, tableConfig, queryStr, args...)
	case query.QueryTypeDelete:
		return d.execDelete(ctx, parsedQuery, tableConfig, queryStr, args...)
	default:
		return d.db.ExecContext(ctx, queryStr, args...)
	}
}

// Query executes a query that returns rows.
// SELECT queries are passed through directly to the underlying database.
func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryContext executes a query that returns rows, with context support.
// SELECT queries are passed through directly to the underlying database.
func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row, with context support.
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
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

// execInsert handles INSERT operations with event capture.
func (d *DB) execInsert(ctx context.Context, parsedQuery *query.ParsedQuery, tableConfig *TableConfig, queryStr string, args ...any) (sql.Result, error) {
	// Execute the INSERT statement
	result, err := d.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute insert: %w", err)
	}

	// Extract row ID from result
	// For PostgreSQL, we need to use RETURNING clause or extract from args
	// For now, try to use LastInsertId as a fallback
	rowID, err := extractRowID(result, tableConfig.PrimaryKey, nil)
	if err != nil {
		// If we can't extract the row ID automatically, we need RETURNING clause
		// For MVP, we'll require users to use RETURNING for tracked tables
		return result, nil // Return result but skip event capture
	}

	// Capture after state
	afterState, err := d.captureAfterState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
	if err != nil {
		return nil, fmt.Errorf("capture after state: %w", err)
	}

	// Extract metadata from context if configured
	var metadata map[string]any
	if d.config.MetadataFunc != nil {
		metadata = d.config.MetadataFunc(ctx)
	}

	// Create event payload
	payload := event.EventPayload{
		Before:         nil, // INSERT has no before state
		After:          afterState,
		ChangedColumns: nil, // INSERT doesn't have changed columns
	}

	// Append event to store
	_, err = d.store.AppendEvent(ctx, parsedQuery.TableName, rowID, event.EventTypeInsert, payload, metadata)
	if err != nil {
		return nil, fmt.Errorf("append insert event: %w", err)
	}

	return result, nil
}

// execUpdate handles UPDATE operations with event capture.
func (d *DB) execUpdate(ctx context.Context, parsedQuery *query.ParsedQuery, tableConfig *TableConfig, queryStr string, args ...any) (sql.Result, error) {
	// First, identify all affected rows before the update
	affectedRowIDs, err := d.captureAffectedRowIDs(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, parsedQuery.WhereClause, args)
	if err != nil {
		return nil, fmt.Errorf("capture affected rows: %w", err)
	}

	// Capture before state for all affected rows
	beforeStates := make(map[string]map[string]any)
	for _, rowID := range affectedRowIDs {
		beforeState, err := d.captureBeforeState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
		if err != nil {
			return nil, fmt.Errorf("capture before state for row %q: %w", rowID, err)
		}
		beforeStates[rowID] = beforeState
	}

	// Execute the UPDATE statement
	result, err := d.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute update: %w", err)
	}

	// Capture after state for all affected rows and create events
	for _, rowID := range affectedRowIDs {
		afterState, err := d.captureAfterState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
		if err != nil {
			return nil, fmt.Errorf("capture after state for row %q: %w", rowID, err)
		}

		// Detect changed columns
		changedCols := detectChangedColumns(beforeStates[rowID], afterState)

		// Extract metadata from context if configured
		var metadata map[string]any
		if d.config.MetadataFunc != nil {
			metadata = d.config.MetadataFunc(ctx)
		}

		// Create event payload
		payload := event.EventPayload{
			Before:         beforeStates[rowID],
			After:          afterState,
			ChangedColumns: changedCols,
		}

		// Append event to store
		_, err = d.store.AppendEvent(ctx, parsedQuery.TableName, rowID, event.EventTypeUpdate, payload, metadata)
		if err != nil {
			return nil, fmt.Errorf("append update event for row %q: %w", rowID, err)
		}
	}

	return result, nil
}

// execDelete handles DELETE operations with event capture.
func (d *DB) execDelete(ctx context.Context, parsedQuery *query.ParsedQuery, tableConfig *TableConfig, queryStr string, args ...any) (sql.Result, error) {
	// First, identify all affected rows before the delete
	affectedRowIDs, err := d.captureAffectedRowIDs(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, parsedQuery.WhereClause, args)
	if err != nil {
		return nil, fmt.Errorf("capture affected rows: %w", err)
	}

	// Capture before state for all affected rows
	beforeStates := make(map[string]map[string]any)
	for _, rowID := range affectedRowIDs {
		beforeState, err := d.captureBeforeState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
		if err != nil {
			return nil, fmt.Errorf("capture before state for row %q: %w", rowID, err)
		}
		beforeStates[rowID] = beforeState
	}

	// Execute the DELETE statement
	result, err := d.db.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute delete: %w", err)
	}

	// Create events for all deleted rows
	for _, rowID := range affectedRowIDs {
		// Extract metadata from context if configured
		var metadata map[string]any
		if d.config.MetadataFunc != nil {
			metadata = d.config.MetadataFunc(ctx)
		}

		// Create event payload
		payload := event.EventPayload{
			Before:         beforeStates[rowID],
			After:          nil, // DELETE has no after state
			ChangedColumns: nil, // DELETE doesn't have changed columns
		}

		// Append event to store
		_, err = d.store.AppendEvent(ctx, parsedQuery.TableName, rowID, event.EventTypeDelete, payload, metadata)
		if err != nil {
			return nil, fmt.Errorf("append delete event for row %q: %w", rowID, err)
		}
	}

	return result, nil
}
