package analyzer

import (
	"fmt"
	"sync"

	"github.com/Chahine-tech/sql-parser-go/pkg/dialect"
	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

type Analyzer struct {
	analysis           QueryAnalysis
	cache              map[string]QueryAnalysis
	mu                 sync.RWMutex
	optimizationEngine *OptimizationEngine
}

func New() *Analyzer {
	return &Analyzer{
		analysis: QueryAnalysis{
			Tables:     make([]TableInfo, 0, 8),
			Columns:    make([]ColumnInfo, 0, 16),
			Joins:      make([]JoinInfo, 0, 4),
			Conditions: make([]ConditionInfo, 0, 8),
		},
		cache: make(map[string]QueryAnalysis, 64),
	}
}

func NewWithDialect(d dialect.Dialect) *Analyzer {
	return &Analyzer{
		analysis: QueryAnalysis{
			Tables:     make([]TableInfo, 0, 8),
			Columns:    make([]ColumnInfo, 0, 16),
			Joins:      make([]JoinInfo, 0, 4),
			Conditions: make([]ConditionInfo, 0, 8),
		},
		cache:              make(map[string]QueryAnalysis, 64),
		optimizationEngine: NewOptimizationEngine(d),
	}
}

func (a *Analyzer) Analyze(stmt parser.Statement) QueryAnalysis {
	a.analysis.Tables = a.analysis.Tables[:0]
	a.analysis.Columns = a.analysis.Columns[:0]
	a.analysis.Joins = a.analysis.Joins[:0]
	a.analysis.Conditions = a.analysis.Conditions[:0]

	switch s := stmt.(type) {
	case *parser.SelectStatement:
		a.analyzeSelectStatement(s)
		a.analysis.QueryType = "SELECT"
	case *parser.InsertStatement:
		a.analyzeInsertStatement(s)
		a.analysis.QueryType = "INSERT"
	case *parser.UpdateStatement:
		a.analyzeUpdateStatement(s)
		a.analysis.QueryType = "UPDATE"
	case *parser.DeleteStatement:
		a.analyzeDeleteStatement(s)
		a.analysis.QueryType = "DELETE"
	}

	a.analysis.Complexity = a.calculateComplexity()
	return a.analysis
}

func (a *Analyzer) AnalyzeWithCache(stmt parser.Statement, cacheKey string) QueryAnalysis {
	if cacheKey != "" {
		a.mu.RLock()
		if cached, exists := a.cache[cacheKey]; exists {
			a.mu.RUnlock()
			return cached
		}
		a.mu.RUnlock()
	}

	analysis := a.Analyze(stmt)

	if cacheKey != "" {
		a.mu.Lock()
		a.cache[cacheKey] = analysis
		a.mu.Unlock()
	}

	return analysis
}

func (a *Analyzer) analyzeSelectStatement(stmt *parser.SelectStatement) {
	if stmt.From != nil {
		for _, table := range stmt.From.Tables {
			a.analysis.Tables = append(a.analysis.Tables, TableInfo{
				Schema: table.Schema,
				Name:   table.Name,
				Alias:  table.Alias,
				Usage:  "SELECT",
			})
		}
	}

	for _, join := range stmt.Joins {
		a.analysis.Tables = append(a.analysis.Tables, TableInfo{
			Schema: join.Table.Schema,
			Name:   join.Table.Name,
			Alias:  join.Table.Alias,
			Usage:  "SELECT",
		})

		a.analysis.Joins = append(a.analysis.Joins, JoinInfo{
			Type:       join.JoinType,
			RightTable: join.Table.Name,
			Condition:  join.Condition.String(),
		})

		a.analyzeExpression(join.Condition, "JOIN")
	}

	for _, col := range stmt.Columns {
		a.analyzeExpression(col, "SELECT")
	}

	if stmt.Where != nil {
		a.analyzeExpression(stmt.Where, "WHERE")
	}

	for _, expr := range stmt.GroupBy {
		a.analyzeExpression(expr, "GROUP_BY")
	}

	if stmt.Having != nil {
		a.analyzeExpression(stmt.Having, "HAVING")
	}

	for _, orderBy := range stmt.OrderBy {
		a.analyzeExpression(orderBy.Expression, "ORDER_BY")
	}
}

func (a *Analyzer) analyzeExpression(expr parser.Expression, usage string) {
	switch e := expr.(type) {
	case *parser.ColumnReference:
		a.analysis.Columns = append(a.analysis.Columns, ColumnInfo{
			Table: e.Table,
			Name:  e.Column,
			Usage: usage,
		})
	case *parser.BinaryExpression:
		a.analyzeExpression(e.Left, usage)
		a.analyzeExpression(e.Right, usage)

		// Extract conditions for WHERE clauses
		if usage == "WHERE" || usage == "HAVING" || usage == "JOIN" {
			a.extractCondition(e, usage)
		}
	case *parser.FunctionCall:
		for _, arg := range e.Arguments {
			a.analyzeExpression(arg, usage)
		}
	case *parser.StarExpression:
		a.analysis.Columns = append(a.analysis.Columns, ColumnInfo{
			Table: e.Table,
			Name:  "*",
			Usage: usage,
		})
	case *parser.UnaryExpression:
		a.analyzeExpression(e.Operand, usage)
	case *parser.InExpression:
		a.analyzeExpression(e.Expression, usage)
		for _, val := range e.Values {
			a.analyzeExpression(val, usage)
		}
	}
}

func (a *Analyzer) extractCondition(expr *parser.BinaryExpression, _ string) {
	// Try to extract simple conditions like column = value
	if leftCol, ok := expr.Left.(*parser.ColumnReference); ok {
		if rightLit, ok := expr.Right.(*parser.Literal); ok {
			a.analysis.Conditions = append(a.analysis.Conditions, ConditionInfo{
				Table:    leftCol.Table,
				Column:   leftCol.Column,
				Operator: expr.Operator,
				Value:    fmt.Sprintf("%v", rightLit.Value),
			})
		}
	}
}

func (a *Analyzer) analyzeInsertStatement(stmt *parser.InsertStatement) {
	a.analysis.Tables = append(a.analysis.Tables, TableInfo{
		Schema: stmt.Table.Schema,
		Name:   stmt.Table.Name,
		Alias:  stmt.Table.Alias,
		Usage:  "INSERT",
	})

	// Analyze column list
	for _, col := range stmt.Columns {
		a.analysis.Columns = append(a.analysis.Columns, ColumnInfo{
			Name:  col,
			Usage: "INSERT",
		})
	}
}

func (a *Analyzer) analyzeUpdateStatement(stmt *parser.UpdateStatement) {
	a.analysis.Tables = append(a.analysis.Tables, TableInfo{
		Schema: stmt.Table.Schema,
		Name:   stmt.Table.Name,
		Alias:  stmt.Table.Alias,
		Usage:  "UPDATE",
	})

	// Analyze SET clause
	for _, assignment := range stmt.Set {
		a.analysis.Columns = append(a.analysis.Columns, ColumnInfo{
			Name:  assignment.Column,
			Usage: "UPDATE",
		})
		a.analyzeExpression(assignment.Value, "UPDATE")
	}

	// Analyze WHERE clause
	if stmt.Where != nil {
		a.analyzeExpression(stmt.Where, "WHERE")
	}
}

func (a *Analyzer) analyzeDeleteStatement(stmt *parser.DeleteStatement) {
	a.analysis.Tables = append(a.analysis.Tables, TableInfo{
		Schema: stmt.From.Schema,
		Name:   stmt.From.Name,
		Alias:  stmt.From.Alias,
		Usage:  "DELETE",
	})

	// Analyze WHERE clause
	if stmt.Where != nil {
		a.analyzeExpression(stmt.Where, "WHERE")
	}
}

func (a *Analyzer) calculateComplexity() int {
	complexity := 0

	// Base complexity
	complexity += len(a.analysis.Tables)
	complexity += len(a.analysis.Joins) * 2
	complexity += len(a.analysis.Conditions)

	// Additional complexity for functions
	for _, col := range a.analysis.Columns {
		if col.Usage == "SELECT" {
			complexity++
		}
	}

	return complexity
}

// SuggestOptimizations provides optimization suggestions for a query
func (a *Analyzer) SuggestOptimizations(stmt *parser.SelectStatement) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// Check for SELECT *
	if a.hasSelectAll(stmt) {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "SELECT_ALL",
			Description: "Avoid SELECT * for better performance. Specify only needed columns.",
			Severity:    "WARNING",
		})
	}

	// Check for missing JOIN conditions (potential Cartesian product)
	if a.hasCartesianProduct(stmt) {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "CARTESIAN_PRODUCT",
			Description: "Possible Cartesian product detected. Ensure proper JOIN conditions.",
			Severity:    "ERROR",
		})
	}

	// Check for large table count
	if len(a.analysis.Tables) > 5 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        "COMPLEX_QUERY",
			Description: "Query involves many tables. Consider breaking into smaller queries.",
			Severity:    "INFO",
		})
	}

	return suggestions
}

// GetEnhancedOptimizations returns comprehensive optimization suggestions using the new engine
func (a *Analyzer) GetEnhancedOptimizations(stmt parser.Statement) []EnhancedOptimizationSuggestion {
	if a.optimizationEngine == nil {
		// Fallback to basic suggestions if no optimization engine is available
		basicSuggestions := a.SuggestOptimizations(stmt.(*parser.SelectStatement))
		enhanced := make([]EnhancedOptimizationSuggestion, len(basicSuggestions))
		for i, basic := range basicSuggestions {
			enhanced[i] = EnhancedOptimizationSuggestion{
				Type:        basic.Type,
				Description: basic.Description,
				Severity:    basic.Severity,
				Category:    "GENERAL",
				Rule:        basic.Type,
				Line:        basic.Line,
				Suggestion:  "Review query for optimization opportunities",
				Impact:      "MEDIUM",
				AutoFixable: false,
			}
		}
		return enhanced
	}

	return a.optimizationEngine.AnalyzeOptimizations(stmt)
}

// SetOptimizationEngine allows setting a custom optimization engine
func (a *Analyzer) SetOptimizationEngine(engine *OptimizationEngine) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.optimizationEngine = engine
}

func (a *Analyzer) hasSelectAll(stmt *parser.SelectStatement) bool {
	for _, col := range stmt.Columns {
		if _, ok := col.(*parser.StarExpression); ok {
			return true
		}
	}
	return false
}

func (a *Analyzer) hasCartesianProduct(stmt *parser.SelectStatement) bool {
	tableCount := 0
	if stmt.From != nil {
		tableCount += len(stmt.From.Tables)
	}
	tableCount += len(stmt.Joins)

	// If we have multiple tables but no joins or WHERE conditions,
	// it might be a Cartesian product
	if tableCount > 1 && len(stmt.Joins) == 0 && stmt.Where == nil {
		return true
	}

	return false
}
