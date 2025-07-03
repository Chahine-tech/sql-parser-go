package tests

import (
	"testing"

	"github.com/Chahine-tech/sql-parser-go/pkg/lexer"
)

func TestNextToken(t *testing.T) {
	input := `SELECT name, age FROM users WHERE age > 21;`

	tests := []struct {
		expectedType    lexer.TokenType
		expectedLiteral string
	}{
		{lexer.SELECT, "SELECT"},
		{lexer.IDENT, "name"},
		{lexer.COMMA, ","},
		{lexer.IDENT, "age"},
		{lexer.FROM, "FROM"},
		{lexer.IDENT, "users"},
		{lexer.WHERE, "WHERE"},
		{lexer.IDENT, "age"},
		{lexer.GT, ">"},
		{lexer.NUMBER, "21"},
		{lexer.SEMICOLON, ";"},
		{lexer.EOF, ""},
	}

	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestTokenizeSQL(t *testing.T) {
	input := `SELECT * FROM users`
	tokens := lexer.TokenizeSQL(input)

	expectedTypes := []lexer.TokenType{lexer.SELECT, lexer.ASTERISK, lexer.FROM, lexer.IDENT, lexer.EOF}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("wrong number of tokens. expected=%d, got=%d", len(expectedTypes), len(tokens))
	}

	for i, expectedType := range expectedTypes {
		if tokens[i].Type != expectedType {
			t.Fatalf("token[%d] wrong type. expected=%q, got=%q", i, expectedType, tokens[i].Type)
		}
	}
}

func TestCommentHandling(t *testing.T) {
	input := `-- This is a comment
SELECT * FROM users`

	l := lexer.New(input)

	// Should skip the comment and get SELECT token
	tok := l.NextToken()
	if tok.Type != lexer.SELECT {
		t.Fatalf("expected SELECT token, got %q", tok.Type)
	}
}

func TestStringLiterals(t *testing.T) {
	input := `SELECT 'hello world' FROM users`

	l := lexer.New(input)

	l.NextToken()        // SELECT
	tok := l.NextToken() // 'hello world'

	if tok.Type != lexer.STRING {
		t.Fatalf("expected STRING token, got %q", tok.Type)
	}

	if tok.Literal != "hello world" {
		t.Fatalf("expected 'hello world', got %q", tok.Literal)
	}
}
