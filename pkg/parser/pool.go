package parser

import (
	"sync"
)

// Object pools for reducing garbage collection pressure
var (
	// Pool for SelectStatement objects
	selectStatementPool = sync.Pool{
		New: func() interface{} {
			return &SelectStatement{
				Columns: make([]Expression, 0, 8),
				Joins:   make([]*JoinClause, 0, 4),
			}
		},
	}

	// Pool for ColumnReference objects
	columnReferencePool = sync.Pool{
		New: func() interface{} {
			return &ColumnReference{}
		},
	}

	// Pool for BinaryExpression objects
	binaryExpressionPool = sync.Pool{
		New: func() interface{} {
			return &BinaryExpression{}
		},
	}

	// Pool for JoinClause objects
	joinClausePool = sync.Pool{
		New: func() interface{} {
			return &JoinClause{}
		},
	}
)

// GetSelectStatement gets a pooled SelectStatement
func GetSelectStatement() *SelectStatement {
	stmt := selectStatementPool.Get().(*SelectStatement)
	// Reset the statement
	stmt.Distinct = false
	stmt.Top = nil
	stmt.Columns = stmt.Columns[:0]
	stmt.From = nil
	stmt.Joins = stmt.Joins[:0]
	stmt.Where = nil
	stmt.GroupBy = nil
	stmt.Having = nil
	stmt.OrderBy = nil
	return stmt
}

// PutSelectStatement returns a SelectStatement to the pool
func PutSelectStatement(stmt *SelectStatement) {
	if stmt != nil {
		selectStatementPool.Put(stmt)
	}
}

// GetColumnReference gets a pooled ColumnReference
func GetColumnReference() *ColumnReference {
	return columnReferencePool.Get().(*ColumnReference)
}

// PutColumnReference returns a ColumnReference to the pool
func PutColumnReference(col *ColumnReference) {
	if col != nil {
		col.Table = ""
		col.Column = ""
		columnReferencePool.Put(col)
	}
}

// GetBinaryExpression gets a pooled BinaryExpression
func GetBinaryExpression() *BinaryExpression {
	return binaryExpressionPool.Get().(*BinaryExpression)
}

// PutBinaryExpression returns a BinaryExpression to the pool
func PutBinaryExpression(expr *BinaryExpression) {
	if expr != nil {
		expr.Left = nil
		expr.Right = nil
		expr.Operator = ""
		binaryExpressionPool.Put(expr)
	}
}

// GetJoinClause gets a pooled JoinClause
func GetJoinClause() *JoinClause {
	return joinClausePool.Get().(*JoinClause)
}

// PutJoinClause returns a JoinClause to the pool
func PutJoinClause(join *JoinClause) {
	if join != nil {
		join.JoinType = ""
		join.Table = TableReference{}
		join.Condition = nil
		joinClausePool.Put(join)
	}
}
