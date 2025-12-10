package integridb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// columnMetadata holds cached column information for a table.
type columnMetadata struct {
	columns []string
}

// columnCache caches column metadata to avoid repeated information_schema queries.
type columnCache struct {
	mu    sync.RWMutex
	cache map[string]*columnMetadata
}

// newColumnCache creates a new column cache.
func newColumnCache() *columnCache {
	return &columnCache{
		cache: make(map[string]*columnMetadata),
	}
}

// get retrieves column metadata from the cache.
// Returns nil if not cached.
func (c *columnCache) get(tableName string) *columnMetadata {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[tableName]
}

// set stores column metadata in the cache.
func (c *columnCache) set(tableName string, metadata *columnMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[tableName] = metadata
}

// getColumns retrieves the list of columns for a table.
// Uses cached metadata if available, otherwise queries information_schema.
func (d *DB) getColumns(ctx context.Context, tableName string) ([]string, error) {
	// Check cache first
	if d.colCache == nil {
		d.colCache = newColumnCache()
	}

	if metadata := d.colCache.get(tableName); metadata != nil {
		return metadata.columns, nil
	}

	// Query information_schema for column names
	const querySQL = `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := d.db.QueryContext(ctx, querySQL, tableName)
	if err != nil {
		return nil, fmt.Errorf("query columns for table %q: %w", tableName, err)
	}
	defer func() { _ = rows.Close() }()

	var columns []string
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, fmt.Errorf("scan column name: %w", err)
		}
		columns = append(columns, colName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate columns: %w", err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("table %q not found or has no columns", tableName)
	}

	// Cache the metadata
	d.colCache.set(tableName, &columnMetadata{columns: columns})

	return columns, nil
}

// captureBeforeState queries the current state of a row before a mutation.
// This is used for UPDATE and DELETE operations.
func (d *DB) captureBeforeState(ctx context.Context, tableName string, primaryKeyCols []string, rowID string) (map[string]any, error) {
	columns, err := d.getColumns(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	// Build SELECT query
	selectCols := strings.Join(columns, ", ")
	whereClause := buildWhereClause(primaryKeyCols, 1)
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectCols, tableName, whereClause)

	// Parse rowID into values for WHERE clause
	pkValues, err := parseRowID(rowID, len(primaryKeyCols))
	if err != nil {
		return nil, fmt.Errorf("parse row ID: %w", err)
	}

	// Execute query
	row := d.db.QueryRowContext(ctx, querySQL, pkValues...)

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

// captureAfterState queries the current state of a row after a mutation.
// This is used for INSERT and UPDATE operations.
func (d *DB) captureAfterState(ctx context.Context, tableName string, primaryKeyCols []string, rowID string) (map[string]any, error) {
	columns, err := d.getColumns(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	// Build SELECT query
	selectCols := strings.Join(columns, ", ")
	whereClause := buildWhereClause(primaryKeyCols, 1)
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectCols, tableName, whereClause)

	// Parse rowID into values for WHERE clause
	pkValues, err := parseRowID(rowID, len(primaryKeyCols))
	if err != nil {
		return nil, fmt.Errorf("parse row ID: %w", err)
	}

	// Execute query
	row := d.db.QueryRowContext(ctx, querySQL, pkValues...)

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

// captureAffectedRowIDs identifies all row IDs affected by an UPDATE or DELETE.
// It executes a SELECT with the same WHERE clause to get all matching primary keys.
func (d *DB) captureAffectedRowIDs(ctx context.Context, tableName string, primaryKeyCols []string, whereClause string, args []any) ([]string, error) {
	if len(primaryKeyCols) == 0 {
		return nil, fmt.Errorf("no primary key columns specified")
	}

	// Build SELECT query to get affected row IDs
	pkCols := strings.Join(primaryKeyCols, ", ")
	querySQL := fmt.Sprintf("SELECT %s FROM %s", pkCols, tableName)

	// Add WHERE clause if present
	if whereClause != "" {
		querySQL += " WHERE " + whereClause
	}

	rows, err := d.db.QueryContext(ctx, querySQL, args...)
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

// extractRowID extracts the primary key value from query result or arguments.
// For INSERT with RETURNING, it parses the result.
// For INSERT without RETURNING, it uses LastInsertId if available.
// For UPDATE/DELETE, it must be provided explicitly.
func extractRowID(result sql.Result, primaryKeyCols []string, returningValue *string) (string, error) {
	// If RETURNING clause was used, the row ID is already parsed
	if returningValue != nil && *returningValue != "" {
		return *returningValue, nil
	}

	// Try to use LastInsertId (only works for some drivers and single-column integer PKs)
	if len(primaryKeyCols) == 1 {
		id, err := result.LastInsertId()
		if err == nil && id > 0 {
			return fmt.Sprintf("%d", id), nil
		}
	}

	return "", fmt.Errorf("could not extract row ID: use RETURNING clause or provide explicit row ID")
}

// detectChangedColumns compares before and after states to identify changed columns.
// Returns a list of column names that have different values.
func detectChangedColumns(before, after map[string]any) []string {
	var changed []string

	for col, beforeVal := range before {
		afterVal, exists := after[col]
		if !exists {
			continue // Column doesn't exist in after state (shouldn't happen)
		}

		// Compare values using reflection to handle different types
		if !reflect.DeepEqual(beforeVal, afterVal) {
			changed = append(changed, col)
		}
	}

	return changed
}

// buildWhereClause constructs a WHERE clause for primary key lookup.
// Example: "id = $1" or "user_id = $1 AND order_id = $2"
func buildWhereClause(primaryKeyCols []string, startParam int) string {
	var parts []string
	for i, col := range primaryKeyCols {
		parts = append(parts, fmt.Sprintf("%s = $%d", col, startParam+i))
	}
	return strings.Join(parts, " AND ")
}

// buildRowID constructs a row ID string from primary key values.
// For single-column keys: the value as a string.
// For composite keys: JSON representation.
func buildRowID(pkValues []any) string {
	if len(pkValues) == 1 {
		// Single column primary key
		val := pkValues[0]
		if ptr, ok := val.(*any); ok {
			val = *ptr
		}
		return fmt.Sprintf("%v", val)
	}

	// Composite primary key - build JSON-like representation
	var parts []string
	for _, val := range pkValues {
		if ptr, ok := val.(*any); ok {
			val = *ptr
		}
		parts = append(parts, fmt.Sprintf("%v", val))
	}
	return strings.Join(parts, "|")
}

// parseRowID parses a row ID string into individual primary key values.
// Handles both single-column and composite keys.
func parseRowID(rowID string, numCols int) ([]any, error) {
	if numCols == 1 {
		return []any{rowID}, nil
	}

	// Composite key - split by delimiter
	parts := strings.Split(rowID, "|")
	if len(parts) != numCols {
		return nil, fmt.Errorf("row ID has %d parts, expected %d", len(parts), numCols)
	}

	values := make([]any, len(parts))
	for i, part := range parts {
		values[i] = part
	}

	return values, nil
}

// scanRowIntoMap scans a database row into a map[string]any.
func scanRowIntoMap(row *sql.Row, columns []string) (map[string]any, error) {
	// Create scan destinations
	scanDest := make([]any, len(columns))
	for i := range scanDest {
		scanDest[i] = new(any)
	}

	// Scan the row
	if err := row.Scan(scanDest...); err != nil {
		return nil, err
	}

	// Build map
	result := make(map[string]any)
	for i, col := range columns {
		val := scanDest[i].(*any)
		result[col] = *val
	}

	return result, nil
}
