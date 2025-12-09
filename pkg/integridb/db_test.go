package integridb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 2, config.MaxIdleConns)
	assert.Empty(t, config.TrackedTables)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Driver: "postgres",
				DSN:    "postgres://localhost/test",
				TrackedTables: []TableConfig{
					{Name: "users", PrimaryKey: []string{"id"}},
				},
			},
			wantErr: false,
		},
		{
			name: "missing driver",
			config: &Config{
				DSN: "postgres://localhost/test",
			},
			wantErr: true,
		},
		{
			name: "missing DSN",
			config: &Config{
				Driver: "postgres",
			},
			wantErr: true,
		},
		{
			name: "empty table name",
			config: &Config{
				Driver: "postgres",
				DSN:    "postgres://localhost/test",
				TrackedTables: []TableConfig{
					{Name: "", PrimaryKey: []string{"id"}},
				},
			},
			wantErr: true,
		},
		{
			name: "empty primary key",
			config: &Config{
				Driver: "postgres",
				DSN:    "postgres://localhost/test",
				TrackedTables: []TableConfig{
					{Name: "users", PrimaryKey: []string{}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_FindTableConfig(t *testing.T) {
	config := &Config{
		TrackedTables: []TableConfig{
			{Name: "users", PrimaryKey: []string{"id"}},
			{Name: "orders", PrimaryKey: []string{"order_id"}},
		},
	}

	// Found
	tableConfig := config.FindTableConfig("users")
	require.NotNil(t, tableConfig)
	assert.Equal(t, "users", tableConfig.Name)
	assert.Equal(t, []string{"id"}, tableConfig.PrimaryKey)

	// Not found
	tableConfig = config.FindTableConfig("products")
	assert.Nil(t, tableConfig)
}

func TestConfig_IsTracked(t *testing.T) {
	config := &Config{
		TrackedTables: []TableConfig{
			{Name: "users", PrimaryKey: []string{"id"}},
		},
	}

	assert.True(t, config.IsTracked("users"))
	assert.False(t, config.IsTracked("products"))
}

// Note: Full integration tests with real database connections will be added in Stage 6.
// These tests verify the configuration and structure without requiring a database.
