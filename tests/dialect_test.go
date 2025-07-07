package tests

import (
	"testing"

	"github.com/Chahine-tech/sql-parser-go/pkg/dialect"
)

func TestGetDialect(t *testing.T) {
	tests := []struct {
		name     string
		dialect  string
		expected string
	}{
		{"MySQL", "mysql", "MySQL"},
		{"PostgreSQL", "postgresql", "PostgreSQL"},
		{"PostgreSQL short", "postgres", "PostgreSQL"},
		{"SQL Server", "sqlserver", "SQL Server"},
		{"SQL Server short", "mssql", "SQL Server"},
		{"SQLite", "sqlite", "SQLite"},
		{"Oracle", "oracle", "Oracle"},
		{"Default fallback", "unknown", "SQL Server"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dialect.GetDialect(tt.dialect)
			if d.Name() != tt.expected {
				t.Errorf("GetDialect(%q) = %q, want %q", tt.dialect, d.Name(), tt.expected)
			}
		})
	}
}

func TestMySQL(t *testing.T) {
	d := &dialect.MySQLDialect{}

	if d.Name() != "MySQL" {
		t.Errorf("Expected MySQL, got %s", d.Name())
	}

	if d.QuoteIdentifier("table") != "`table`" {
		t.Errorf("Expected `table`, got %s", d.QuoteIdentifier("table"))
	}

	if !d.SupportsFeature(dialect.FeatureJSONSupport) {
		t.Error("MySQL should support JSON")
	}

	if d.GetLimitSyntax() != dialect.LimitSyntaxStandard {
		t.Error("MySQL should use standard LIMIT syntax")
	}
}

func TestPostgreSQL(t *testing.T) {
	d := &dialect.PostgreSQLDialect{}

	if d.Name() != "PostgreSQL" {
		t.Errorf("Expected PostgreSQL, got %s", d.Name())
	}

	if d.QuoteIdentifier("table") != `"table"` {
		t.Errorf("Expected \"table\", got %s", d.QuoteIdentifier("table"))
	}

	if !d.SupportsFeature(dialect.FeatureArraySupport) {
		t.Error("PostgreSQL should support arrays")
	}

	if !d.SupportsFeature(dialect.FeatureReturningClause) {
		t.Error("PostgreSQL should support RETURNING clause")
	}

	if d.GetLimitSyntax() != dialect.LimitSyntaxStandard {
		t.Error("PostgreSQL should use standard LIMIT syntax")
	}
}

func TestSQLServer(t *testing.T) {
	d := &dialect.SQLServerDialect{}

	if d.Name() != "SQL Server" {
		t.Errorf("Expected SQL Server, got %s", d.Name())
	}

	if d.QuoteIdentifier("table") != "[table]" {
		t.Errorf("Expected [table], got %s", d.QuoteIdentifier("table"))
	}

	if !d.SupportsFeature(dialect.FeatureXMLSupport) {
		t.Error("SQL Server should support XML")
	}

	if d.GetLimitSyntax() != dialect.LimitSyntaxSQLServer {
		t.Error("SQL Server should use TOP syntax")
	}
}

func TestSQLite(t *testing.T) {
	d := &dialect.SQLiteDialect{}

	if d.Name() != "SQLite" {
		t.Errorf("Expected SQLite, got %s", d.Name())
	}

	if d.QuoteIdentifier("table") != `"table"` {
		t.Errorf("Expected \"table\", got %s", d.QuoteIdentifier("table"))
	}

	if !d.SupportsFeature(dialect.FeatureUpsert) {
		t.Error("SQLite should support UPSERT")
	}

	if d.GetLimitSyntax() != dialect.LimitSyntaxStandard {
		t.Error("SQLite should use standard LIMIT syntax")
	}
}

func TestOracle(t *testing.T) {
	d := &dialect.OracleDialect{}

	if d.Name() != "Oracle" {
		t.Errorf("Expected Oracle, got %s", d.Name())
	}

	if d.QuoteIdentifier("table") != `"table"` {
		t.Errorf("Expected \"table\", got %s", d.QuoteIdentifier("table"))
	}

	if !d.SupportsFeature(dialect.FeaturePartitioning) {
		t.Error("Oracle should support partitioning")
	}

	if d.GetLimitSyntax() != dialect.LimitSyntaxOracle {
		t.Error("Oracle should use ROWNUM syntax")
	}
}
