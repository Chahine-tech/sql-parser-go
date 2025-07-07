// Package lexer provides SQL tokenization functionality.
package lexer

import (
	"strings"

	"github.com/Chahine-tech/sql-parser-go/pkg/dialect"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
	dialect      dialect.Dialect
}

func New(input string) *Lexer {
	return NewWithDialect(input, dialect.GetDialect("sqlserver")) // Default to SQL Server
}

func NewWithDialect(input string, d dialect.Dialect) *Lexer {
	l := &Lexer{
		input:   input,
		line:    1,
		column:  0,
		dialect: d,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// lookupIdent checks if an identifier is a keyword using the dialect
func (l *Lexer) lookupIdent(ident string) TokenType {
	// First check common SQL keywords
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	// If not found in common keywords, check if it's a dialect-specific reserved word
	if l.dialect.IsReservedWord(ident) {
		// For dialect-specific keywords, we'll treat them as identifiers for now
		// but could extend this to have dialect-specific token types
		return IDENT
	}

	return IDENT
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	// Handle line comments
	if l.ch == '-' && l.peekChar() == '-' {
		l.skipLineComment()
		return l.NextToken()
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: EQ, Literal: literal, Position: l.position, Line: l.line, Column: l.column}
		} else {
			tok = newToken(ASSIGN, l.ch, l.position, l.line, l.column)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: NOT_EQ, Literal: literal, Position: l.position, Line: l.line, Column: l.column}
		} else {
			tok = newToken(ILLEGAL, l.ch, l.position, l.line, l.column)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: LTE, Literal: literal, Position: l.position, Line: l.line, Column: l.column}
		} else {
			tok = newToken(LT, l.ch, l.position, l.line, l.column)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: GTE, Literal: literal, Position: l.position, Line: l.line, Column: l.column}
		} else {
			tok = newToken(GT, l.ch, l.position, l.line, l.column)
		}
	case ',':
		tok = newToken(COMMA, l.ch, l.position, l.line, l.column)
	case ';':
		tok = newToken(SEMICOLON, l.ch, l.position, l.line, l.column)
	case '(':
		tok = newToken(LPAREN, l.ch, l.position, l.line, l.column)
	case ')':
		tok = newToken(RPAREN, l.ch, l.position, l.line, l.column)
	case '.':
		tok = newToken(DOT, l.ch, l.position, l.line, l.column)
	case '*':
		tok = newToken(ASTERISK, l.ch, l.position, l.line, l.column)
	case '+':
		tok = newToken(PLUS, l.ch, l.position, l.line, l.column)
	case '-':
		tok = newToken(MINUS, l.ch, l.position, l.line, l.column)
	case '/':
		tok = newToken(SLASH, l.ch, l.position, l.line, l.column)
	case '%':
		tok = newToken(PERCENT, l.ch, l.position, l.line, l.column)
	case '\'':
		tok.Type = STRING
		tok.Literal = l.readString()
		tok.Position = l.position
		tok.Line = l.line
		tok.Column = l.column
	case '"':
		// Double quotes can be string literals or identifiers depending on dialect
		if l.dialect.Name() == "PostgreSQL" || l.dialect.Name() == "SQLite" || l.dialect.Name() == "Oracle" {
			tok.Type = IDENT
			tok.Literal = l.readDoubleQuotedIdentifier()
		} else {
			tok.Type = STRING
			tok.Literal = l.readDoubleQuotedString()
		}
		tok.Position = l.position
		tok.Line = l.line
		tok.Column = l.column
	case '`':
		// Backticks are MySQL-specific quoted identifiers
		if l.dialect.Name() == "MySQL" {
			tok.Type = IDENT
			tok.Literal = l.readBacktickIdentifier()
		} else {
			tok = newToken(ILLEGAL, l.ch, l.position, l.line, l.column)
		}
		tok.Position = l.position
		tok.Line = l.line
		tok.Column = l.column
	case '[':
		// Brackets are SQL Server-specific quoted identifiers
		if l.dialect.Name() == "SQL Server" {
			tok.Type = IDENT
			tok.Literal = l.readBracketedIdentifier()
		} else {
			tok = newToken(ILLEGAL, l.ch, l.position, l.line, l.column)
		}
		tok.Position = l.position
		tok.Line = l.line
		tok.Column = l.column
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Position = l.position
		tok.Line = l.line
		tok.Column = l.column
	default:
		if isLetter(l.ch) {
			tok.Position = l.position
			tok.Line = l.line
			tok.Column = l.column
			tok.Literal = l.readIdentifier()
			tok.Type = l.lookupIdent(strings.ToUpper(tok.Literal))
			return tok
		} else if isDigit(l.ch) {
			tok.Type = NUMBER
			tok.Literal = l.readNumber()
			tok.Position = l.position
			tok.Line = l.line
			tok.Column = l.column
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch, l.position, l.line, l.column)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readBracketedIdentifier() string {
	l.readChar()
	position := l.position
	for l.ch != ']' && l.ch != 0 {
		l.readChar()
	}
	result := l.input[position:l.position]
	if l.ch == ']' {
		l.readChar()
	}
	return result
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	l.readChar() // skip the opening quote
	position := l.position
	for l.ch != '\'' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // skip escape character
		}
		l.readChar()
	}
	result := l.input[position:l.position]
	return result
}

func (l *Lexer) readDoubleQuotedString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readDoubleQuotedIdentifier() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readBacktickIdentifier() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '`' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType TokenType, ch byte, pos, line, col int) Token {
	return Token{Type: tokenType, Literal: string(ch), Position: pos, Line: line, Column: col}
}

// TokenizeSQL tokenizes a complete SQL string and returns all tokens
// Optimized version with pre-allocated slice
func TokenizeSQL(input string) []Token {
	lexer := New(input)
	// Pre-allocate with estimated capacity based on input length
	estimatedTokens := len(input) / 6 // rough estimate: avg 6 chars per token
	if estimatedTokens < 10 {
		estimatedTokens = 10
	}
	tokens := make([]Token, 0, estimatedTokens)

	for {
		token := lexer.NextToken()
		tokens = append(tokens, token)
		if token.Type == EOF {
			break
		}
	}

	return tokens
}

// TokenizeWithBuffer reuses a provided buffer for better memory efficiency
func TokenizeWithBuffer(input string, buffer []Token) []Token {
	lexer := New(input)
	tokens := buffer[:0] // reset slice but keep capacity

	for {
		token := lexer.NextToken()
		tokens = append(tokens, token)
		if token.Type == EOF {
			break
		}
	}

	return tokens
}
