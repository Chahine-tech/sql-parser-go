package tests

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Chahine-tech/sql-parser-go/pkg/analyzer"
	"github.com/Chahine-tech/sql-parser-go/pkg/lexer"
	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

func BenchmarkLexer(b *testing.B) {
	input := `SELECT u.name, u.email, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.status = 'active' 
			  GROUP BY u.id, u.name, u.email 
			  ORDER BY order_count DESC`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(input)
		for {
			tok := l.NextToken()
			if tok.Type == lexer.EOF {
				break
			}
		}
	}
}

func BenchmarkLexerWithBuffer(b *testing.B) {
	input := `SELECT u.name, u.email, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.status = 'active' 
			  GROUP BY u.id, u.name, u.email 
			  ORDER BY order_count DESC`

	buffer := make([]lexer.Token, 0, 50) // pre-allocate buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer.TokenizeWithBuffer(input, buffer)
	}
}

func BenchmarkParser(b *testing.B) {
	input := `SELECT u.name, u.email, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.status = 'active' 
			  GROUP BY u.id, u.name, u.email 
			  ORDER BY order_count DESC`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.New(input)
		_, err := p.ParseStatement()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAnalyzer(b *testing.B) {
	input := `SELECT u.name, u.email, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.status = 'active' 
			  GROUP BY u.id, u.name, u.email 
			  ORDER BY order_count DESC`

	p := parser.New(input)
	stmt, err := p.ParseStatement()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := analyzer.New()
		a.Analyze(stmt)
	}
}

func BenchmarkAnalyzerWithCache(b *testing.B) {
	input := `SELECT u.name, u.email, COUNT(o.id) as order_count 
			  FROM users u 
			  LEFT JOIN orders o ON u.id = o.user_id 
			  WHERE u.status = 'active' 
			  GROUP BY u.id, u.name, u.email 
			  ORDER BY order_count DESC`

	p := parser.New(input)
	stmt, err := p.ParseStatement()
	if err != nil {
		b.Fatal(err)
	}

	a := analyzer.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AnalyzeWithCache(stmt, "test_query")
	}
}

func BenchmarkComplexQuery(b *testing.B) {
	input := `SELECT 
		u.user_id,
		u.username,
		u.email,
		p.first_name,
		p.last_name,
		COUNT(o.order_id) as total_orders,
		SUM(o.total_amount) as total_spent,
		AVG(o.total_amount) as avg_order_value,
		MAX(o.order_date) as last_order_date
	FROM 
		users u
		INNER JOIN profiles p ON u.user_id = p.user_id
		LEFT JOIN orders o ON u.user_id = o.customer_id
		LEFT JOIN order_items oi ON o.order_id = oi.order_id
		LEFT JOIN products pr ON oi.product_id = pr.product_id
	WHERE 
		u.created_date >= '2023-01-01'
		AND u.status = 'active'
		AND p.country IN ('US', 'CA', 'UK')
		AND pr.category = 'electronics'
	GROUP BY 
		u.user_id, u.username, u.email, p.first_name, p.last_name
	HAVING 
		COUNT(o.order_id) > 5
		AND SUM(o.total_amount) > 1000
	ORDER BY 
		total_spent DESC, last_order_date DESC`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.New(input)
		stmt, err := p.ParseStatement()
		if err != nil {
			b.Fatal(err)
		}

		a := analyzer.New()
		a.Analyze(stmt)
	}
}

func BenchmarkConcurrentAnalyzer(b *testing.B) {
	// Create test jobs
	jobs := make([]analyzer.AnalysisJob, 100)
	for i := 0; i < 100; i++ {
		query := fmt.Sprintf("SELECT col1, col2 FROM table%d JOIN table%d_ref ON table%d.id = table%d_ref.ref_id", i, i, i, i)
		p := parser.New(query)
		stmt, _ := p.ParseStatement()

		jobs[i] = analyzer.AnalysisJob{
			ID:    fmt.Sprintf("query_%d", i),
			Query: query,
			Stmt:  stmt,
		}
	}

	ca := analyzer.NewConcurrentAnalyzer(4)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := ca.AnalyzeConcurrently(ctx, jobs)
		if len(results) != len(jobs) {
			b.Errorf("Expected %d results, got %d", len(jobs), len(results))
		}
	}
}

func BenchmarkMemoryUsage(b *testing.B) {
	query := `SELECT u.id, u.name, u.email, p.title, p.content, c.name as category
			  FROM users u 
			  JOIN posts p ON u.id = p.user_id 
			  JOIN categories c ON p.category_id = c.id 
			  WHERE u.active = 1 AND p.published = 1 
			  ORDER BY p.created_at DESC`

	var memBefore, memAfter runtime.MemStats

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.ReadMemStats(&memBefore)

		p := parser.New(query)
		stmt, _ := p.ParseStatement()
		a := analyzer.New()
		_ = a.Analyze(stmt)

		runtime.ReadMemStats(&memAfter)

		// Force cleanup to measure actual memory usage
		runtime.GC()
	}
}

func TestParserTimeout(t *testing.T) {
	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(1 * time.Millisecond)

	input := "SELECT * FROM users WHERE id = 1"
	p := parser.NewWithContext(ctx, input)

	_, err := p.ParseStatement()
	if err == nil {
		t.Error("Expected timeout error, but got none")
	}

	errors := p.Errors()
	if len(errors) == 0 {
		t.Error("Expected parsing errors due to timeout")
	}
}

func TestLargeQuery(t *testing.T) {
	// Generate a large query
	var builder strings.Builder
	builder.WriteString("SELECT ")

	// Add many columns
	for i := 0; i < 100; i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("col")
		builder.WriteString(strings.Repeat("_", i%10))
	}

	builder.WriteString(" FROM ")

	// Add many tables with joins
	for i := 0; i < 20; i++ {
		if i == 0 {
			builder.WriteString("table0")
		} else {
			builder.WriteString(" JOIN table")
			builder.WriteString(strings.Repeat("0", i%3+1))
			builder.WriteString(" ON table0.id = table")
			builder.WriteString(strings.Repeat("0", i%3+1))
			builder.WriteString(".ref_id")
		}
	}

	input := builder.String()

	start := time.Now()
	p := parser.New(input)
	stmt, err := p.ParseStatement()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to parse large query: %v", err)
	}

	if duration > 5*time.Second {
		t.Errorf("Parsing took too long: %v", duration)
	}

	// Analyze the parsed statement
	start = time.Now()
	a := analyzer.New()
	analysis := a.Analyze(stmt)
	duration = time.Since(start)

	if duration > 1*time.Second {
		t.Errorf("Analysis took too long: %v", duration)
	}

	if len(analysis.Tables) != 20 {
		t.Errorf("Expected 20 tables, got %d", len(analysis.Tables))
	}
}
