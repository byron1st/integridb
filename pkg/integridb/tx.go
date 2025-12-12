package integridb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/byron1st/integridb/pkg/event"
	"github.com/byron1st/integridb/pkg/query"
)

// pendingEvent represents an event to be persisted on transaction commit.
type pendingEvent struct {
	tableName string
	rowID     string
	eventType event.EventType
	payload   event.EventPayload
	metadata  map[string]any
}

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
	pendingEvents []*pendingEvent

	// committed tracks whether the transaction has been committed
	committed bool

	// rolledBack tracks whether the transaction has been rolled back
	rolledBack bool
}

// Exec executes a query without returning any rows.
// For tracked tables, mutations are captured as events and buffered until Commit.
func (t *Tx) Exec(queryStr string, args ...any) (sql.Result, error) {
	return t.ExecContext(context.Background(), queryStr, args...)
}

// ExecContext executes a query without returning any rows, with context support.
// For tracked tables, mutations are captured as events and buffered until Commit.
func (t *Tx) ExecContext(ctx context.Context, queryStr string, args ...any) (sql.Result, error) {
	// Parse the SQL query
	parsedQuery, err := query.Parse(queryStr)
	if err != nil {
		// If parsing fails, fall back to direct execution
		return t.tx.ExecContext(ctx, queryStr, args...)
	}

	// Check if this is a mutation on a tracked table
	if !parsedQuery.IsMutation() {
		return t.tx.ExecContext(ctx, queryStr, args...)
	}

	tableConfig := t.db.config.FindTableConfig(parsedQuery.TableName)
	if tableConfig == nil {
		// Table is not tracked, execute directly
		return t.tx.ExecContext(ctx, queryStr, args...)
	}

	// Route to appropriate handler based on query type
	switch parsedQuery.Type {
	case query.QueryTypeInsert:
		return t.execInsert(ctx, parsedQuery, tableConfig, queryStr, args...)
	case query.QueryTypeUpdate:
		return t.execUpdate(ctx, parsedQuery, tableConfig, queryStr, args...)
	case query.QueryTypeDelete:
		return t.execDelete(ctx, parsedQuery, tableConfig, queryStr, args...)
	default:
		return t.tx.ExecContext(ctx, queryStr, args...)
	}
}

// Query executes a query that returns rows.
// SELECT queries are passed through directly to the underlying transaction.
func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryContext executes a query that returns rows, with context support.
// SELECT queries are passed through directly to the underlying transaction.
func (t *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row, with context support.
func (t *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
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

// execInsert handles INSERT operations within a transaction.
func (t *Tx) execInsert(ctx context.Context, parsedQuery *query.ParsedQuery, tableConfig *TableConfig, queryStr string, args ...any) (sql.Result, error) {
	// Execute the INSERT statement
	result, err := t.tx.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute insert: %w", err)
	}

	// Extract row ID from result
	rowID, err := extractRowID(result, tableConfig.PrimaryKey, nil)
	if err != nil {
		// If we can't extract the row ID automatically, skip event capture
		return result, nil
	}

	// Capture after state using transaction context
	afterState, err := t.captureAfterState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
	if err != nil {
		return nil, fmt.Errorf("capture after state: %w", err)
	}

	// Extract metadata from context and MetadataFunc
	metadata := t.db.extractMetadata(ctx)

	// Create event payload
	payload := event.EventPayload{
		Before:         nil, // INSERT has no before state
		After:          afterState,
		ChangedColumns: nil, // INSERT doesn't have changed columns
	}

	// Buffer event for commit
	t.pendingEvents = append(t.pendingEvents, &pendingEvent{
		tableName: parsedQuery.TableName,
		rowID:     rowID,
		eventType: event.EventTypeInsert,
		payload:   payload,
		metadata:  metadata,
	})

	return result, nil
}

// execUpdate handles UPDATE operations within a transaction.
func (t *Tx) execUpdate(ctx context.Context, parsedQuery *query.ParsedQuery, tableConfig *TableConfig, queryStr string, args ...any) (sql.Result, error) {
	// First, identify all affected rows before the update
	affectedRowIDs, err := t.captureAffectedRowIDs(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, parsedQuery.WhereClause, args)
	if err != nil {
		return nil, fmt.Errorf("capture affected rows: %w", err)
	}

	// Capture before state for all affected rows
	beforeStates := make(map[string]map[string]any)
	for _, rowID := range affectedRowIDs {
		beforeState, err := t.captureBeforeState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
		if err != nil {
			return nil, fmt.Errorf("capture before state for row %q: %w", rowID, err)
		}
		beforeStates[rowID] = beforeState
	}

	// Execute the UPDATE statement
	result, err := t.tx.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute update: %w", err)
	}

	// Capture after state for all affected rows and buffer events
	for _, rowID := range affectedRowIDs {
		afterState, err := t.captureAfterState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
		if err != nil {
			return nil, fmt.Errorf("capture after state for row %q: %w", rowID, err)
		}

		// Detect changed columns
		changedCols := detectChangedColumns(beforeStates[rowID], afterState)

		// Extract metadata from context and MetadataFunc
		metadata := t.db.extractMetadata(ctx)

		// Create event payload
		payload := event.EventPayload{
			Before:         beforeStates[rowID],
			After:          afterState,
			ChangedColumns: changedCols,
		}

		// Buffer event for commit
		t.pendingEvents = append(t.pendingEvents, &pendingEvent{
			tableName: parsedQuery.TableName,
			rowID:     rowID,
			eventType: event.EventTypeUpdate,
			payload:   payload,
			metadata:  metadata,
		})
	}

	return result, nil
}

// execDelete handles DELETE operations within a transaction.
func (t *Tx) execDelete(ctx context.Context, parsedQuery *query.ParsedQuery, tableConfig *TableConfig, queryStr string, args ...any) (sql.Result, error) {
	// First, identify all affected rows before the delete
	affectedRowIDs, err := t.captureAffectedRowIDs(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, parsedQuery.WhereClause, args)
	if err != nil {
		return nil, fmt.Errorf("capture affected rows: %w", err)
	}

	// Capture before state for all affected rows
	beforeStates := make(map[string]map[string]any)
	for _, rowID := range affectedRowIDs {
		beforeState, err := t.captureBeforeState(ctx, parsedQuery.TableName, tableConfig.PrimaryKey, rowID)
		if err != nil {
			return nil, fmt.Errorf("capture before state for row %q: %w", rowID, err)
		}
		beforeStates[rowID] = beforeState
	}

	// Execute the DELETE statement
	result, err := t.tx.ExecContext(ctx, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute delete: %w", err)
	}

	// Buffer events for all deleted rows
	for _, rowID := range affectedRowIDs {
		// Extract metadata from context and MetadataFunc
		metadata := t.db.extractMetadata(ctx)

		// Create event payload
		payload := event.EventPayload{
			Before:         beforeStates[rowID],
			After:          nil, // DELETE has no after state
			ChangedColumns: nil, // DELETE doesn't have changed columns
		}

		// Buffer event for commit
		t.pendingEvents = append(t.pendingEvents, &pendingEvent{
			tableName: parsedQuery.TableName,
			rowID:     rowID,
			eventType: event.EventTypeDelete,
			payload:   payload,
			metadata:  metadata,
		})
	}

	return result, nil
}

// captureBeforeState queries the current state of a row before a mutation within the transaction.
func (t *Tx) captureBeforeState(ctx context.Context, tableName string, primaryKeyCols []string, rowID string) (map[string]any, error) {
	columns, err := t.db.getColumns(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	// Build SELECT query
	selectCols := joinColumns(columns)
	whereClause := buildWhereClause(primaryKeyCols, 1)
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectCols, tableName, whereClause)

	// Parse rowID into values for WHERE clause
	pkValues, err := parseRowID(rowID, len(primaryKeyCols))
	if err != nil {
		return nil, fmt.Errorf("parse row ID: %w", err)
	}

	// Execute query within transaction
	row := t.tx.QueryRowContext(ctx, querySQL, pkValues...)

	// Scan into map
	state, err := scanRowIntoMap(row, columns)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("row not found: table=%q rowID=%q", tableName, rowID)
		}
		return nil, fmt.Errorf("scan row state: %w", err)
	}

	return state, nil
}

// captureAfterState queries the current state of a row after a mutation within the transaction.
func (t *Tx) captureAfterState(ctx context.Context, tableName string, primaryKeyCols []string, rowID string) (map[string]any, error) {
	columns, err := t.db.getColumns(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	// Build SELECT query
	selectCols := joinColumns(columns)
	whereClause := buildWhereClause(primaryKeyCols, 1)
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectCols, tableName, whereClause)

	// Parse rowID into values for WHERE clause
	pkValues, err := parseRowID(rowID, len(primaryKeyCols))
	if err != nil {
		return nil, fmt.Errorf("parse row ID: %w", err)
	}

	// Execute query within transaction
	row := t.tx.QueryRowContext(ctx, querySQL, pkValues...)

	// Scan into map
	state, err := scanRowIntoMap(row, columns)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("row not found after mutation: table=%q rowID=%q", tableName, rowID)
		}
		return nil, fmt.Errorf("scan row state: %w", err)
	}

	return state, nil
}

// captureAffectedRowIDs identifies all row IDs affected by an UPDATE or DELETE within the transaction.
func (t *Tx) captureAffectedRowIDs(ctx context.Context, tableName string, primaryKeyCols []string, whereClause string, args []any) ([]string, error) {
	if len(primaryKeyCols) == 0 {
		return nil, fmt.Errorf("no primary key columns specified")
	}

	// Build SELECT query to get affected row IDs
	pkCols := joinColumns(primaryKeyCols)
	querySQL := fmt.Sprintf("SELECT %s FROM %s", pkCols, tableName)

	// Add WHERE clause if present
	if whereClause != "" {
		querySQL += " WHERE " + whereClause
	}

	rows, err := t.tx.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query affected rows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var rowIDs []string
	for rows.Next() {
		// Create scan destinations for primary key columns
		scanDest := make([]any, len(primaryKeyCols))
		for i := range scanDest {
			scanDest[i] = new(any)
		}

		if err := rows.Scan(scanDest...); err != nil {
			return nil, fmt.Errorf("scan row ID: %w", err)
		}

		// Build row ID from primary key values
		rowID := buildRowID(scanDest)
		rowIDs = append(rowIDs, rowID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate affected rows: %w", err)
	}

	return rowIDs, nil
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

	// Persist all pending events in order
	ctx := t.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	for _, pe := range t.pendingEvents {
		_, err := t.db.store.AppendEvent(ctx, pe.tableName, pe.rowID, pe.eventType, pe.payload, pe.metadata)
		if err != nil {
			// If event persistence fails, rollback the transaction
			_ = t.tx.Rollback()
			return fmt.Errorf("append event during commit: %w", err)
		}
	}

	// Commit the underlying transaction
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

// joinColumns joins column names with commas for SQL queries.
func joinColumns(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}
