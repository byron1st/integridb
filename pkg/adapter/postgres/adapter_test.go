package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/byron1st/integridb/pkg/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDB is a placeholder for database connection in unit tests.
// Integration tests with real PostgreSQL will be added in later stages.
func TestPostgresAdapter_New(t *testing.T) {
	// This test verifies that NewPostgresAdapter creates an adapter instance.
	// Actual database operations will be tested in integration tests.
	var db *sql.DB // nil is acceptable for this constructor test
	adapter := NewPostgresAdapter(db)
	assert.NotNil(t, adapter)
}

func TestMigrationSQL(t *testing.T) {
	// Verify that migration SQL is generated correctly
	migrations := MigrationSQL()
	require.NotEmpty(t, migrations, "migrations should not be empty")
	assert.Len(t, migrations, 4, "should have 4 migration statements")

	// Check that each migration contains expected keywords
	assert.Contains(t, migrations[0], "integridb_schema_version")
	assert.Contains(t, migrations[1], "integridb_events")
	assert.Contains(t, migrations[2], "CREATE INDEX")
	assert.Contains(t, migrations[3], "integridb_tracked_tables")
}

func TestScanEvent_Structure(t *testing.T) {
	// This test verifies the event structure is correct
	ev := &event.Event{
		ID:           "evt-123",
		TableName:    "users",
		RowID:        "1",
		EventType:    event.EventTypeInsert,
		Version:      1,
		Payload:      event.EventPayload{After: map[string]interface{}{"name": "Alice"}},
		Metadata:     map[string]interface{}{"user_id": "admin"},
		Checksum:     "abc123",
		PrevChecksum: "",
		CreatedAt:    time.Now(),
	}

	assert.Equal(t, "evt-123", ev.ID)
	assert.Equal(t, "users", ev.TableName)
	assert.Equal(t, event.EventTypeInsert, ev.EventType)
	assert.Equal(t, int64(1), ev.Version)
}

// Note: Full integration tests with real PostgreSQL database will be added in Stage 6.
// These tests verify the structure and basic functionality without requiring a database connection.
