package dialect

import "strings"

// MySQLDialect implements MySQL-specific features
type MySQLDialect struct{}

func (d *MySQLDialect) Name() string {
	return "MySQL"
}

func (d *MySQLDialect) QuoteIdentifier(identifier string) string {
	return "`" + identifier + "`"
}

func (d *MySQLDialect) SupportsFeature(feature Feature) bool {
	switch feature {
	case FeatureCTE:
		return true // MySQL 8.0+
	case FeatureWindowFunctions:
		return true // MySQL 8.0+
	case FeatureJSONSupport:
		return true // MySQL 5.7+
	case FeatureRecursiveCTE:
		return true // MySQL 8.0+
	case FeatureFullTextSearch:
		return true
	case FeatureUpsert:
		return true // INSERT ... ON DUPLICATE KEY UPDATE
	default:
		return false
	}
}

func (d *MySQLDialect) GetKeywords() []string {
	mysql := []string{
		"AUTO_INCREMENT", "BINARY", "BLOB", "ENUM", "MEDIUMTEXT", "LONGTEXT",
		"TINYTEXT", "MEDIUMBLOB", "LONGBLOB", "TINYBLOB", "ZEROFILL", "UNSIGNED",
		"SHOW", "DATABASES", "TABLES", "COLUMNS", "STATUS", "VARIABLES",
		"REPLACE", "DUPLICATE", "IGNORE", "DELAYED", "LOW_PRIORITY", "HIGH_PRIORITY",
		"STRAIGHT_JOIN", "FORCE", "USE", "PARTITION", "SUBPARTITION",
	}
	return append(CommonKeywords, mysql...)
}

func (d *MySQLDialect) GetDataTypes() []string {
	mysql := []string{
		"TINYINT", "MEDIUMINT", "ENUM", "SET",
		"TINYTEXT", "MEDIUMTEXT", "LONGTEXT",
		"TINYBLOB", "MEDIUMBLOB", "LONGBLOB",
		"YEAR", "JSON", "GEOMETRY", "POINT", "LINESTRING", "POLYGON",
	}
	return append(CommonDataTypes, mysql...)
}

func (d *MySQLDialect) IsReservedWord(word string) bool {
	reserved := map[string]bool{
		"ADD": true, "ALL": true, "ALTER": true, "ANALYZE": true, "AND": true,
		"AS": true, "ASC": true, "ASENSITIVE": true, "BEFORE": true, "BETWEEN": true,
		"BIGINT": true, "BINARY": true, "BLOB": true, "BOTH": true, "BY": true,
		"CALL": true, "CASCADE": true, "CASE": true, "CHANGE": true, "CHAR": true,
		"CHARACTER": true, "CHECK": true, "COLLATE": true, "COLUMN": true, "CONDITION": true,
		"CONSTRAINT": true, "CONTINUE": true, "CONVERT": true, "CREATE": true, "CROSS": true,
		"CURRENT_DATE": true, "CURRENT_TIME": true, "CURRENT_TIMESTAMP": true, "CURRENT_USER": true,
		"CURSOR": true, "DATABASE": true, "DATABASES": true, "DAY_HOUR": true, "DAY_MICROSECOND": true,
		"DAY_MINUTE": true, "DAY_SECOND": true, "DEC": true, "DECIMAL": true, "DECLARE": true,
		"DEFAULT": true, "DELAYED": true, "DELETE": true, "DESC": true, "DESCRIBE": true,
		"DETERMINISTIC": true, "DISTINCT": true, "DISTINCTROW": true, "DIV": true, "DOUBLE": true,
		"DROP": true, "DUAL": true, "EACH": true, "ELSE": true, "ELSEIF": true,
		"ENCLOSED": true, "ESCAPED": true, "EXISTS": true, "EXIT": true, "EXPLAIN": true,
		"FALSE": true, "FETCH": true, "FLOAT": true, "FLOAT4": true, "FLOAT8": true,
		"FOR": true, "FORCE": true, "FOREIGN": true, "FROM": true, "FULLTEXT": true,
		"GRANT": true, "GROUP": true, "HAVING": true, "HIGH_PRIORITY": true, "HOUR_MICROSECOND": true,
		"HOUR_MINUTE": true, "HOUR_SECOND": true, "IF": true, "IGNORE": true, "IN": true,
		"INDEX": true, "INFILE": true, "INNER": true, "INOUT": true, "INSENSITIVE": true,
		"INSERT": true, "INT": true, "INT1": true, "INT2": true, "INT3": true,
		"INT4": true, "INT8": true, "INTEGER": true, "INTERVAL": true, "INTO": true,
		"IS": true, "ITERATE": true, "JOIN": true, "KEY": true, "KEYS": true,
		"KILL": true, "LEADING": true, "LEAVE": true, "LEFT": true, "LIKE": true,
		"LIMIT": true, "LINEAR": true, "LINES": true, "LOAD": true, "LOCALTIME": true,
		"LOCALTIMESTAMP": true, "LOCK": true, "LONG": true, "LONGBLOB": true, "LONGTEXT": true,
		"LOOP": true, "LOW_PRIORITY": true, "MATCH": true, "MEDIUMBLOB": true, "MEDIUMINT": true,
		"MEDIUMTEXT": true, "MIDDLEINT": true, "MINUTE_MICROSECOND": true, "MINUTE_SECOND": true,
		"MOD": true, "MODIFIES": true, "NATURAL": true, "NOT": true, "NO_WRITE_TO_BINLOG": true,
		"NULL": true, "NUMERIC": true, "ON": true, "OPTIMIZE": true, "OPTION": true,
		"OPTIONALLY": true, "OR": true, "ORDER": true, "OUT": true, "OUTER": true,
		"OUTFILE": true, "PRECISION": true, "PRIMARY": true, "PROCEDURE": true, "PURGE": true,
		"RANGE": true, "READ": true, "READS": true, "REAL": true, "REFERENCES": true,
		"REGEXP": true, "RELEASE": true, "RENAME": true, "REPEAT": true, "REPLACE": true,
		"REQUIRE": true, "RESTRICT": true, "RETURN": true, "REVOKE": true, "RIGHT": true,
		"RLIKE": true, "SCHEMA": true, "SCHEMAS": true, "SECOND_MICROSECOND": true, "SELECT": true,
		"SENSITIVE": true, "SEPARATOR": true, "SET": true, "SHOW": true, "SMALLINT": true,
		"SPATIAL": true, "SPECIFIC": true, "SQL": true, "SQLEXCEPTION": true, "SQLSTATE": true,
		"SQLWARNING": true, "SQL_BIG_RESULT": true, "SQL_CALC_FOUND_ROWS": true, "SQL_SMALL_RESULT": true,
		"SSL": true, "STARTING": true, "STRAIGHT_JOIN": true, "TABLE": true, "TERMINATED": true,
		"THEN": true, "TINYBLOB": true, "TINYINT": true, "TINYTEXT": true, "TO": true,
		"TRAILING": true, "TRIGGER": true, "TRUE": true, "UNDO": true, "UNION": true,
		"UNIQUE": true, "UNLOCK": true, "UNSIGNED": true, "UPDATE": true, "USAGE": true,
		"USE": true, "USING": true, "UTC_DATE": true, "UTC_TIME": true, "UTC_TIMESTAMP": true,
		"VALUES": true, "VARBINARY": true, "VARCHAR": true, "VARCHARACTER": true, "VARYING": true,
		"WHEN": true, "WHERE": true, "WHILE": true, "WITH": true, "WRITE": true,
		"X509": true, "XOR": true, "YEAR_MONTH": true, "ZEROFILL": true,
	}
	return reserved[strings.ToUpper(word)]
}

func (d *MySQLDialect) GetLimitSyntax() LimitSyntax {
	return LimitSyntaxStandard
}
