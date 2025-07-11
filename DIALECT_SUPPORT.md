# SQL Dialect Support

This SQL Parser supports multiple SQL dialects with dialect-specific features and syntax parsing.

## Supported Dialects

### MySQL
- **Identifier quoting**: Backticks (`` ` ``)
- **Features**: JSON support, Common Table Expressions (CTE), Window Functions, Full-text search
- **Usage**: `--dialect mysql`

### PostgreSQL
- **Identifier quoting**: Double quotes (`"`)
- **Features**: Array support, JSON/JSONB, XML, RETURNING clause, Advanced window functions
- **Usage**: `--dialect postgresql` or `--dialect postgres`

### SQL Server
- **Identifier quoting**: Square brackets (`[]`)
- **Features**: XML support, TOP clause, MERGE statement, Advanced analytics
- **Usage**: `--dialect sqlserver` or `--dialect mssql`

### SQLite
- **Identifier quoting**: Double quotes (`"`)
- **Features**: JSON support (3.38+), Window functions (3.25+), Full-text search
- **Usage**: `--dialect sqlite`

### Oracle
- **Identifier quoting**: Double quotes (`"`)
- **Features**: XML support, Advanced partitioning, Hierarchical queries
- **Usage**: `--dialect oracle`

## Usage Examples

### Command Line
```bash
# MySQL with backtick identifiers
./sqlparser -sql "SELECT \`user_id\` FROM \`users\`" -dialect mysql

# PostgreSQL with double quotes
./sqlparser -sql "SELECT \"user_id\" FROM \"users\"" -dialect postgresql

# SQL Server with brackets
./sqlparser -sql "SELECT [user_id] FROM [users]" -dialect sqlserver
```

### Configuration File
```yaml
parser:
  dialect: mysql
  strict_mode: false
  max_query_size: 1000000
```

### Programmatic Usage
```go
import (
    "github.com/Chahine-tech/sql-parser-go/pkg/parser"
    "github.com/Chahine-tech/sql-parser-go/pkg/dialect"
)

// Create parser with specific dialect
d := dialect.GetDialect("mysql")
p := parser.NewWithDialect(ctx, sql, d)

// Parse with dialect-specific features
stmt, err := p.ParseStatement()
```

## Dialect-Specific Features

### Identifier Quoting
Each dialect has its own quoting style for identifiers:
- **MySQL**: `SELECT \`table\`.\`column\` FROM \`database\`.\`table\``
- **PostgreSQL**: `SELECT "table"."column" FROM "schema"."table"`
- **SQL Server**: `SELECT [table].[column] FROM [database].[schema].[table]`
- **SQLite**: `SELECT "table"."column" FROM "table"`
- **Oracle**: `SELECT "table"."column" FROM "schema"."table"`

### Feature Support
The parser recognizes dialect-specific features:
- **CTEs**: Supported in all modern dialects
- **Window Functions**: Supported in MySQL 8.0+, PostgreSQL, SQL Server, SQLite 3.25+
- **JSON Support**: MySQL 5.7+, PostgreSQL, SQL Server 2016+, SQLite 3.38+
- **Array Support**: PostgreSQL only
- **XML Support**: PostgreSQL, SQL Server, Oracle

### Keyword Recognition
Each dialect has its own set of reserved keywords and functions that are recognized during parsing.

## Default Behavior
- **Default dialect**: SQL Server
- **Configuration**: Can be overridden via CLI flag (`-dialect`) or config file
- **Fallback**: Unknown dialects default to SQL Server syntax

## Error Handling
The parser will produce appropriate error messages when:
- Dialect-specific syntax is used with wrong dialect
- Unsupported features are used for a dialect
- Invalid identifier quoting is detected
