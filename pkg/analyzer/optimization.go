package analyzer

import (
	"strings"

	"github.com/Chahine-tech/sql-parser-go/pkg/dialect"
	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

// OptimizationEngine provides comprehensive SQL optimization suggestions
type OptimizationEngine struct {
	dialect dialect.Dialect
	rules   []OptimizationRule
}

// OptimizationRule defines a rule for optimization analysis
type OptimizationRule struct {
	ID          string
	Name        string
	Description string
	Category    string // PERFORMANCE, SECURITY, MAINTAINABILITY, BEST_PRACTICE
	Severity    string // INFO, WARNING, ERROR, CRITICAL
	Enabled     bool
	CheckFunc   func(*OptimizationEngine, parser.Statement) []EnhancedOptimizationSuggestion
}

// EnhancedOptimizationSuggestion extends the basic suggestion with more details
type EnhancedOptimizationSuggestion struct {
	Type          string `json:"type"`
	Description   string `json:"description"`
	Severity      string `json:"severity"`
	Category      string `json:"category"`
	Rule          string `json:"rule"`
	Line          int    `json:"line,omitempty"`
	Column        int    `json:"column,omitempty"`
	Table         string `json:"table,omitempty"`
	ColumnName    string `json:"column_name,omitempty"`
	Suggestion    string `json:"suggestion"`
	Impact        string `json:"impact"` // HIGH, MEDIUM, LOW
	Dialect       string `json:"dialect,omitempty"`
	AutoFixable   bool   `json:"auto_fixable"`
	FixSuggestion string `json:"fix_suggestion,omitempty"`
}

// NewOptimizationEngine creates a new optimization engine for a specific dialect
func NewOptimizationEngine(d dialect.Dialect) *OptimizationEngine {
	engine := &OptimizationEngine{
		dialect: d,
		rules:   make([]OptimizationRule, 0),
	}

	// Register all optimization rules
	engine.registerCoreRules()
	engine.registerPerformanceRules()
	engine.registerSecurityRules()
	engine.registerDialectSpecificRules()

	return engine
}

// AnalyzeOptimizations performs comprehensive optimization analysis
func (oe *OptimizationEngine) AnalyzeOptimizations(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	var suggestions []EnhancedOptimizationSuggestion

	for _, rule := range oe.rules {
		if rule.Enabled {
			ruleSuggestions := rule.CheckFunc(oe, stmt)
			suggestions = append(suggestions, ruleSuggestions...)
		}
	}

	return suggestions
}

// Core optimization rules
func (oe *OptimizationEngine) registerCoreRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "SELECT_STAR",
			Name:        "Avoid SELECT *",
			Description: "Using SELECT * can impact performance and maintainability",
			Category:    "PERFORMANCE",
			Severity:    "WARNING",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkSelectStar(stmt)
			},
		},
		{
			ID:          "MISSING_WHERE",
			Name:        "Missing WHERE clause",
			Description: "Queries without WHERE clause may return excessive data",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkMissingWhere(stmt)
			},
		},
		{
			ID:          "CARTESIAN_PRODUCT",
			Name:        "Potential Cartesian Product",
			Description: "Missing JOIN conditions can cause Cartesian products",
			Category:    "PERFORMANCE",
			Severity:    "CRITICAL",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkCartesianProduct(stmt)
			},
		},
		{
			ID:          "FUNCTION_IN_WHERE",
			Name:        "Function in WHERE clause",
			Description: "Functions on columns in WHERE prevent index usage",
			Category:    "PERFORMANCE",
			Severity:    "WARNING",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkFunctionInWhere(stmt)
			},
		},
	}...)
}

// Performance-specific rules
func (oe *OptimizationEngine) registerPerformanceRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "INEFFICIENT_SUBQUERY",
			Name:        "Inefficient Subquery",
			Description: "Subqueries in WHERE clause can often be optimized as JOINs",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkInefficientSubquery(stmt)
			},
		},
		{
			ID:          "UNNECESSARY_DISTINCT",
			Name:        "Unnecessary DISTINCT",
			Description: "DISTINCT may be unnecessary if data is already unique",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkUnnecessaryDistinct(stmt)
			},
		},
		{
			ID:          "MISSING_INDEX_HINT",
			Name:        "Index Suggestion",
			Description: "Columns in WHERE/JOIN conditions may benefit from indexes",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkIndexOpportunities(stmt)
			},
		},
		{
			ID:          "INEFFICIENT_JOIN_ORDER",
			Name:        "Join Order Optimization",
			Description: "Join order may impact query performance",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkJoinOrder(stmt)
			},
		},
	}...)
}

// Security-focused rules
func (oe *OptimizationEngine) registerSecurityRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "SQL_INJECTION_RISK",
			Name:        "SQL Injection Risk",
			Description: "Dynamic SQL construction may be vulnerable to injection",
			Category:    "SECURITY",
			Severity:    "CRITICAL",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkSQLInjectionRisk(stmt)
			},
		},
		{
			ID:          "OVERPRIVILEGED_SELECT",
			Name:        "Overprivileged Query",
			Description: "Query accesses more data than potentially needed",
			Category:    "SECURITY",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkOverprivilegedSelect(stmt)
			},
		},
	}...)
}

// Dialect-specific optimization rules
func (oe *OptimizationEngine) registerDialectSpecificRules() {
	dialectName := strings.ToLower(strings.ReplaceAll(oe.dialect.Name(), " ", ""))

	switch dialectName {
	case "mysql":
		oe.registerMySQLRules()
	case "postgresql":
		oe.registerPostgreSQLRules()
	case "sqlserver":
		oe.registerSQLServerRules()
	case "sqlite":
		oe.registerSQLiteRules()
	case "oracle":
		oe.registerOracleRules()
	}
}

// MySQL-specific rules
func (oe *OptimizationEngine) registerMySQLRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "MYSQL_LIMIT_OPTIMIZATION",
			Name:        "MySQL LIMIT Optimization",
			Description: "Use LIMIT with ORDER BY for better performance in MySQL",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkMySQLLimit(stmt)
			},
		},
		{
			ID:          "MYSQL_ENGINE_HINT",
			Name:        "MySQL Storage Engine",
			Description: "Consider InnoDB for transactional queries",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkMySQLEngine(stmt)
			},
		},
	}...)
}

// PostgreSQL-specific rules
func (oe *OptimizationEngine) registerPostgreSQLRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "POSTGRESQL_JSON_OPTIMIZATION",
			Name:        "PostgreSQL JSON Optimization",
			Description: "Use JSONB instead of JSON for better performance",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkPostgreSQLJSON(stmt)
			},
		},
		{
			ID:          "POSTGRESQL_ARRAY_OPTIMIZATION",
			Name:        "PostgreSQL Array Usage",
			Description: "Consider array operations for better performance",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkPostgreSQLArrays(stmt)
			},
		},
	}...)
}

// SQL Server-specific rules
func (oe *OptimizationEngine) registerSQLServerRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "SQLSERVER_TOP_OPTIMIZATION",
			Name:        "SQL Server TOP vs LIMIT",
			Description: "Use TOP instead of LIMIT for SQL Server compatibility",
			Category:    "PERFORMANCE",
			Severity:    "WARNING",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkSQLServerTop(stmt)
			},
		},
		{
			ID:          "SQLSERVER_NOLOCK_WARNING",
			Name:        "NOLOCK Hint Warning",
			Description: "NOLOCK can cause dirty reads and data inconsistency",
			Category:    "SECURITY",
			Severity:    "WARNING",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkSQLServerNoLock(stmt)
			},
		},
	}...)
}

// SQLite-specific rules
func (oe *OptimizationEngine) registerSQLiteRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "SQLITE_PRAGMA_OPTIMIZATION",
			Name:        "SQLite PRAGMA Suggestions",
			Description: "Consider PRAGMA settings for better performance",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkSQLitePragma(stmt)
			},
		},
	}...)
}

// Oracle-specific rules
func (oe *OptimizationEngine) registerOracleRules() {
	oe.rules = append(oe.rules, []OptimizationRule{
		{
			ID:          "ORACLE_HINT_OPTIMIZATION",
			Name:        "Oracle Hint Usage",
			Description: "Consider using Oracle hints for complex queries",
			Category:    "PERFORMANCE",
			Severity:    "INFO",
			Enabled:     true,
			CheckFunc: func(engine *OptimizationEngine, stmt parser.Statement) []EnhancedOptimizationSuggestion {
				return engine.checkOracleHints(stmt)
			},
		},
	}...)
}
