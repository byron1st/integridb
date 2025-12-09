package postgres

const (
	// SchemaVersion is the current schema version for migration tracking.
	SchemaVersion = 1

	// CreateEventsTableSQL creates the integridb_events table.
	CreateEventsTableSQL = `
CREATE TABLE IF NOT EXISTS integridb_events (
	id TEXT PRIMARY KEY,
	table_name TEXT NOT NULL,
	row_id TEXT NOT NULL,
	event_type TEXT NOT NULL,
	version BIGINT NOT NULL,
	payload JSONB NOT NULL,
	metadata JSONB,
	checksum TEXT NOT NULL,
	prev_checksum TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	UNIQUE(table_name, row_id, version)
);
`

	// CreateEventsIndexSQL creates indexes for efficient event lookup.
	CreateEventsIndexSQL = `
CREATE INDEX IF NOT EXISTS idx_events_table_row ON integridb_events(table_name, row_id);
CREATE INDEX IF NOT EXISTS idx_events_table_row_version ON integridb_events(table_name, row_id, version);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON integridb_events(created_at);
CREATE INDEX IF NOT EXISTS idx_events_checksum ON integridb_events(checksum);
`

	// CreateTrackedTablesSQL creates the integridb_tracked_tables registry.
	CreateTrackedTablesSQL = `
CREATE TABLE IF NOT EXISTS integridb_tracked_tables (
	table_name TEXT PRIMARY KEY,
	primary_key_columns TEXT[] NOT NULL,
	tracked_since TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	last_event_id TEXT,
	last_event_at TIMESTAMP WITH TIME ZONE
);
`

	// CreateSchemaVersionSQL creates the schema version tracking table.
	CreateSchemaVersionSQL = `
CREATE TABLE IF NOT EXISTS integridb_schema_version (
	version INTEGER PRIMARY KEY,
	applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`
)

// MigrationSQL returns the SQL statements to create the IntegriDB schema.
func MigrationSQL() []string {
	return []string{
		CreateSchemaVersionSQL,
		CreateEventsTableSQL,
		CreateEventsIndexSQL,
		CreateTrackedTablesSQL,
	}
}
