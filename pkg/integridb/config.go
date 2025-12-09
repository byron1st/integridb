package integridb

import (
	"context"
	"time"
)

// Config defines the configuration for an IntegriDB connection.
type Config struct {
	// Driver is the database driver name (e.g., "postgres").
	Driver string

	// DSN is the data source name / connection string.
	DSN string

	// TrackedTables is the list of tables to track for event sourcing.
	// Only operations on these tables will generate events.
	TrackedTables []TableConfig

	// MaxOpenConns sets the maximum number of open connections to the database.
	// Default: 0 (unlimited)
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections in the pool.
	// Default: 2
	MaxIdleConns int

	// ConnMaxLifetime sets the maximum duration a connection may be reused.
	// Default: 0 (reuse forever)
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime sets the maximum duration a connection may be idle.
	// Default: 0 (no limit)
	ConnMaxIdleTime time.Duration

	// MetadataFunc is an optional function that extracts metadata from the context.
	// This allows injecting user-defined metadata (e.g., user_id, request_id) into events.
	MetadataFunc func(ctx context.Context) map[string]interface{}
}

// TableConfig defines configuration for a tracked table.
type TableConfig struct {
	// Name is the table name (case-sensitive).
	Name string

	// PrimaryKey is the list of column names that form the primary key.
	// For single-column keys, this will be a slice with one element.
	// For composite keys, list all columns in order.
	PrimaryKey []string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MaxIdleConns:    2,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		TrackedTables:   []TableConfig{},
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Driver == "" {
		return ErrInvalidConfig{Field: "Driver", Reason: "driver cannot be empty"}
	}
	if c.DSN == "" {
		return ErrInvalidConfig{Field: "DSN", Reason: "DSN cannot be empty"}
	}
	for i, table := range c.TrackedTables {
		if table.Name == "" {
			return ErrInvalidConfig{
				Field:  "TrackedTables",
				Reason: "table name cannot be empty",
				Index:  &i,
			}
		}
		if len(table.PrimaryKey) == 0 {
			return ErrInvalidConfig{
				Field:  "TrackedTables",
				Reason: "primary key cannot be empty",
				Index:  &i,
			}
		}
	}
	return nil
}

// FindTableConfig returns the TableConfig for a given table name.
// Returns nil if the table is not tracked.
func (c *Config) FindTableConfig(tableName string) *TableConfig {
	for i := range c.TrackedTables {
		if c.TrackedTables[i].Name == tableName {
			return &c.TrackedTables[i]
		}
	}
	return nil
}

// IsTracked returns true if the given table is configured for tracking.
func (c *Config) IsTracked(tableName string) bool {
	return c.FindTableConfig(tableName) != nil
}
