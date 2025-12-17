package db

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveDatabaseFromConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		dbName   string
		expected string
	}{
		// Стандартные случаи
		{
			name:     "Standard connection string",
			input:    "postgresql://user:pass@localhost:5432/mydb",
			dbName:   "mydb",
			expected: "postgresql://user:pass@localhost:5432",
		},
		{
			name:     "Database name matches username",
			input:    "postgresql://user:pass@localhost:5432/user",
			dbName:   "user",
			expected: "postgresql://user:pass@localhost:5432",
		},
		// С параметрами
		{
			name:     "With query parameters",
			input:    "postgresql://user:pass@localhost:5432/mydb?sslmode=disable",
			dbName:   "mydb",
			expected: "postgresql://user:pass@localhost:5432?sslmode=disable",
		},
		{
			name:     "With multiple query parameters",
			input:    "postgresql://user:pass@localhost:5432/mydb?sslmode=disable&pool_max_conns=10",
			dbName:   "mydb",
			expected: "postgresql://user:pass@localhost:5432?sslmode=disable&pool_max_conns=10",
		},
		// Edge cases
		{
			name:     "Empty database name",
			input:    "postgresql://user:pass@localhost:5432/",
			dbName:   "",
			expected: "postgresql://user:pass@localhost:5432",
		},
		{
			name:     "No database in string",
			input:    "postgresql://user:pass@localhost:5432",
			dbName:   "mydb",
			expected: "postgresql://user:pass@localhost:5432",
		},
		{
			name:     "Database name with special chars",
			input:    "postgresql://user:pass@localhost:5432/my-db_123",
			dbName:   "my-db_123",
			expected: "postgresql://user:pass@localhost:5432",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDatabaseFromConnectionString(tt.input, tt.dbName)
			require.Equal(t, tt.expected, got)
		})
	}
}
