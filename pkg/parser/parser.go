// Package parser provides SQL parsing functionality for SQL Server queries.
package parser

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Chahine-tech/sql-parser-go/pkg/lexer"
)

type Parser struct {
	l *lexer.Lexer

	curToken  lexer.Token
	peekToken lexer.Token

	errors []string

	parseStartTime time.Time
	tokenCount     int

	ctx context.Context
}

func New(input string) *Parser {
	return NewWithContext(context.Background(), input)
}

func NewWithContext(ctx context.Context, input string) *Parser {
	l := lexer.New(input)
	p := &Parser{
		l:              l,
		errors:         make([]string, 0, 4),
		parseStartTime: time.Now(),
		ctx:            ctx,
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	select {
	case <-p.ctx.Done():
		p.errors = append(p.errors, "parsing cancelled due to timeout")
		return
	default:
		// Continue with normal token processing
	}

	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	p.tokenCount++
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	syntaxErr := NewSyntaxError(
		t.String(),
		p.peekToken.Type.String(),
		p.peekToken.Line,
		p.peekToken.Column,
	)
	p.errors = append(p.errors, syntaxErr.Error())
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) GetParseMetrics() map[string]interface{} {
	duration := time.Since(p.parseStartTime)
	return map[string]interface{}{
		"parse_duration_ms": duration.Milliseconds(),
		"tokens_processed":  p.tokenCount,
		"tokens_per_second": float64(p.tokenCount) / duration.Seconds(),
		"error_count":       len(p.errors),
	}
}

func (p *Parser) ParseStatement() (Statement, error) {
	switch p.curToken.Type {
	case lexer.SELECT:
		return p.parseSelectStatement()
	case lexer.INSERT:
		return p.parseInsertStatement()
	case lexer.UPDATE:
		return p.parseUpdateStatement()
	case lexer.DELETE:
		return p.parseDeleteStatement()
	default:
		return nil, fmt.Errorf("unsupported statement type: %s", p.curToken.Literal)
	}
}

// Parse SELECT statement
func (p *Parser) parseSelectStatement() (*SelectStatement, error) {
	stmt := GetSelectStatement() // Use object pool

	if !p.curTokenIs(lexer.SELECT) {
		PutSelectStatement(stmt)
		return nil, fmt.Errorf("expected SELECT, got %s", p.curToken.Literal)
	}

	p.nextToken()

	if p.curTokenIs(lexer.DISTINCT) {
		stmt.Distinct = true
		p.nextToken()
	}

	if p.curTokenIs(lexer.TOP) {
		topClause, err := p.parseTopClause()
		if err != nil {
			return nil, err
		}
		stmt.Top = topClause
	}

	columns, err := p.parseSelectList()
	if err != nil {
		return nil, err
	}
	stmt.Columns = columns

	if p.curTokenIs(lexer.FROM) {
		fromClause, err := p.parseFromClause()
		if err != nil {
			return nil, err
		}
		stmt.From = fromClause
	}

	for p.curTokenIs(lexer.JOIN) || p.curTokenIs(lexer.INNER) || p.curTokenIs(lexer.LEFT) || p.curTokenIs(lexer.RIGHT) || p.curTokenIs(lexer.FULL) {
		joinClause, err := p.parseJoinClause()
		if err != nil {
			return nil, err
		}
		stmt.Joins = append(stmt.Joins, joinClause)
	}

	if p.curTokenIs(lexer.WHERE) {
		p.nextToken()
		whereExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Where = whereExpr
	}

	if p.curTokenIs(lexer.GROUP) {
		groupBy, err := p.parseGroupByClause()
		if err != nil {
			return nil, err
		}
		stmt.GroupBy = groupBy
	}

	if p.curTokenIs(lexer.HAVING) {
		p.nextToken()
		havingExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Having = havingExpr
	}

	if p.curTokenIs(lexer.ORDER) {
		orderBy, err := p.parseOrderByClause()
		if err != nil {
			return nil, err
		}
		stmt.OrderBy = orderBy
	}

	return stmt, nil
}

func (p *Parser) parseTopClause() (*TopClause, error) {
	if !p.curTokenIs(lexer.TOP) {
		return nil, fmt.Errorf("expected TOP, got %s", p.curToken.Literal)
	}

	p.nextToken()

	if !p.curTokenIs(lexer.NUMBER) {
		return nil, fmt.Errorf("expected number after TOP, got %s", p.curToken.Literal)
	}

	count, err := strconv.Atoi(p.curToken.Literal)
	if err != nil {
		return nil, fmt.Errorf("invalid number in TOP clause: %s", p.curToken.Literal)
	}

	topClause := &TopClause{Count: count}
	p.nextToken()

	// Check for PERCENT
	if p.curTokenIs(lexer.IDENT) && strings.ToUpper(p.curToken.Literal) == "PERCENT" {
		topClause.Percent = true
		p.nextToken()
	}

	return topClause, nil
}

func (p *Parser) parseSelectList() ([]Expression, error) {
	var columns []Expression

	if p.curTokenIs(lexer.ASTERISK) {
		columns = append(columns, &StarExpression{})
		p.nextToken()
		return columns, nil
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	columns = append(columns, expr)

	for p.curTokenIs(lexer.COMMA) {
		p.nextToken()

		if p.curTokenIs(lexer.ASTERISK) {
			columns = append(columns, &StarExpression{})
			p.nextToken()
		} else {
			expr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			columns = append(columns, expr)
		}
	}

	return columns, nil
}

func (p *Parser) parseFromClause() (*FromClause, error) {
	if !p.curTokenIs(lexer.FROM) {
		return nil, fmt.Errorf("expected FROM, got %s", p.curToken.Literal)
	}

	p.nextToken()

	fromClause := &FromClause{}

	table, err := p.parseTableReference()
	if err != nil {
		return nil, err
	}
	fromClause.Tables = append(fromClause.Tables, *table)

	for p.curTokenIs(lexer.COMMA) {
		p.nextToken()
		table, err := p.parseTableReference()
		if err != nil {
			return nil, err
		}
		fromClause.Tables = append(fromClause.Tables, *table)
	}

	return fromClause, nil
}

func (p *Parser) parseTableReference() (*TableReference, error) {
	if !p.curTokenIs(lexer.IDENT) {
		return nil, fmt.Errorf("expected table name, got %s", p.curToken.Literal)
	}

	table := &TableReference{}

	firstIdent := p.curToken.Literal
	p.nextToken()

	if p.curTokenIs(lexer.DOT) {
		p.nextToken()
		if !p.curTokenIs(lexer.IDENT) {
			return nil, fmt.Errorf("expected table name after dot, got %s", p.curToken.Literal)
		}
		table.Schema = firstIdent
		table.Name = p.curToken.Literal
		p.nextToken()
	} else {
		table.Name = firstIdent
	}

	if p.curTokenIs(lexer.AS) {
		p.nextToken()
		if !p.curTokenIs(lexer.IDENT) {
			return nil, fmt.Errorf("expected alias after AS, got %s", p.curToken.Literal)
		}
		table.Alias = p.curToken.Literal
		p.nextToken()
	} else if p.curTokenIs(lexer.IDENT) {
		// Implicit alias (no AS keyword)
		table.Alias = p.curToken.Literal
		p.nextToken()
	}

	return table, nil
}

func (p *Parser) parseJoinClause() (*JoinClause, error) {
	joinClause := GetJoinClause()

	if p.curTokenIs(lexer.INNER) {
		joinClause.JoinType = "INNER"
		p.nextToken()
		if !p.expectPeek(lexer.JOIN) {
			PutJoinClause(joinClause)
			return nil, fmt.Errorf("expected JOIN after INNER")
		}
	} else if p.curTokenIs(lexer.LEFT) {
		joinClause.JoinType = "LEFT"
		p.nextToken()
		if !p.expectPeek(lexer.JOIN) {
			PutJoinClause(joinClause)
			return nil, fmt.Errorf("expected JOIN after LEFT")
		}
	} else if p.curTokenIs(lexer.RIGHT) {
		joinClause.JoinType = "RIGHT"
		p.nextToken()
		if !p.expectPeek(lexer.JOIN) {
			PutJoinClause(joinClause)
			return nil, fmt.Errorf("expected JOIN after RIGHT")
		}
	} else if p.curTokenIs(lexer.FULL) {
		joinClause.JoinType = "FULL"
		p.nextToken()
		if !p.expectPeek(lexer.JOIN) {
			PutJoinClause(joinClause)
			return nil, fmt.Errorf("expected JOIN after FULL")
		}
	} else if p.curTokenIs(lexer.JOIN) {
		joinClause.JoinType = "INNER"
		p.nextToken()
	}

	table, err := p.parseTableReference()
	if err != nil {
		PutJoinClause(joinClause)
		return nil, err
	}
	joinClause.Table = *table

	// Parse ON condition
	if !p.curTokenIs(lexer.ON) {
		PutJoinClause(joinClause) // Return to pool on error
		return nil, fmt.Errorf("expected ON after JOIN table, got %s", p.curToken.Literal)
	}

	p.nextToken()
	condition, err := p.parseExpression()
	if err != nil {
		PutJoinClause(joinClause) // Return to pool on error
		return nil, err
	}
	joinClause.Condition = condition

	return joinClause, nil
}

func (p *Parser) parseGroupByClause() ([]Expression, error) {
	if !p.curTokenIs(lexer.GROUP) {
		return nil, fmt.Errorf("expected GROUP, got %s", p.curToken.Literal)
	}

	p.nextToken()
	if !p.expectPeek(lexer.BY) {
		return nil, fmt.Errorf("expected BY after GROUP")
	}

	p.nextToken()

	var expressions []Expression

	// Parse first expression
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	expressions = append(expressions, expr)

	// Parse additional expressions
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (p *Parser) parseOrderByClause() ([]*OrderByClause, error) {
	if !p.curTokenIs(lexer.ORDER) {
		return nil, fmt.Errorf("expected ORDER, got %s", p.curToken.Literal)
	}

	p.nextToken()
	if !p.expectPeek(lexer.BY) {
		return nil, fmt.Errorf("expected BY after ORDER")
	}

	p.nextToken()

	var clauses []*OrderByClause

	// Parse first clause
	clause, err := p.parseOrderByItem()
	if err != nil {
		return nil, err
	}
	clauses = append(clauses, clause)

	// Parse additional clauses
	for p.curTokenIs(lexer.COMMA) {
		p.nextToken()
		clause, err := p.parseOrderByItem()
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
	}

	return clauses, nil
}

func (p *Parser) parseOrderByItem() (*OrderByClause, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	clause := &OrderByClause{
		Expression: expr,
		Direction:  "ASC", // Default
	}

	// Check for ASC/DESC
	if p.curTokenIs(lexer.IDENT) {
		direction := strings.ToUpper(p.curToken.Literal)
		if direction == "ASC" || direction == "DESC" {
			clause.Direction = direction
			p.nextToken()
		}
	}

	return clause, nil
}

// Basic expression parsing (simplified for now)
func (p *Parser) parseExpression() (Expression, error) {
	return p.parseInfixExpression()
}

func (p *Parser) parseInfixExpression() (Expression, error) {
	left, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for p.isInfixOperator(p.curToken.Type) {
		operator := p.curToken.Literal
		p.nextToken()

		right, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}

		expr := GetBinaryExpression() // Use object pool
		expr.Left = left
		expr.Operator = operator
		expr.Right = right
		left = expr
	}

	return left, nil
}

func (p *Parser) parsePrimaryExpression() (Expression, error) {
	switch p.curToken.Type {
	case lexer.IDENT:
		return p.parseIdentifierExpression()
	case lexer.NUMBER:
		return p.parseNumberLiteral()
	case lexer.STRING:
		return p.parseStringLiteral()
	case lexer.ASTERISK:
		expr := &StarExpression{}
		p.nextToken()
		return expr, nil
	case lexer.LPAREN:
		return p.parseGroupedExpression()
	default:
		return nil, fmt.Errorf("unexpected token in expression: %s", p.curToken.Literal)
	}
}

func (p *Parser) parseIdentifierExpression() (Expression, error) {
	firstIdent := p.curToken.Literal
	p.nextToken()

	// Check if it's a qualified column (table.column)
	if p.curTokenIs(lexer.DOT) {
		p.nextToken()
		if !p.curTokenIs(lexer.IDENT) && !p.curTokenIs(lexer.ASTERISK) {
			return nil, fmt.Errorf("expected column name after dot, got %s", p.curToken.Literal)
		}

		if p.curTokenIs(lexer.ASTERISK) {
			expr := &StarExpression{Table: firstIdent}
			p.nextToken()
			return expr, nil
		}

		expr := GetColumnReference() // Use object pool
		expr.Table = firstIdent
		expr.Column = p.curToken.Literal
		p.nextToken()
		return expr, nil
	}

	// Check if it's a function call
	if p.curTokenIs(lexer.LPAREN) {
		return p.parseFunctionCall(firstIdent)
	}

	// It's a simple column reference
	expr := GetColumnReference() // Use object pool
	expr.Column = firstIdent
	return expr, nil
}

func (p *Parser) parseFunctionCall(name string) (Expression, error) {
	if !p.curTokenIs(lexer.LPAREN) {
		return nil, fmt.Errorf("expected '(' for function call, got %s", p.curToken.Literal)
	}

	p.nextToken()

	var arguments []Expression

	if !p.curTokenIs(lexer.RPAREN) {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, arg)

		for p.curTokenIs(lexer.COMMA) {
			p.nextToken()
			arg, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, arg)
		}
	}

	if !p.curTokenIs(lexer.RPAREN) {
		return nil, fmt.Errorf("expected ')' to close function call, got %s", p.curToken.Literal)
	}

	p.nextToken() // consume the closing paren

	return &FunctionCall{
		Name:      name,
		Arguments: arguments,
	}, nil
}

func (p *Parser) parseNumberLiteral() (Expression, error) {
	literal := &Literal{}

	if strings.Contains(p.curToken.Literal, ".") {
		value, err := strconv.ParseFloat(p.curToken.Literal, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q as float", p.curToken.Literal)
		}
		literal.Value = value
	} else {
		value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q as integer", p.curToken.Literal)
		}
		literal.Value = value
	}

	p.nextToken()
	return literal, nil
}

func (p *Parser) parseStringLiteral() (Expression, error) {
	literal := &Literal{Value: p.curToken.Literal}
	p.nextToken()
	return literal, nil
}

func (p *Parser) parseGroupedExpression() (Expression, error) {
	p.nextToken()

	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil, fmt.Errorf("expected ')' to close grouped expression")
	}

	return exp, nil
}

func (p *Parser) isInfixOperator(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.ASSIGN, lexer.EQ, lexer.NOT_EQ, lexer.LT, lexer.GT, lexer.LTE, lexer.GTE,
		lexer.AND, lexer.OR, lexer.PLUS, lexer.MINUS, lexer.ASTERISK, lexer.SLASH,
		lexer.LIKE, lexer.IN:
		return true
	default:
		return false
	}
}

// Stub implementations for other statement types
func (p *Parser) parseInsertStatement() (*InsertStatement, error) {
	return nil, fmt.Errorf("INSERT statement parsing not implemented yet")
}

func (p *Parser) parseUpdateStatement() (*UpdateStatement, error) {
	return nil, fmt.Errorf("UPDATE statement parsing not implemented yet")
}

func (p *Parser) parseDeleteStatement() (*DeleteStatement, error) {
	return nil, fmt.Errorf("DELETE statement parsing not implemented yet")
}
