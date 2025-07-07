package dialect

import "strings"

// Dialect represents a SQL dialect with its specific features and syntax
type Dialect interface {
	// Name returns the dialect name
	Name() string

	// QuoteIdentifier returns the proper way to quote identifiers in this dialect
	QuoteIdentifier(identifier string) string

	// SupportsFeature checks if a feature is supported in this dialect
	SupportsFeature(feature Feature) bool

	// GetKeywords returns dialect-specific keywords
	GetKeywords() []string

	// GetDataTypes returns dialect-specific data types
	GetDataTypes() []string

	// IsReservedWord checks if a word is reserved in this dialect
	IsReservedWord(word string) bool

	// GetLimitSyntax returns the LIMIT clause syntax for this dialect
	GetLimitSyntax() LimitSyntax
}

// Feature represents SQL features that may vary between dialects
type Feature int

const (
	FeatureCTE Feature = iota
	FeatureWindowFunctions
	FeatureJSONSupport
	FeatureArraySupport
	FeatureRecursiveCTE
	FeaturePartitioning
	FeatureFullTextSearch
	FeatureXMLSupport
	FeatureUpsert
	FeatureReturningClause
)

// LimitSyntax represents different ways to limit results
type LimitSyntax int

const (
	LimitSyntaxStandard  LimitSyntax = iota // LIMIT n OFFSET m
	LimitSyntaxSQLServer                    // TOP n
	LimitSyntaxOracle                       // ROWNUM
)

// GetDialect returns the appropriate dialect implementation
func GetDialect(name string) Dialect {
	switch strings.ToLower(name) {
	case "mysql":
		return &MySQLDialect{}
	case "postgresql", "postgres":
		return &PostgreSQLDialect{}
	case "sqlserver", "mssql":
		return &SQLServerDialect{}
	case "sqlite":
		return &SQLiteDialect{}
	case "oracle":
		return &OracleDialect{}
	default:
		return &SQLServerDialect{} // Default fallback
	}
}

// Common keywords shared across most SQL dialects
var CommonKeywords = []string{
	"SELECT", "FROM", "WHERE", "JOIN", "INNER", "LEFT", "RIGHT", "FULL", "OUTER",
	"ON", "AS", "AND", "OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS", "NULL",
	"INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE", "CREATE", "TABLE", "INDEX",
	"DROP", "ALTER", "ADD", "COLUMN", "PRIMARY", "KEY", "FOREIGN", "REFERENCES",
	"UNIQUE", "CHECK", "DEFAULT", "AUTO_INCREMENT", "IDENTITY",
	"GROUP", "BY", "HAVING", "ORDER", "ASC", "DESC", "LIMIT", "OFFSET",
	"UNION", "ALL", "DISTINCT", "CASE", "WHEN", "THEN", "ELSE", "END",
	"CAST", "CONVERT", "COUNT", "SUM", "AVG", "MIN", "MAX",
}

// Common data types
var CommonDataTypes = []string{
	"INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT",
	"DECIMAL", "NUMERIC", "FLOAT", "REAL", "DOUBLE",
	"CHAR", "VARCHAR", "TEXT", "NCHAR", "NVARCHAR",
	"DATE", "TIME", "DATETIME", "TIMESTAMP",
	"BOOLEAN", "BOOL", "BIT",
	"BLOB", "BINARY", "VARBINARY",
}
