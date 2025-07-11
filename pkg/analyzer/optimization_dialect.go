package analyzer

import (
	"strings"

	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

// Dialect-specific optimization rule implementations

// MySQL-specific optimizations
func (oe *OptimizationEngine) checkMySQLLimit(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		if selectStmt.Limit != nil && selectStmt.OrderBy == nil {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "MYSQL_LIMIT_WITHOUT_ORDER",
				Description:   "LIMIT without ORDER BY may return inconsistent results in MySQL",
				Severity:      "WARNING",
				Category:      "PERFORMANCE",
				Rule:          "MYSQL_LIMIT_OPTIMIZATION",
				Dialect:       "mysql",
				Suggestion:    "Add ORDER BY clause with LIMIT",
				Impact:        "MEDIUM",
				AutoFixable:   false,
				FixSuggestion: "Always use ORDER BY with LIMIT to ensure consistent results: SELECT ... ORDER BY column LIMIT n",
			})
		}
	}

	return suggestions
}

func (oe *OptimizationEngine) checkMySQLEngine(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	// For complex queries with JOINs, suggest InnoDB
	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		if len(selectStmt.Joins) > 1 {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "MYSQL_STORAGE_ENGINE",
				Description:   "Complex queries with multiple JOINs may benefit from InnoDB storage engine",
				Severity:      "INFO",
				Category:      "PERFORMANCE",
				Rule:          "MYSQL_ENGINE_HINT",
				Dialect:       "mysql",
				Suggestion:    "Consider using InnoDB for transactional queries",
				Impact:        "LOW",
				AutoFixable:   false,
				FixSuggestion: "Ensure tables use InnoDB storage engine for better JOIN performance and ACID compliance",
			})
		}
	}

	return suggestions
}

// PostgreSQL-specific optimizations
func (oe *OptimizationEngine) checkPostgreSQLJSON(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		// Check for JSON operations in WHERE clause
		if selectStmt.Where != nil {
			whereStr := strings.ToUpper(selectStmt.Where.String())
			if strings.Contains(whereStr, "JSON") && !strings.Contains(whereStr, "JSONB") {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "POSTGRESQL_JSON_TYPE",
					Description:   "Consider using JSONB instead of JSON for better performance in PostgreSQL",
					Severity:      "INFO",
					Category:      "PERFORMANCE",
					Rule:          "POSTGRESQL_JSON_OPTIMIZATION",
					Dialect:       "postgresql",
					Suggestion:    "Use JSONB for JSON operations",
					Impact:        "MEDIUM",
					AutoFixable:   false,
					FixSuggestion: "Change JSON columns to JSONB for better indexing and query performance",
				})
			}
		}
	}

	return suggestions
}

func (oe *OptimizationEngine) checkPostgreSQLArrays(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		// Check for IN clauses with many values
		if selectStmt.Where != nil {
			whereStr := strings.ToUpper(selectStmt.Where.String())
			if strings.Contains(whereStr, " IN (") {
				// Count commas to estimate number of values
				inPart := whereStr[strings.Index(whereStr, " IN ("):]
				commaCount := strings.Count(inPart, ",")
				if commaCount > 5 {
					suggestions = append(suggestions, EnhancedOptimizationSuggestion{
						Type:          "POSTGRESQL_ARRAY_USAGE",
						Description:   "Large IN clause may benefit from PostgreSQL array operations",
						Severity:      "INFO",
						Category:      "PERFORMANCE",
						Rule:          "POSTGRESQL_ARRAY_OPTIMIZATION",
						Dialect:       "postgresql",
						Suggestion:    "Consider using array operations",
						Impact:        "MEDIUM",
						AutoFixable:   false,
						FixSuggestion: "Use ANY(ARRAY[...]) instead of large IN clauses for better performance",
					})
				}
			}
		}
	}

	return suggestions
}

// SQL Server-specific optimizations
func (oe *OptimizationEngine) checkSQLServerTop(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		if selectStmt.Limit != nil {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "SQLSERVER_TOP_VS_LIMIT",
				Description:   "Use TOP instead of LIMIT for SQL Server compatibility",
				Severity:      "WARNING",
				Category:      "PERFORMANCE",
				Rule:          "SQLSERVER_TOP_OPTIMIZATION",
				Dialect:       "sqlserver",
				Suggestion:    "Replace LIMIT with TOP",
				Impact:        "LOW",
				AutoFixable:   true,
				FixSuggestion: "Use SELECT TOP n instead of SELECT ... LIMIT n",
			})
		}
	}

	return suggestions
}

func (oe *OptimizationEngine) checkSQLServerNoLock(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	// Check for NOLOCK hints in the query
	queryStr := strings.ToUpper(stmt.String())
	if strings.Contains(queryStr, "NOLOCK") || strings.Contains(queryStr, "WITH (NOLOCK)") {
		suggestions = append(suggestions, EnhancedOptimizationSuggestion{
			Type:          "SQLSERVER_NOLOCK_WARNING",
			Description:   "NOLOCK hint can cause dirty reads and data inconsistency",
			Severity:      "WARNING",
			Category:      "SECURITY",
			Rule:          "SQLSERVER_NOLOCK_WARNING",
			Dialect:       "sqlserver",
			Suggestion:    "Consider alternatives to NOLOCK",
			Impact:        "HIGH",
			AutoFixable:   false,
			FixSuggestion: "Use READ UNCOMMITTED isolation level or consider if dirty reads are acceptable",
		})
	}

	return suggestions
}

// SQLite-specific optimizations
func (oe *OptimizationEngine) checkSQLitePragma(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	// For complex queries, suggest PRAGMA optimizations
	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		if len(selectStmt.Joins) >= 2 || (selectStmt.Where != nil && len(selectStmt.Where.String()) > 100) {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "SQLITE_PRAGMA_SUGGESTION",
				Description:   "Complex queries may benefit from SQLite PRAGMA optimizations",
				Severity:      "INFO",
				Category:      "PERFORMANCE",
				Rule:          "SQLITE_PRAGMA_OPTIMIZATION",
				Dialect:       "sqlite",
				Suggestion:    "Consider SQLite PRAGMA settings",
				Impact:        "MEDIUM",
				AutoFixable:   false,
				FixSuggestion: "Use PRAGMA query_planner = ON and PRAGMA optimize before complex queries",
			})
		}
	}

	return suggestions
}

// Oracle-specific optimizations
func (oe *OptimizationEngine) checkOracleHints(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		// For complex queries with multiple tables, suggest considering hints
		tableCount := 0
		if selectStmt.From != nil {
			tableCount += len(selectStmt.From.Tables)
		}
		tableCount += len(selectStmt.Joins)

		if tableCount > 3 {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "ORACLE_HINT_SUGGESTION",
				Description:   "Complex multi-table query may benefit from Oracle optimizer hints",
				Severity:      "INFO",
				Category:      "PERFORMANCE",
				Rule:          "ORACLE_HINT_OPTIMIZATION",
				Dialect:       "oracle",
				Suggestion:    "Consider using Oracle optimizer hints",
				Impact:        "MEDIUM",
				AutoFixable:   false,
				FixSuggestion: "Consider hints like /*+ USE_INDEX */ or /*+ LEADING */ for complex queries",
			})
		}
	}

	return suggestions
}
