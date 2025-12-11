package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/byron1st/integridb/pkg/event"
	"github.com/lib/pq"
)

// PostgresAdapter implements event storage for PostgreSQL.
type PostgresAdapter struct {
	db *sql.DB
}

// NewPostgresAdapter creates a new PostgreSQL adapter.
func NewPostgresAdapter(db *sql.DB) *PostgresAdapter {
	return &PostgresAdapter{
		db: db,
	}
}

// Migrate creates the IntegriDB schema tables if they don't exist.
// This operation is idempotent and safe to run multiple times.
func (a *PostgresAdapter) Migrate(ctx context.Context) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Execute migration SQL statements
	for _, sql := range MigrationSQL() {
		if _, err := tx.ExecContext(ctx, sql); err != nil {
			return fmt.Errorf("execute migration: %w", err)
		}
	}

	// Record schema version
	const insertVersionSQL = `
		INSERT INTO integridb_schema_version (version)
		VALUES ($1)
		ON CONFLICT (version) DO NOTHING
	`
	if _, err := tx.ExecContext(ctx, insertVersionSQL, SchemaVersion); err != nil {
		return fmt.Errorf("record schema version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}

	return nil
}

// InsertEvent persists a new event to the event store.
func (a *PostgresAdapter) InsertEvent(ctx context.Context, event *event.Event) error {
	const insertSQL = `
		INSERT INTO integridb_events (
			id, table_name, row_id, event_type, version,
			payload, metadata, checksum, prev_checksum, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var metadataJSON []byte
	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	}

	_, err = a.db.ExecContext(ctx, insertSQL,
		event.ID,
		event.TableName,
		event.RowID,
		string(event.EventType),
		event.Version,
		payloadJSON,
		metadataJSON,
		event.Checksum,
		event.PrevChecksum,
		event.CreatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("event already exists: %w", err)
		}
		return fmt.Errorf("insert event: %w", err)
	}

	return nil
}

// GetEvents retrieves all events for a specific table and row, ordered by version.
func (a *PostgresAdapter) GetEvents(ctx context.Context, tableName, rowID string) ([]*event.Event, error) {
	const querySQL = `
		SELECT id, table_name, row_id, event_type, version,
		       payload, metadata, checksum, prev_checksum, created_at
		FROM integridb_events
		WHERE table_name = $1 AND row_id = $2
		ORDER BY version ASC
	`

	rows, err := a.db.QueryContext(ctx, querySQL, tableName, rowID)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []*event.Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return events, nil
}

// GetLastEvent retrieves the most recent event for a specific row.
func (a *PostgresAdapter) GetLastEvent(ctx context.Context, tableName, rowID string) (*event.Event, error) {
	const querySQL = `
		SELECT id, table_name, row_id, event_type, version,
		       payload, metadata, checksum, prev_checksum, created_at
		FROM integridb_events
		WHERE table_name = $1 AND row_id = $2
		ORDER BY version DESC
		LIMIT 1
	`

	row := a.db.QueryRowContext(ctx, querySQL, tableName, rowID)
	ev, err := scanEvent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, event.ErrRowNotFound{TableName: tableName, RowID: rowID}
		}
		return nil, fmt.Errorf("query last event: %w", err)
	}

	return ev, nil
}

// GetEventsByTable retrieves all events for a table, ordered by row_id and version.
func (a *PostgresAdapter) GetEventsByTable(ctx context.Context, tableName string) ([]*event.Event, error) {
	const querySQL = `
		SELECT id, table_name, row_id, event_type, version,
		       payload, metadata, checksum, prev_checksum, created_at
		FROM integridb_events
		WHERE table_name = $1
		ORDER BY row_id, version ASC
	`

	rows, err := a.db.QueryContext(ctx, querySQL, tableName)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []*event.Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return events, nil
}

// GetEventByID retrieves a specific event by its ID.
func (a *PostgresAdapter) GetEventByID(ctx context.Context, eventID string) (*event.Event, error) {
	const querySQL = `
		SELECT id, table_name, row_id, event_type, version,
		       payload, metadata, checksum, prev_checksum, created_at
		FROM integridb_events
		WHERE id = $1
	`

	row := a.db.QueryRowContext(ctx, querySQL, eventID)
	ev, err := scanEvent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, event.ErrEventNotFound{EventID: eventID}
		}
		return nil, fmt.Errorf("query event: %w", err)
	}

	return ev, nil
}

// scanEvent is a helper interface for scanning both sql.Row and sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// scanEvent scans a database row into an Event struct.
func scanEvent(s scanner) (*event.Event, error) {
	var ev event.Event
	var eventTypeStr string
	var payloadJSON, metadataJSON []byte

	err := s.Scan(
		&ev.ID,
		&ev.TableName,
		&ev.RowID,
		&eventTypeStr,
		&ev.Version,
		&payloadJSON,
		&metadataJSON,
		&ev.Checksum,
		&ev.PrevChecksum,
		&ev.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	ev.EventType = event.EventType(eventTypeStr)

	if err := json.Unmarshal(payloadJSON, &ev.Payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &ev.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return &ev, nil
}

// TrackedTableInfo contains metadata about a tracked table.
type TrackedTableInfo struct {
	TableName         string
	PrimaryKeyColumns []string
	LastEventID       *string
	LastEventAt       *time.Time
}

// RegisterTrackedTable records a table in the tracked tables registry.
func (a *PostgresAdapter) RegisterTrackedTable(ctx context.Context, tableName string, primaryKey []string) error {
	const insertSQL = `
		INSERT INTO integridb_tracked_tables (table_name, primary_key_columns)
		VALUES ($1, $2)
		ON CONFLICT (table_name) DO UPDATE
		SET primary_key_columns = EXCLUDED.primary_key_columns
	`

	_, err := a.db.ExecContext(ctx, insertSQL, tableName, pq.Array(primaryKey))
	if err != nil {
		return fmt.Errorf("register tracked table: %w", err)
	}

	return nil
}

// UnregisterTrackedTable removes a table from the tracked tables registry.
func (a *PostgresAdapter) UnregisterTrackedTable(ctx context.Context, tableName string) error {
	const deleteSQL = `
		DELETE FROM integridb_tracked_tables
		WHERE table_name = $1
	`

	_, err := a.db.ExecContext(ctx, deleteSQL, tableName)
	if err != nil {
		return fmt.Errorf("unregister tracked table: %w", err)
	}

	return nil
}

// GetTrackedTables retrieves all tracked tables from the registry.
func (a *PostgresAdapter) GetTrackedTables(ctx context.Context) ([]TrackedTableInfo, error) {
	const querySQL = `
		SELECT table_name, primary_key_columns, last_event_id, last_event_at
		FROM integridb_tracked_tables
		ORDER BY table_name
	`

	rows, err := a.db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, fmt.Errorf("query tracked tables: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []TrackedTableInfo
	for rows.Next() {
		var table TrackedTableInfo
		var pkCols []string

		err := rows.Scan(
			&table.TableName,
			pq.Array(&pkCols),
			&table.LastEventID,
			&table.LastEventAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tracked table: %w", err)
		}

		table.PrimaryKeyColumns = pkCols
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tracked tables: %w", err)
	}

	return tables, nil
}

// GetTrackedTable retrieves a specific tracked table from the registry.
func (a *PostgresAdapter) GetTrackedTable(ctx context.Context, tableName string) (*TrackedTableInfo, error) {
	const querySQL = `
		SELECT table_name, primary_key_columns, last_event_id, last_event_at
		FROM integridb_tracked_tables
		WHERE table_name = $1
	`

	row := a.db.QueryRowContext(ctx, querySQL, tableName)

	var table TrackedTableInfo
	var pkCols []string

	err := row.Scan(
		&table.TableName,
		pq.Array(&pkCols),
		&table.LastEventID,
		&table.LastEventAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Table not tracked
		}
		return nil, fmt.Errorf("query tracked table: %w", err)
	}

	table.PrimaryKeyColumns = pkCols
	return &table, nil
}

// UpdateLastEvent updates the last event information for a tracked table.
func (a *PostgresAdapter) UpdateLastEvent(ctx context.Context, tableName, eventID string, eventTime time.Time) error {
	const updateSQL = `
		UPDATE integridb_tracked_tables
		SET last_event_id = $2, last_event_at = $3
		WHERE table_name = $1
	`

	_, err := a.db.ExecContext(ctx, updateSQL, tableName, eventID, eventTime)
	if err != nil {
		return fmt.Errorf("update last event: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (a *PostgresAdapter) Close() error {
	return a.db.Close()
}
