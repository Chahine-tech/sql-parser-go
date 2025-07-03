// Package parser provides SQL parsing functionality for SQL Server queries.
// It builds Abstract Syntax Trees (AST) from tokenized SQL input and supports
// SELECT, INSERT, UPDATE, and DELETE statements with complex joins and expressions.
package parser

import "fmt"

// ParseError represents errors that occur during SQL parsing.
// It provides detailed information about the location and nature of parsing errors.
type ParseError struct {
	Message string
	Line    int
	Column  int
	Token   string
}

// Error returns a formatted error message implementing the error interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %s (near '%s')",
		e.Line, e.Column, e.Message, e.Token)
}

// NewParseError creates a new ParseError with the given details.
func NewParseError(message, token string, line, column int) *ParseError {
	return &ParseError{
		Message: message,
		Line:    line,
		Column:  column,
		Token:   token,
	}
}

// SyntaxError represents syntax errors in SQL statements.
type SyntaxError struct {
	Expected string
	Found    string
	Line     int
	Column   int
}

// Error returns a formatted syntax error message.
func (e *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error at line %d, column %d: expected %s, found %s",
		e.Line, e.Column, e.Expected, e.Found)
}

// NewSyntaxError creates a new SyntaxError with the given details.
func NewSyntaxError(expected, found string, line, column int) *SyntaxError {
	return &SyntaxError{
		Expected: expected,
		Found:    found,
		Line:     line,
		Column:   column,
	}
}
