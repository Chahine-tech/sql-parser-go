package tests

import (
	"context"
	"testing"

	"github.com/Chahine-tech/sql-parser-go/pkg/dialect"
	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

func TestNewWithDialect(t *testing.T) {
	sql := "SELECT * FROM users"
	ctx := context.Background()

	// Test with MySQL dialect
	mysqlDialect := dialect.GetDialect("mysql")
	p := parser.NewWithDialect(ctx, sql, mysqlDialect)

	if p.GetDialect().Name() != "MySQL" {
		t.Errorf("Expected MySQL dialect, got %s", p.GetDialect().Name())
	}

	// Test with PostgreSQL dialect
	postgresDialect := dialect.GetDialect("postgresql")
	p2 := parser.NewWithDialect(ctx, sql, postgresDialect)

	if p2.GetDialect().Name() != "PostgreSQL" {
		t.Errorf("Expected PostgreSQL dialect, got %s", p2.GetDialect().Name())
	}
}

func TestSetDialect(t *testing.T) {
	sql := "SELECT * FROM users"
	ctx := context.Background()

	p := parser.NewWithContext(ctx, sql)

	// Should start with SQL Server (default)
	if p.GetDialect().Name() != "SQL Server" {
		t.Errorf("Expected SQL Server dialect, got %s", p.GetDialect().Name())
	}

	// Change to MySQL
	mysqlDialect := dialect.GetDialect("mysql")
	p.SetDialect(mysqlDialect)

	if p.GetDialect().Name() != "MySQL" {
		t.Errorf("Expected MySQL dialect after SetDialect, got %s", p.GetDialect().Name())
	}
}

func TestDialectSpecificParsing(t *testing.T) {
	// Test cases for dialect-specific features
	tests := []struct {
		name        string
		dialect     string
		sql         string
		shouldParse bool
	}{
		{
			name:        "MySQL backticks",
			dialect:     "mysql",
			sql:         "SELECT `user_id` FROM `users`",
			shouldParse: true,
		},
		{
			name:        "PostgreSQL double quotes",
			dialect:     "postgresql",
			sql:         "SELECT \"user_id\" FROM \"users\"",
			shouldParse: true,
		},
		{
			name:        "SQL Server brackets",
			dialect:     "sqlserver",
			sql:         "SELECT [user_id] FROM [users]",
			shouldParse: true,
		},
		{
			name:        "SQLite double quotes",
			dialect:     "sqlite",
			sql:         "SELECT \"user_id\" FROM \"users\"",
			shouldParse: true,
		},
		{
			name:        "Oracle double quotes",
			dialect:     "oracle",
			sql:         "SELECT \"user_id\" FROM \"users\"",
			shouldParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialect)
			p := parser.NewWithDialect(context.Background(), tt.sql, d)

			_, err := p.ParseStatement()
			if tt.shouldParse && err != nil {
				t.Errorf("Expected %s to parse successfully with %s dialect, but got error: %v", tt.sql, tt.dialect, err)
			}
		})
	}
}

// TestDialectSpecificQuotingValidation tests that incorrect quoting styles fail with wrong dialects
func TestDialectSpecificQuotingValidation(t *testing.T) {
	tests := []struct {
		name          string
		dialect       string
		sql           string
		shouldFail    bool
		expectedError string
	}{
		{
			name:       "MySQL backticks should fail with PostgreSQL",
			dialect:    "postgresql",
			sql:        "SELECT `user_id` FROM `users`",
			shouldFail: true,
		},
		{
			name:       "SQL Server brackets should fail with MySQL",
			dialect:    "mysql",
			sql:        "SELECT [user_id] FROM [users]",
			shouldFail: true,
		},
		{
			name:       "PostgreSQL double quotes should work with Oracle",
			dialect:    "oracle",
			sql:        "SELECT \"user_id\" FROM \"users\"",
			shouldFail: false,
		},
		{
			name:       "SQLite double quotes should work with PostgreSQL",
			dialect:    "postgresql",
			sql:        "SELECT \"user_id\" FROM \"users\"",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialect)
			p := parser.NewWithDialect(context.Background(), tt.sql, d)

			_, err := p.ParseStatement()
			if tt.shouldFail && err == nil {
				t.Errorf("Expected %s to fail with %s dialect, but parsing succeeded", tt.sql, tt.dialect)
			} else if !tt.shouldFail && err != nil {
				t.Errorf("Expected %s to succeed with %s dialect, but got error: %v", tt.sql, tt.dialect, err)
			}
		})
	}
}

// TestDialectFeatureSupport tests that dialect-specific features are correctly identified
func TestDialectFeatureSupport(t *testing.T) {
	tests := []struct {
		dialectName string
		feature     string
		supported   bool
	}{
		{"mysql", "JSON", true},
		{"postgresql", "Arrays", true},
		{"postgresql", "RETURNING", true},
		{"sqlserver", "XML", true},
		{"sqlite", "UPSERT", true},
		{"oracle", "Partitioning", true},
		{"mysql", "Arrays", false},
		{"sqlite", "Arrays", false},
	}

	for _, tt := range tests {
		t.Run(tt.dialectName+"_"+tt.feature, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialectName)

			var supported bool
			switch tt.feature {
			case "JSON":
				supported = d.SupportsFeature(dialect.FeatureJSONSupport)
			case "Arrays":
				supported = d.SupportsFeature(dialect.FeatureArraySupport)
			case "RETURNING":
				supported = d.SupportsFeature(dialect.FeatureReturningClause)
			case "XML":
				supported = d.SupportsFeature(dialect.FeatureXMLSupport)
			case "UPSERT":
				supported = d.SupportsFeature(dialect.FeatureUpsert)
			case "Partitioning":
				supported = d.SupportsFeature(dialect.FeaturePartitioning)
			}

			if supported != tt.supported {
				t.Errorf("Expected %s %s support to be %v, got %v", tt.dialectName, tt.feature, tt.supported, supported)
			}
		})
	}
}

// TestDialectKeywordRecognition tests that dialect-specific keywords are properly recognized
func TestDialectKeywordRecognition(t *testing.T) {
	tests := []struct {
		dialectName string
		keyword     string
		isReserved  bool
	}{
		// Test some common keywords that should be recognized
		{"mysql", "SHOW", true},
		{"mysql", "REPLACE", true},
		{"mysql", "DESCRIBE", true}, // This is actually in common keywords
		{"postgresql", "RETURNING", true},
		{"postgresql", "LATERAL", true},
		{"sqlserver", "TOP", true},
		{"sqlserver", "IDENTITY", true},
		{"sqlserver", "MERGE", true},
		{"sqlite", "AUTOINCREMENT", true},
		{"sqlite", "PRAGMA", true},
		{"oracle", "ROWNUM", true},
		{"oracle", "CONNECT", true},
		{"oracle", "PRIOR", true},
		// Test keywords that should NOT be in specific dialects
		{"mysql", "RANDOMNONEXISTENTWORD", false},
		{"postgresql", "RANDOMNONEXISTENTWORD", false},
	}

	for _, tt := range tests {
		t.Run(tt.dialectName+"_"+tt.keyword, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialectName)
			isReserved := d.IsReservedWord(tt.keyword)

			if isReserved != tt.isReserved {
				t.Errorf("Expected %s keyword '%s' reserved status to be %v, got %v", tt.dialectName, tt.keyword, tt.isReserved, isReserved)
			}
		})
	}
}

// TestDialectLimitSyntax tests that each dialect reports the correct LIMIT syntax
func TestDialectLimitSyntax(t *testing.T) {
	tests := []struct {
		dialectName    string
		expectedSyntax dialect.LimitSyntax
	}{
		{"mysql", dialect.LimitSyntaxStandard},
		{"postgresql", dialect.LimitSyntaxStandard},
		{"sqlite", dialect.LimitSyntaxStandard},
		{"sqlserver", dialect.LimitSyntaxSQLServer},
		{"oracle", dialect.LimitSyntaxOracle},
	}

	for _, tt := range tests {
		t.Run(tt.dialectName, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialectName)
			syntax := d.GetLimitSyntax()

			if syntax != tt.expectedSyntax {
				t.Errorf("Expected %s to use %v syntax, got %v", tt.dialectName, tt.expectedSyntax, syntax)
			}
		})
	}
}

// TestComplexDialectSpecificQueries tests more complex dialect-specific SQL constructs
func TestComplexDialectSpecificQueries(t *testing.T) {
	tests := []struct {
		name        string
		dialect     string
		sql         string
		shouldParse bool
	}{
		{
			name:        "MySQL with database.table notation",
			dialect:     "mysql",
			sql:         "SELECT `db`.`table`.`column` FROM `database`.`table`",
			shouldParse: true,
		},
		{
			name:        "PostgreSQL with schema.table notation",
			dialect:     "postgresql",
			sql:         "SELECT \"schema\".\"table\".\"column\" FROM \"schema\".\"table\"",
			shouldParse: true,
		},
		{
			name:        "SQL Server with database.schema.table notation",
			dialect:     "sqlserver",
			sql:         "SELECT [database].[schema].[table].[column] FROM [database].[schema].[table]",
			shouldParse: true,
		},
		{
			name:        "Mixed quoting styles should work with appropriate dialects",
			dialect:     "mysql",
			sql:         "SELECT `quoted_col`, unquoted_col FROM `table`",
			shouldParse: true,
		},
		{
			name:        "PostgreSQL case sensitivity test",
			dialect:     "postgresql",
			sql:         "SELECT \"CaseSensitive\", lowercasenoquotes FROM \"users\"",
			shouldParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialect)
			p := parser.NewWithDialect(context.Background(), tt.sql, d)

			_, err := p.ParseStatement()
			if tt.shouldParse && err != nil {
				t.Errorf("Expected %s to parse successfully with %s dialect, but got error: %v", tt.sql, tt.dialect, err)
			} else if !tt.shouldParse && err == nil {
				t.Errorf("Expected %s to fail with %s dialect, but parsing succeeded", tt.sql, tt.dialect)
			}
		})
	}
}

// TestDialectErrorMessages tests that appropriate error messages are generated for dialect issues
func TestDialectErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		dialect     string
		sql         string
		expectError bool
	}{
		{
			name:        "Invalid backtick in PostgreSQL",
			dialect:     "postgresql",
			sql:         "SELECT `invalid` FROM users",
			expectError: true,
		},
		{
			name:        "Invalid bracket in MySQL",
			dialect:     "mysql",
			sql:         "SELECT [invalid] FROM users",
			expectError: true,
		},
		{
			name:        "Valid query should not error",
			dialect:     "mysql",
			sql:         "SELECT id FROM users",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialect)
			p := parser.NewWithDialect(context.Background(), tt.sql, d)

			_, err := p.ParseStatement()
			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s with %s dialect, but got none", tt.sql, tt.dialect)
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error for %s with %s dialect, but got: %v", tt.sql, tt.dialect, err)
			}
		})
	}
}

// TestDialectDataTypes tests that dialect-specific data types are recognized
func TestDialectDataTypes(t *testing.T) {
	tests := []struct {
		dialectName string
		dataType    string
		shouldExist bool
	}{
		{"mysql", "MEDIUMINT", true},
		{"mysql", "ENUM", true},
		{"mysql", "BIGSERIAL", false}, // PostgreSQL specific
		{"postgresql", "BIGSERIAL", true},
		{"postgresql", "JSONB", true},
		{"postgresql", "ENUM", false}, // MySQL has this, PostgreSQL creates custom types
		{"sqlserver", "UNIQUEIDENTIFIER", true},
		{"sqlserver", "NVARCHAR", true},
		{"sqlserver", "SERIAL", false}, // PostgreSQL specific
		{"sqlite", "INTEGER", true},
		{"sqlite", "TEXT", true},
		{"oracle", "NUMBER", true},
		{"oracle", "VARCHAR2", true},
		{"oracle", "AUTOINCREMENT", false}, // SQLite specific
	}

	for _, tt := range tests {
		t.Run(tt.dialectName+"_"+tt.dataType, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialectName)
			dataTypes := d.GetDataTypes()

			found := false
			for _, dt := range dataTypes {
				if dt == tt.dataType {
					found = true
					break
				}
			}

			if found != tt.shouldExist {
				t.Errorf("Expected %s data type '%s' existence to be %v, got %v", tt.dialectName, tt.dataType, tt.shouldExist, found)
			}
		})
	}
}
