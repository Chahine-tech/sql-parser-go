package dialect

import "strings"

// PostgreSQLDialect implements PostgreSQL-specific features
type PostgreSQLDialect struct{}

func (d *PostgreSQLDialect) Name() string {
	return "PostgreSQL"
}

func (d *PostgreSQLDialect) QuoteIdentifier(identifier string) string {
	return `"` + identifier + `"`
}

func (d *PostgreSQLDialect) SupportsFeature(feature Feature) bool {
	switch feature {
	case FeatureCTE:
		return true
	case FeatureWindowFunctions:
		return true
	case FeatureJSONSupport:
		return true
	case FeatureArraySupport:
		return true
	case FeatureRecursiveCTE:
		return true
	case FeaturePartitioning:
		return true
	case FeatureFullTextSearch:
		return true
	case FeatureXMLSupport:
		return true
	case FeatureUpsert:
		return true // INSERT ... ON CONFLICT
	case FeatureReturningClause:
		return true
	default:
		return false
	}
}

func (d *PostgreSQLDialect) GetKeywords() []string {
	postgresql := []string{
		"SERIAL", "BIGSERIAL", "SMALLSERIAL", "BYTEA", "INET", "CIDR", "MACADDR",
		"UUID", "XML", "JSONB", "HSTORE", "ARRAY", "TSQUERY", "TSVECTOR",
		"CONFLICT", "EXCLUDED", "RETURNING", "LATERAL", "TABLESAMPLE",
		"INHERITS", "LIKE", "INCLUDING", "EXCLUDING", "STORAGE", "COMMENTS",
		"DEFAULTS", "CONSTRAINTS", "INDEXES", "STATISTICS", "ALL",
		"CONCURRENTLY", "VACUUM", "ANALYZE", "CLUSTER", "REINDEX",
		"EXPLAIN", "VERBOSE", "COSTS", "BUFFERS", "TIMING", "SUMMARY", "FORMAT",
		"ILIKE", "SIMILAR", "POSIX", "REGEXP_REPLACE", "REGEXP_SPLIT_TO_TABLE",
		"WINDOW", "OVER", "PARTITION", "ROWS", "RANGE", "UNBOUNDED", "PRECEDING",
		"FOLLOWING", "CURRENT", "ROW", "GROUPS", "EXCLUDE", "TIES", "OTHERS",
	}
	return append(CommonKeywords, postgresql...)
}

func (d *PostgreSQLDialect) GetDataTypes() []string {
	postgresql := []string{
		"SERIAL", "BIGSERIAL", "SMALLSERIAL",
		"BYTEA", "INET", "CIDR", "MACADDR", "MACADDR8",
		"UUID", "XML", "JSON", "JSONB",
		"ARRAY", "HSTORE", "LTREE", "CUBE", "ISN",
		"TSQUERY", "TSVECTOR", "TXID_SNAPSHOT",
		"INT4RANGE", "INT8RANGE", "NUMRANGE", "TSRANGE", "TSTZRANGE", "DATERANGE",
		"MONEY", "PG_LSN", "PG_SNAPSHOT",
		"INTERVAL", "TIME WITH TIME ZONE", "TIMESTAMP WITH TIME ZONE",
		"POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE",
	}
	return append(CommonDataTypes, postgresql...)
}

func (d *PostgreSQLDialect) IsReservedWord(word string) bool {
	reserved := map[string]bool{
		"ALL": true, "ANALYSE": true, "ANALYZE": true, "AND": true, "ANY": true,
		"ARRAY": true, "AS": true, "ASC": true, "ASYMMETRIC": true, "AUTHORIZATION": true,
		"BINARY": true, "BOTH": true, "CASE": true, "CAST": true, "CHECK": true,
		"COLLATE": true, "COLLATION": true, "COLUMN": true, "CONCURRENTLY": true, "CONSTRAINT": true,
		"CREATE": true, "CROSS": true, "CURRENT_CATALOG": true, "CURRENT_DATE": true, "CURRENT_ROLE": true,
		"CURRENT_SCHEMA": true, "CURRENT_TIME": true, "CURRENT_TIMESTAMP": true, "CURRENT_USER": true,
		"DEFAULT": true, "DEFERRABLE": true, "DESC": true, "DISTINCT": true, "DO": true,
		"ELSE": true, "END": true, "EXCEPT": true, "FALSE": true, "FETCH": true,
		"FOR": true, "FOREIGN": true, "FREEZE": true, "FROM": true, "FULL": true,
		"GRANT": true, "GROUP": true, "HAVING": true, "ILIKE": true, "IN": true,
		"INITIALLY": true, "INNER": true, "INTERSECT": true, "INTO": true, "IS": true,
		"ISNULL": true, "JOIN": true, "LATERAL": true, "LEADING": true, "LEFT": true,
		"LIKE": true, "LIMIT": true, "LOCALTIME": true, "LOCALTIMESTAMP": true, "NATURAL": true,
		"NOT": true, "NOTNULL": true, "NULL": true, "OFFSET": true, "ON": true,
		"ONLY": true, "OR": true, "ORDER": true, "OUTER": true, "OVERLAPS": true,
		"PLACING": true, "PRIMARY": true, "REFERENCES": true, "RETURNING": true, "RIGHT": true,
		"SELECT": true, "SESSION_USER": true, "SIMILAR": true, "SOME": true, "SYMMETRIC": true,
		"TABLE": true, "TABLESAMPLE": true, "THEN": true, "TO": true, "TRAILING": true,
		"TRUE": true, "UNION": true, "UNIQUE": true, "USER": true, "USING": true,
		"VARIADIC": true, "VERBOSE": true, "WHEN": true, "WHERE": true, "WINDOW": true,
		"WITH": true,
	}
	return reserved[strings.ToUpper(word)]
}

func (d *PostgreSQLDialect) GetLimitSyntax() LimitSyntax {
	return LimitSyntaxStandard
}
