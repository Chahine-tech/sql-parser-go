package analyzer

import (
	"fmt"
	"strings"

	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

// Core optimization rule implementations

// checkSelectStar detects usage of SELECT *
func (oe *OptimizationEngine) checkSelectStar(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		for _, col := range selectStmt.Columns {
			if _, isStar := col.(*parser.StarExpression); isStar {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "SELECT_STAR",
					Description:   "Avoid SELECT * - specify needed columns explicitly for better performance",
					Severity:      "WARNING",
					Category:      "PERFORMANCE",
					Rule:          "SELECT_STAR",
					Suggestion:    "Replace SELECT * with specific column names",
					Impact:        "MEDIUM",
					AutoFixable:   false,
					FixSuggestion: "List only the columns you actually need: SELECT col1, col2, col3 FROM ...",
				})
				break
			}
		}
	}

	return suggestions
}

// checkMissingWhere detects queries without WHERE clauses
func (oe *OptimizationEngine) checkMissingWhere(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		if selectStmt.Where == nil {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "MISSING_WHERE",
				Description:   "Consider adding WHERE clause to limit result set",
				Severity:      "INFO",
				Category:      "PERFORMANCE",
				Rule:          "MISSING_WHERE",
				Suggestion:    "Add WHERE clause to filter data",
				Impact:        "HIGH",
				AutoFixable:   false,
				FixSuggestion: "Add a WHERE clause to limit the number of rows returned",
			})
		}
	}

	return suggestions
}

// checkCartesianProduct detects potential Cartesian products
func (oe *OptimizationEngine) checkCartesianProduct(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		tableCount := 0
		if selectStmt.From != nil {
			tableCount += len(selectStmt.From.Tables)
		}

		joinCount := len(selectStmt.Joins)

		// If we have multiple tables but no proper joins, potential Cartesian product
		if tableCount > 1 && joinCount == 0 {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "CARTESIAN_PRODUCT",
				Description:   "Potential Cartesian product detected - missing JOIN conditions",
				Severity:      "CRITICAL",
				Category:      "PERFORMANCE",
				Rule:          "CARTESIAN_PRODUCT",
				Suggestion:    "Add proper JOIN conditions between tables",
				Impact:        "HIGH",
				AutoFixable:   false,
				FixSuggestion: "Use explicit JOIN syntax with ON conditions instead of comma-separated tables",
			})
		}

		// Check for JOINs without conditions
		for _, join := range selectStmt.Joins {
			if join.Condition == nil {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "CARTESIAN_PRODUCT",
					Description:   fmt.Sprintf("JOIN without condition detected for table '%s'", join.Table.Name),
					Severity:      "CRITICAL",
					Category:      "PERFORMANCE",
					Rule:          "CARTESIAN_PRODUCT",
					Table:         join.Table.Name,
					Suggestion:    "Add JOIN condition",
					Impact:        "HIGH",
					AutoFixable:   false,
					FixSuggestion: "Add an ON condition to specify how tables should be joined",
				})
			}
		}
	}

	return suggestions
}

// checkFunctionInWhere detects functions applied to columns in WHERE clauses
func (oe *OptimizationEngine) checkFunctionInWhere(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok && selectStmt.Where != nil {
		// This is a simplified check - in a real implementation, you'd parse the WHERE expression
		whereStr := selectStmt.Where.String()

		// Common functions that prevent index usage
		functionPatterns := []string{"UPPER(", "LOWER(", "TRIM(", "SUBSTRING(", "DATEPART(", "YEAR(", "MONTH(", "DAY("}

		for _, pattern := range functionPatterns {
			if strings.Contains(strings.ToUpper(whereStr), pattern) {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "FUNCTION_IN_WHERE",
					Description:   fmt.Sprintf("Function %s in WHERE clause prevents index usage", strings.TrimSuffix(pattern, "(")),
					Severity:      "WARNING",
					Category:      "PERFORMANCE",
					Rule:          "FUNCTION_IN_WHERE",
					Suggestion:    "Avoid functions on columns in WHERE clause",
					Impact:        "MEDIUM",
					AutoFixable:   false,
					FixSuggestion: "Consider using computed columns or restructuring the query to avoid functions on indexed columns",
				})
			}
		}
	}

	return suggestions
}

// checkInefficientSubquery detects subqueries that could be optimized
func (oe *OptimizationEngine) checkInefficientSubquery(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok && selectStmt.Where != nil {
		// Check for InExpression with subqueries
		if inExpr, ok := selectStmt.Where.(*parser.InExpression); ok {
			// Check if any of the values in the IN expression are subqueries
			for _, value := range inExpr.Values {
				if _, isSubquery := value.(*parser.SubqueryExpression); isSubquery {
					suggestions = append(suggestions, EnhancedOptimizationSuggestion{
						Type:          "INEFFICIENT_SUBQUERY",
						Description:   "Subquery in WHERE clause may be optimized as a JOIN",
						Severity:      "INFO",
						Category:      "PERFORMANCE",
						Rule:          "INEFFICIENT_SUBQUERY",
						Suggestion:    "Consider converting subquery to JOIN",
						Impact:        "MEDIUM",
						AutoFixable:   false,
						FixSuggestion: "Replace correlated subqueries with JOINs when possible for better performance",
					})
					break // Only need to suggest once per statement
				}
			}
		}

		// Also check for string patterns as fallback (for cases we might have missed)
		whereStr := selectStmt.Where.String()
		if strings.Contains(strings.ToUpper(whereStr), " EXISTS (SELECT") {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "INEFFICIENT_SUBQUERY",
				Description:   "EXISTS subquery in WHERE clause may be optimized",
				Severity:      "INFO",
				Category:      "PERFORMANCE",
				Rule:          "INEFFICIENT_SUBQUERY",
				Suggestion:    "Consider converting EXISTS subquery to JOIN",
				Impact:        "MEDIUM",
				AutoFixable:   false,
				FixSuggestion: "Replace EXISTS subqueries with JOINs when possible for better performance",
			})
		}
	}

	return suggestions
}

// checkUnnecessaryDistinct detects potentially unnecessary DISTINCT usage
func (oe *OptimizationEngine) checkUnnecessaryDistinct(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		if selectStmt.Distinct {
			// Check if there are JOINs that might create duplicates
			if len(selectStmt.Joins) == 0 {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "UNNECESSARY_DISTINCT",
					Description:   "DISTINCT may be unnecessary without JOINs",
					Severity:      "INFO",
					Category:      "PERFORMANCE",
					Rule:          "UNNECESSARY_DISTINCT",
					Suggestion:    "Review if DISTINCT is needed",
					Impact:        "LOW",
					AutoFixable:   true,
					FixSuggestion: "Remove DISTINCT if the query doesn't produce duplicate rows",
				})
			}
		}
	}

	return suggestions
}

// checkIndexOpportunities suggests potential index opportunities
func (oe *OptimizationEngine) checkIndexOpportunities(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		// Analyze WHERE clause for potential index candidates
		if selectStmt.Where != nil {
			whereStr := selectStmt.Where.String()

			// This is simplified - in practice, you'd parse the expression tree
			if strings.Contains(whereStr, "=") || strings.Contains(whereStr, ">") || strings.Contains(whereStr, "<") {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "INDEX_SUGGESTION",
					Description:   "Columns in WHERE clause may benefit from indexes",
					Severity:      "INFO",
					Category:      "PERFORMANCE",
					Rule:          "INDEX_SUGGESTION",
					Suggestion:    "Consider adding indexes on frequently queried columns",
					Impact:        "HIGH",
					AutoFixable:   false,
					FixSuggestion: "Analyze query execution plan and consider creating indexes on columns used in WHERE, JOIN, and ORDER BY clauses",
				})
			}
		}

		// Check JOIN conditions
		for _, join := range selectStmt.Joins {
			if join.Condition != nil {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "INDEX_SUGGESTION",
					Description:   fmt.Sprintf("JOIN condition on table '%s' may benefit from index", join.Table.Name),
					Severity:      "INFO",
					Category:      "PERFORMANCE",
					Rule:          "INDEX_SUGGESTION",
					Table:         join.Table.Name,
					Suggestion:    "Consider adding index on JOIN columns",
					Impact:        "HIGH",
					AutoFixable:   false,
					FixSuggestion: "Create indexes on columns used in JOIN conditions for better performance",
				})
			}
		}
	}

	return suggestions
}

// checkJoinOrder analyzes JOIN order for optimization opportunities
func (oe *OptimizationEngine) checkJoinOrder(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		if len(selectStmt.Joins) > 2 {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "JOIN_ORDER",
				Description:   "Complex JOIN query - consider optimizing JOIN order",
				Severity:      "INFO",
				Category:      "PERFORMANCE",
				Rule:          "INEFFICIENT_JOIN_ORDER",
				Suggestion:    "Review JOIN order for performance",
				Impact:        "MEDIUM",
				AutoFixable:   false,
				FixSuggestion: "Order JOINs to process smaller result sets first, starting with the most selective conditions",
			})
		}
	}

	return suggestions
}

// Security-related checks

// checkSQLInjectionRisk detects potential SQL injection vulnerabilities
func (oe *OptimizationEngine) checkSQLInjectionRisk(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	// This is a basic check - in practice, you'd need more sophisticated analysis
	// For now, we'll check for common patterns that might indicate dynamic SQL

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok && selectStmt.Where != nil {
		whereStr := selectStmt.Where.String()

		// Check for potential string concatenation patterns
		suspiciousPatterns := []string{"'+'", "CONCAT(", "||"}

		for _, pattern := range suspiciousPatterns {
			if strings.Contains(strings.ToUpper(whereStr), pattern) {
				suggestions = append(suggestions, EnhancedOptimizationSuggestion{
					Type:          "SQL_INJECTION_RISK",
					Description:   "Potential SQL injection risk detected in WHERE clause",
					Severity:      "CRITICAL",
					Category:      "SECURITY",
					Rule:          "SQL_INJECTION_RISK",
					Suggestion:    "Use parameterized queries",
					Impact:        "HIGH",
					AutoFixable:   false,
					FixSuggestion: "Replace string concatenation with parameterized queries or prepared statements",
				})
			}
		}
	}

	return suggestions
}

// checkOverprivilegedSelect detects queries that may access too much data
func (oe *OptimizationEngine) checkOverprivilegedSelect(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
		// Check for SELECT * combined with no WHERE clause
		hasStar := false
		for _, col := range selectStmt.Columns {
			if _, isStar := col.(*parser.StarExpression); isStar {
				hasStar = true
				break
			}
		}

		if hasStar && selectStmt.Where == nil {
			suggestions = append(suggestions, EnhancedOptimizationSuggestion{
				Type:          "OVERPRIVILEGED_SELECT",
				Description:   "Query accesses all columns and all rows - potential security/performance risk",
				Severity:      "WARNING",
				Category:      "SECURITY",
				Rule:          "OVERPRIVILEGED_SELECT",
				Suggestion:    "Limit both columns and rows accessed",
				Impact:        "MEDIUM",
				AutoFixable:   false,
				FixSuggestion: "Specify only needed columns and add appropriate WHERE conditions",
			})
		}
	}

	return suggestions
}
