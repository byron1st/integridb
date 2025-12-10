package query

import (
	"testing"
)

func TestParse_INSERT(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		wantType      QueryType
		wantTableName string
		wantColumns   []string
		wantErr       bool
	}{
		{
			name:          "simple INSERT",
			sql:           "INSERT INTO users (id, name) VALUES ($1, $2)",
			wantType:      QueryTypeInsert,
			wantTableName: "users",
			wantColumns:   []string{"id", "name"},
			wantErr:       false,
		},
		{
			name:          "INSERT with quoted table name",
			sql:           `INSERT INTO "users" (id, name) VALUES ($1, $2)`,
			wantType:      QueryTypeInsert,
			wantTableName: "users",
			wantColumns:   []string{"id", "name"},
			wantErr:       false,
		},
		{
			name:          "INSERT without column list",
			sql:           "INSERT INTO users VALUES ($1, $2, $3)",
			wantType:      QueryTypeInsert,
			wantTableName: "users",
			wantColumns:   nil,
			wantErr:       false,
		},
		{
			name:          "INSERT with uppercase",
			sql:           "INSERT INTO USERS (ID, NAME) VALUES ($1, $2)",
			wantType:      QueryTypeInsert,
			wantTableName: "USERS",
			wantColumns:   []string{"ID", "NAME"},
			wantErr:       false,
		},
		{
			name:          "INSERT with leading whitespace",
			sql:           "  INSERT INTO users (id) VALUES ($1)",
			wantType:      QueryTypeInsert,
			wantTableName: "users",
			wantColumns:   []string{"id"},
			wantErr:       false,
		},
		{
			name:          "INSERT with multiple columns",
			sql:           "INSERT INTO orders (id, user_id, total, created_at) VALUES ($1, $2, $3, $4)",
			wantType:      QueryTypeInsert,
			wantTableName: "orders",
			wantColumns:   []string{"id", "user_id", "total", "created_at"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.TableName != tt.wantTableName {
				t.Errorf("Parse() TableName = %v, want %v", got.TableName, tt.wantTableName)
			}
			if len(got.Columns) != len(tt.wantColumns) {
				t.Errorf("Parse() Columns length = %v, want %v", len(got.Columns), len(tt.wantColumns))
			} else {
				for i, col := range got.Columns {
					if col != tt.wantColumns[i] {
						t.Errorf("Parse() Columns[%d] = %v, want %v", i, col, tt.wantColumns[i])
					}
				}
			}
		})
	}
}

func TestParse_UPDATE(t *testing.T) {
	tests := []struct {
		name            string
		sql             string
		wantType        QueryType
		wantTableName   string
		wantWhereClause string
		wantErr         bool
	}{
		{
			name:            "simple UPDATE",
			sql:             "UPDATE users SET name = $1 WHERE id = $2",
			wantType:        QueryTypeUpdate,
			wantTableName:   "users",
			wantWhereClause: "id = $2",
			wantErr:         false,
		},
		{
			name:            "UPDATE with quoted table name",
			sql:             `UPDATE "users" SET name = $1 WHERE id = $2`,
			wantType:        QueryTypeUpdate,
			wantTableName:   "users",
			wantWhereClause: "id = $2",
			wantErr:         false,
		},
		{
			name:            "UPDATE without WHERE",
			sql:             "UPDATE users SET name = $1",
			wantType:        QueryTypeUpdate,
			wantTableName:   "users",
			wantWhereClause: "",
			wantErr:         false,
		},
		{
			name:            "UPDATE with multiple SET clauses",
			sql:             "UPDATE users SET name = $1, email = $2 WHERE id = $3",
			wantType:        QueryTypeUpdate,
			wantTableName:   "users",
			wantWhereClause: "id = $3",
			wantErr:         false,
		},
		{
			name:            "UPDATE with complex WHERE",
			sql:             "UPDATE orders SET status = $1 WHERE user_id = $2 AND created_at > $3",
			wantType:        QueryTypeUpdate,
			wantTableName:   "orders",
			wantWhereClause: "user_id = $2 AND created_at > $3",
			wantErr:         false,
		},
		{
			name:            "UPDATE with uppercase",
			sql:             "UPDATE USERS SET NAME = $1 WHERE ID = $2",
			wantType:        QueryTypeUpdate,
			wantTableName:   "USERS",
			wantWhereClause: "ID = $2",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.TableName != tt.wantTableName {
				t.Errorf("Parse() TableName = %v, want %v", got.TableName, tt.wantTableName)
			}
			if got.WhereClause != tt.wantWhereClause {
				t.Errorf("Parse() WhereClause = %v, want %v", got.WhereClause, tt.wantWhereClause)
			}
		})
	}
}

func TestParse_DELETE(t *testing.T) {
	tests := []struct {
		name            string
		sql             string
		wantType        QueryType
		wantTableName   string
		wantWhereClause string
		wantErr         bool
	}{
		{
			name:            "simple DELETE",
			sql:             "DELETE FROM users WHERE id = $1",
			wantType:        QueryTypeDelete,
			wantTableName:   "users",
			wantWhereClause: "id = $1",
			wantErr:         false,
		},
		{
			name:            "DELETE with quoted table name",
			sql:             `DELETE FROM "users" WHERE id = $1`,
			wantType:        QueryTypeDelete,
			wantTableName:   "users",
			wantWhereClause: "id = $1",
			wantErr:         false,
		},
		{
			name:            "DELETE without WHERE",
			sql:             "DELETE FROM users",
			wantType:        QueryTypeDelete,
			wantTableName:   "users",
			wantWhereClause: "",
			wantErr:         false,
		},
		{
			name:            "DELETE with complex WHERE",
			sql:             "DELETE FROM orders WHERE user_id = $1 AND status = $2",
			wantType:        QueryTypeDelete,
			wantTableName:   "orders",
			wantWhereClause: "user_id = $1 AND status = $2",
			wantErr:         false,
		},
		{
			name:            "DELETE with uppercase",
			sql:             "DELETE FROM USERS WHERE ID = $1",
			wantType:        QueryTypeDelete,
			wantTableName:   "USERS",
			wantWhereClause: "ID = $1",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.TableName != tt.wantTableName {
				t.Errorf("Parse() TableName = %v, want %v", got.TableName, tt.wantTableName)
			}
			if got.WhereClause != tt.wantWhereClause {
				t.Errorf("Parse() WhereClause = %v, want %v", got.WhereClause, tt.wantWhereClause)
			}
		})
	}
}

func TestParse_SELECT(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		wantType QueryType
		wantErr  bool
	}{
		{
			name:     "simple SELECT",
			sql:      "SELECT * FROM users",
			wantType: QueryTypeSelect,
			wantErr:  false,
		},
		{
			name:     "SELECT with WHERE",
			sql:      "SELECT id, name FROM users WHERE id = $1",
			wantType: QueryTypeSelect,
			wantErr:  false,
		},
		{
			name:     "SELECT with JOIN",
			sql:      "SELECT u.*, o.* FROM users u JOIN orders o ON u.id = o.user_id",
			wantType: QueryTypeSelect,
			wantErr:  false,
		},
		{
			name:     "SELECT with lowercase",
			sql:      "select * from users",
			wantType: QueryTypeSelect,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", got.Type, tt.wantType)
			}
		})
	}
}

func TestParse_Other(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		wantType QueryType
		wantErr  bool
	}{
		{
			name:     "CREATE TABLE",
			sql:      "CREATE TABLE users (id INT PRIMARY KEY)",
			wantType: QueryTypeOther,
			wantErr:  false,
		},
		{
			name:     "DROP TABLE",
			sql:      "DROP TABLE users",
			wantType: QueryTypeOther,
			wantErr:  false,
		},
		{
			name:     "ALTER TABLE",
			sql:      "ALTER TABLE users ADD COLUMN email VARCHAR(255)",
			wantType: QueryTypeOther,
			wantErr:  false,
		},
		{
			name:     "TRUNCATE",
			sql:      "TRUNCATE TABLE users",
			wantType: QueryTypeOther,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", got.Type, tt.wantType)
			}
		})
	}
}

func TestParse_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "empty string",
			sql:     "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			sql:     "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParsedQuery_IsMutation(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantMut bool
	}{
		{
			name:    "INSERT is mutation",
			sql:     "INSERT INTO users (id) VALUES ($1)",
			wantMut: true,
		},
		{
			name:    "UPDATE is mutation",
			sql:     "UPDATE users SET name = $1 WHERE id = $2",
			wantMut: true,
		},
		{
			name:    "DELETE is mutation",
			sql:     "DELETE FROM users WHERE id = $1",
			wantMut: true,
		},
		{
			name:    "SELECT is not mutation",
			sql:     "SELECT * FROM users",
			wantMut: false,
		},
		{
			name:    "OTHER is not mutation",
			sql:     "CREATE TABLE users (id INT)",
			wantMut: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Parse(tt.sql)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got := pq.IsMutation(); got != tt.wantMut {
				t.Errorf("IsMutation() = %v, want %v", got, tt.wantMut)
			}
		})
	}
}

func TestQueryType_String(t *testing.T) {
	tests := []struct {
		qt   QueryType
		want string
	}{
		{QueryTypeSelect, "SELECT"},
		{QueryTypeInsert, "INSERT"},
		{QueryTypeUpdate, "UPDATE"},
		{QueryTypeDelete, "DELETE"},
		{QueryTypeOther, "OTHER"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.qt.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
