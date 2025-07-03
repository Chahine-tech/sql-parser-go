# SQL Parser Go - SQL Server Query Analysis Project

## Project Architecture

### Folder Structure
```
sql-parser-go/
├── cmd/
│   └── sqlparser/
│       └── main.go                 # CLI entry point
├── pkg/
│   ├── lexer/
│   │   ├── lexer.go               # SQL tokenization
│   │   └── tokens.go              # Defining tokens
│   ├── parser/
│   │   ├── parser.go              # Main parser
│   │   ├── ast.go                 # AST (Abstract Syntax Tree)
│   │   └── statements.go          # SQL statement types
│   ├── analyzer/
│   │   ├── analyzer.go            # Semantic analyzer
│   │   ├── extractor.go           # Table/column extraction
│   │   └── dependencies.go        # Dependency analysis
│   └── logger/
│       ├── parser.go              # Parsing SQL Server logs
│       └── formats.go             # Supported log formats
├── internal/
│   └── config/
│       └── config.go              # Configuration
├── examples/
│   ├── queries/                   # Examples of queries
│   └── logs/                      # Examples of logs
├── tests/
├── go.mod
├── go.sum
├── README.md
└── Makefile
```

## Main Components

### 1. Lexer (Tokenization)

```go
// pkg/lexer/tokens.go
package lexer

type TokenType int

const (
    ILLEGAL TokenType = iota
    EOF
    
    // Identifiers and literals
    IDENT   // table_name, column_name
    STRING  // 'hello'
    NUMBER  // 123, 123.45
    
    // SQL Keywords
    SELECT
    FROM
    WHERE
    JOIN
    INNER
    LEFT
    RIGHT
    ON
    GROUP
    BY
    ORDER
    HAVING
    
    // Operators
    ASSIGN // =
    EQ     // ==
    NOT_EQ // !=
    LT     // <
    GT     // >
    
    // Delimiters
    COMMA     // ,
    SEMICOLON // ;
    LPAREN    // (
    RPAREN    // )
    DOT       // .
    ASTERISK  // *
)

type Token struct {
    Type     TokenType
    Literal  string
    Position int
    Line     int
    Column   int
}
```

### 2. Parser AST

```go
// pkg/parser/ast.go
package parser

type Node interface {
    String() string
    Type() string
}

type Statement interface {
    Node
    statementNode()
}

type Expression interface {
    Node
    expressionNode()
}

// SELECT Statement
type SelectStatement struct {
    Columns    []Expression
    From       *FromClause
    Joins      []*JoinClause
    Where      Expression
    GroupBy    []Expression
    Having     Expression
    OrderBy    []Expression
}

type FromClause struct {
    Tables []TableReference
}

type TableReference struct {
    Schema string
    Name   string
    Alias  string
}

type JoinClause struct {
    Type      string // INNER, LEFT, RIGHT, FULL
    Table     TableReference
    Condition Expression
}
```

### 3. Analyzer and Extractor

```go
// pkg/analyzer/extractor.go
package analyzer

type QueryAnalysis struct {
    Tables     []TableInfo     `json:"tables"`
    Columns    []ColumnInfo    `json:"columns"`
    Joins      []JoinInfo      `json:"joins"`
    Conditions []ConditionInfo `json:"conditions"`
    QueryType  string          `json:"query_type"`
    Complexity int             `json:"complexity"`
}

type TableInfo struct {
    Schema string `json:"schema,omitempty"`
    Name   string `json:"name"`
    Alias  string `json:"alias,omitempty"`
    Usage  string `json:"usage"` // SELECT, UPDATE, DELETE, INSERT
}

type ColumnInfo struct {
    Table  string `json:"table,omitempty"`
    Name   string `json:"name"`
    Usage  string `json:"usage"` // SELECT, WHERE, JOIN, ORDER_BY, GROUP_BY
}

type JoinInfo struct {
    Type        string `json:"type"`
    LeftTable   string `json:"left_table"`
    RightTable  string `json:"right_table"`
    Condition   string `json:"condition"`
}
```

### 4. SQL Server Log Parser

```go
// pkg/logger/parser.go
package logger

import (
    "bufio"
    "regexp"
    "time"
)

type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Duration  int64     `json:"duration_ms"`
    Database  string    `json:"database"`
    User      string    `json:"user"`
    Query     string    `json:"query"`
    Reads     int64     `json:"logical_reads"`
    Writes    int64     `json:"writes"`
}

type SQLServerLogParser struct {
    patterns map[string]*regexp.Regexp
}

func NewSQLServerLogParser() *SQLServerLogParser {
    return &SQLServerLogParser{
        patterns: map[string]*regexp.Regexp{
            "profiler":      regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})\s+(.+)`),
            "extended_events": regexp.MustCompile(`<event name="sql_statement_completed".*?>`),
            "query_store":   regexp.MustCompile(`"query_sql_text":\s*"([^"]+)"`),
        },
    }
}
```

## CLI Interface

### 5. Command Line Interface

```go
// cmd/sqlparser/main.go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"
    
    "github.com/your-username/sql-parser-go/pkg/analyzer"
    "github.com/your-username/sql-parser-go/pkg/parser"
)

func main() {
    var (
        queryFile = flag.String("query", "", "File containing the SQL query")
        logFile   = flag.String("log", "", "SQL Server log file")
        output    = flag.String("output", "json", "Output format (json, table)")
        verbose   = flag.Bool("verbose", false, "Verbose mode")
    )
    flag.Parse()

    if *queryFile != "" {
        analyzeQuery(*queryFile, *output, *verbose)
    } else if *logFile != "" {
        parseLog(*logFile, *output, *verbose)
    } else {
        fmt.Println("Usage: sqlparser -query file.sql or -log logfile.log")
        os.Exit(1)
    }
}

func analyzeQuery(filename, format string, verbose bool) {
    // Read the file
    content, err := os.ReadFile(filename)
    if err != nil {
        panic(err)
    }

    // Parse the query
    p := parser.New(string(content))
    stmt, err := p.ParseStatement()
    if err != nil {
        panic(err)
    }

    // Analyze
    analyzer := analyzer.New()
    analysis := analyzer.Analyze(stmt)

    // Output
    if format == "json" {
        output, _ := json.MarshalIndent(analysis, "", "  ")
        fmt.Println(string(output))
    }
}
```

## Advanced Features

### 6. Pattern Detection and Optimizations

```go
// pkg/analyzer/optimizer.go
package analyzer

type OptimizationSuggestion struct {
    Type        string `json:"type"`
    Description string `json:"description"`
    Severity    string `json:"severity"`
    Line        int    `json:"line,omitempty"`
}

func (a *Analyzer) SuggestOptimizations(query *parser.SelectStatement) []OptimizationSuggestion {
    var suggestions []OptimizationSuggestion
    
    // SELECT * detection
    if a.hasSelectAll(query) {
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:        "SELECT_ALL",
            Description: "Avoid SELECT * for better performance",
            Severity:    "WARNING",
        })
    }
    
    // Detection of joins without potential indexes
    if a.hasCartesianProduct(query) {
        suggestions = append(suggestions, OptimizationSuggestion{
            Type:        "CARTESIAN_PRODUCT",
            Description: "Possible cartesian product detected",
            Severity:    "ERROR",
        })
    }
    
    return suggestions
}
```

## Usage

```bash
# Analyze a query
./sqlparser -query complex_query.sql -output json

# Parse a SQL Server log
./sqlparser -log sqlserver.log -output table -verbose

# Example JSON output
{
  "tables": [
    {"name": "Users", "alias": "u", "usage": "SELECT"},
    {"name": "Orders", "alias": "o", "usage": "SELECT"}
  ],
  "columns": [
    {"table": "u", "name": "user_id", "usage": "SELECT"},
    {"table": "o", "name": "order_date", "usage": "WHERE"}
  ],
  "joins": [
    {"type": "INNER", "left_table": "Users", "right_table": "Orders"}
  ],
  "complexity": 7
}
```

## Technologies and Dependencies

```go
// go.mod
module github.com/Chahine-tech/sql-parser-go

go 1.24.3

require (
    github.com/spf13/cobra 
    github.com/stretchr/testify
    gopkg.in/yaml.v3 
)
```

## Development Steps

1. **Phase 1**: Basic Lexer + Simple SELECT Parser
2. **Phase 2**: Support for JOIN, WHERE, GROUP BY
3. **Phase 3**: SQL Server log parser
4. **Phase 4**: Semantic analysis and metadata extraction
5. **Phase 5**: Optimization suggestions
6. **Phase 6**: Web interface (optional)

This project would give you a robust and extensible tool to analyze your complex SQL Server queries!