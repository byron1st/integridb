# Task 2.1: Implement SQL Parser for Mutation Detection

## Description
Parse SQL statements to identify INSERT, UPDATE, DELETE operations and extract table names and relevant metadata.

## Actions

1. Create `pkg/query/parser.go` with:
   - `ParsedQuery` struct (Type, TableName, Columns, WhereClause)
   - `QueryType` enum (SELECT, INSERT, UPDATE, DELETE, OTHER)
   - `Parse()` function to analyze SQL statements

2. Implement parsing logic:
   - Use regex or simple tokenization (avoid heavy parser libraries for MVP)
   - Handle basic cases: `INSERT INTO table`, `UPDATE table SET`, `DELETE FROM table`
   - Extract table names reliably
   - Handle quoted identifiers
   - Return `QueryTypeOther` for unsupported/complex queries

3. Write comprehensive tests in `pkg/query/parser_test.go`:
   - Test INSERT with various formats
   - Test UPDATE with WHERE clauses
   - Test DELETE with WHERE clauses
   - Test SELECT (should return QueryTypeSelect)
   - Test edge cases (quoted names, schemas, etc.)

## Acceptance Criteria

- [ ] Parser correctly identifies INSERT, UPDATE, DELETE, SELECT
- [ ] Parser extracts table names accurately
- [ ] Parser handles PostgreSQL-specific syntax
- [ ] Unit tests cover >90% of parser code
- [ ] Parser returns clear errors for unparseable queries
