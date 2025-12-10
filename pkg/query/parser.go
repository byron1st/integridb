package query

import (
	"fmt"
	"regexp"
	"strings"
)

// QueryType represents the type of SQL query.
type QueryType int

const (
	// QueryTypeOther represents an unsupported or unrecognized query.
	QueryTypeOther QueryType = iota
	// QueryTypeSelect represents a SELECT query.
	QueryTypeSelect
	// QueryTypeInsert represents an INSERT query.
	QueryTypeInsert
	// QueryTypeUpdate represents an UPDATE query.
	QueryTypeUpdate
	// QueryTypeDelete represents a DELETE query.
	QueryTypeDelete
)

// String returns the string representation of a QueryType.
func (qt QueryType) String() string {
	switch qt {
	case QueryTypeSelect:
		return "SELECT"
	case QueryTypeInsert:
		return "INSERT"
	case QueryTypeUpdate:
		return "UPDATE"
	case QueryTypeDelete:
		return "DELETE"
	default:
		return "OTHER"
	}
}

// ParsedQuery represents a parsed SQL statement.
type ParsedQuery struct {
	// Type is the type of query (INSERT, UPDATE, DELETE, SELECT, OTHER).
	Type QueryType
	// TableName is the name of the table being operated on.
	TableName string
	// Columns is the list of columns mentioned in the query (for INSERT).
	Columns []string
	// WhereClause is the WHERE clause if present (for UPDATE/DELETE).
	WhereClause string
}

var (
	// Regular expressions for parsing SQL statements
	insertRegex = regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+([a-zA-Z_][a-zA-Z0-9_]*|"[^"]+")`)
	updateRegex = regexp.MustCompile(`(?i)^\s*UPDATE\s+([a-zA-Z_][a-zA-Z0-9_]*|"[^"]+")`)
	deleteRegex = regexp.MustCompile(`(?i)^\s*DELETE\s+FROM\s+([a-zA-Z_][a-zA-Z0-9_]*|"[^"]+")`)
	selectRegex = regexp.MustCompile(`(?i)^\s*SELECT\s+`)

	// Regex for extracting WHERE clause
	whereRegex = regexp.MustCompile(`(?i)\s+WHERE\s+(.*)`)

	// Regex for extracting columns in INSERT statement
	insertColumnsRegex = regexp.MustCompile(`(?i)INSERT\s+INTO\s+(?:[a-zA-Z_][a-zA-Z0-9_]*|"[^"]+")\s*\(([^)]+)\)`)
)

// Parse analyzes a SQL statement and returns a ParsedQuery.
// It identifies the query type, extracts the table name, and captures relevant metadata.
func Parse(sql string) (*ParsedQuery, error) {
	if sql == "" {
		return nil, fmt.Errorf("empty SQL statement")
	}

	// Trim whitespace and normalize
	sql = strings.TrimSpace(sql)

	if sql == "" {
		return nil, fmt.Errorf("empty SQL statement")
	}

	pq := &ParsedQuery{
		Type: QueryTypeOther,
	}

	// Check for SELECT
	if selectRegex.MatchString(sql) {
		pq.Type = QueryTypeSelect
		return pq, nil
	}

	// Check for INSERT
	if matches := insertRegex.FindStringSubmatch(sql); len(matches) > 1 {
		pq.Type = QueryTypeInsert
		pq.TableName = normalizeIdentifier(matches[1])

		// Try to extract column names
		if colMatches := insertColumnsRegex.FindStringSubmatch(sql); len(colMatches) > 1 {
			colStr := colMatches[1]
			cols := strings.Split(colStr, ",")
			for _, col := range cols {
				pq.Columns = append(pq.Columns, strings.TrimSpace(col))
			}
		}

		return pq, nil
	}

	// Check for UPDATE
	if matches := updateRegex.FindStringSubmatch(sql); len(matches) > 1 {
		pq.Type = QueryTypeUpdate
		pq.TableName = normalizeIdentifier(matches[1])

		// Extract WHERE clause
		if whereMatches := whereRegex.FindStringSubmatch(sql); len(whereMatches) > 1 {
			pq.WhereClause = strings.TrimSpace(whereMatches[1])
		}

		return pq, nil
	}

	// Check for DELETE
	if matches := deleteRegex.FindStringSubmatch(sql); len(matches) > 1 {
		pq.Type = QueryTypeDelete
		pq.TableName = normalizeIdentifier(matches[1])

		// Extract WHERE clause
		if whereMatches := whereRegex.FindStringSubmatch(sql); len(whereMatches) > 1 {
			pq.WhereClause = strings.TrimSpace(whereMatches[1])
		}

		return pq, nil
	}

	// If we got here, it's an unrecognized query type
	return pq, nil
}

// normalizeIdentifier removes quotes from identifiers and returns the clean name.
func normalizeIdentifier(identifier string) string {
	// Remove leading/trailing whitespace
	identifier = strings.TrimSpace(identifier)

	// Remove double quotes if present
	if strings.HasPrefix(identifier, `"`) && strings.HasSuffix(identifier, `"`) {
		return identifier[1 : len(identifier)-1]
	}

	return identifier
}

// IsMutation returns true if the query is a mutation (INSERT, UPDATE, DELETE).
func (pq *ParsedQuery) IsMutation() bool {
	return pq.Type == QueryTypeInsert ||
		pq.Type == QueryTypeUpdate ||
		pq.Type == QueryTypeDelete
}
