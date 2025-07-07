package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/Chahine-tech/sql-parser-go/pkg/dialect"
	"github.com/Chahine-tech/sql-parser-go/pkg/lexer"
	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

// Dialect-specific SQL queries for benchmarking
var dialectQueries = map[string]string{
	"mysql": "SELECT " +
		"`users`.`user_id`, " +
		"`users`.`email`, " +
		"`users`.`created_at`, " +
		"COUNT(`orders`.`order_id`) AS `order_count`, " +
		"SUM(`orders`.`total_amount`) AS `total_spent` " +
		"FROM `ecommerce`.`users` " +
		"LEFT JOIN `ecommerce`.`orders` ON `users`.`user_id` = `orders`.`user_id` " +
		"WHERE `users`.`status` = 'active' " +
		"AND `users`.`created_at` >= '2024-01-01' " +
		"GROUP BY `users`.`user_id`, `users`.`email`, `users`.`created_at` " +
		"HAVING `order_count` > 0 " +
		"ORDER BY `total_spent` DESC, `order_count` DESC " +
		"LIMIT 100",

	"postgresql": `SELECT 
		"users"."user_id",
		"users"."email",
		"users"."created_at",
		COUNT("orders"."order_id") AS "order_count",
		SUM("orders"."total_amount") AS "total_spent"
	FROM "public"."users"
	LEFT JOIN "public"."orders" ON "users"."user_id" = "orders"."user_id"
	WHERE "users"."status" = 'active' 
		AND "users"."created_at" >= '2024-01-01'
	GROUP BY "users"."user_id", "users"."email", "users"."created_at"
	HAVING "order_count" > 0
	ORDER BY "total_spent" DESC, "order_count" DESC
	LIMIT 100`,

	"sqlserver": `SELECT 
		[users].[user_id],
		[users].[email],
		[users].[created_at],
		COUNT([orders].[order_id]) AS [order_count],
		SUM([orders].[total_amount]) AS [total_spent]
	FROM [ecommerce].[dbo].[users]
	LEFT JOIN [ecommerce].[dbo].[orders] ON [users].[user_id] = [orders].[user_id]
	WHERE [users].[status] = 'active' 
		AND [users].[created_at] >= '2024-01-01'
	GROUP BY [users].[user_id], [users].[email], [users].[created_at]
	HAVING [order_count] > 0
	ORDER BY [total_spent] DESC, [order_count] DESC`,

	"sqlite": `SELECT 
		"users"."user_id",
		"users"."email",
		"users"."created_at",
		COUNT("orders"."order_id") AS "order_count",
		SUM("orders"."total_amount") AS "total_spent"
	FROM "users"
	LEFT JOIN "orders" ON "users"."user_id" = "orders"."user_id"
	WHERE "users"."status" = 'active' 
		AND "users"."created_at" >= '2024-01-01'
	GROUP BY "users"."user_id", "users"."email", "users"."created_at"
	HAVING "order_count" > 0
	ORDER BY "total_spent" DESC, "order_count" DESC
	LIMIT 100`,

	"oracle": `SELECT 
		"users"."user_id",
		"users"."email",
		"users"."created_at",
		COUNT("orders"."order_id") AS "order_count",
		SUM("orders"."total_amount") AS "total_spent"
	FROM "schema"."users"
	LEFT JOIN "schema"."orders" ON "users"."user_id" = "orders"."user_id"
	WHERE "users"."status" = 'active' 
		AND "users"."created_at" >= DATE '2024-01-01'
	GROUP BY "users"."user_id", "users"."email", "users"."created_at"
	HAVING COUNT("orders"."order_id") > 0
	ORDER BY "total_spent" DESC, "order_count" DESC`,
}

// BenchmarkDialectLexing benchmarks lexer performance for each SQL dialect
func BenchmarkDialectLexing(b *testing.B) {
	for dialectName, query := range dialectQueries {
		b.Run(dialectName, func(b *testing.B) {
			d := dialect.GetDialect(dialectName)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				l := lexer.NewWithDialect(query, d)
				tokenCount := 0
				for {
					tok := l.NextToken()
					tokenCount++
					if tok.Type == lexer.EOF {
						break
					}
				}
				b.SetBytes(int64(len(query)))
			}
		})
	}
}

// BenchmarkDialectParsing benchmarks parser performance for each SQL dialect
func BenchmarkDialectParsing(b *testing.B) {
	for dialectName, query := range dialectQueries {
		b.Run(dialectName, func(b *testing.B) {
			d := dialect.GetDialect(dialectName)
			ctx := context.Background()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				p := parser.NewWithDialect(ctx, query, d)
				_, err := p.ParseStatement()
				if err != nil {
					b.Fatalf("Parse failed for %s: %v", dialectName, err)
				}
				b.SetBytes(int64(len(query)))
			}
		})
	}
}

// BenchmarkDialectKeywordLookup benchmarks keyword lookup performance
func BenchmarkDialectKeywordLookup(b *testing.B) {
	keywords := []string{"SELECT", "FROM", "WHERE", "JOIN", "GROUP", "ORDER", "LIMIT", "SHOW", "DESCRIBE", "PRAGMA"}
	
	for dialectName := range dialectQueries {
		b.Run(dialectName, func(b *testing.B) {
			d := dialect.GetDialect(dialectName)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				for _, keyword := range keywords {
					d.IsReservedWord(keyword)
				}
			}
		})
	}
}

// BenchmarkDialectQuoteIdentifier benchmarks identifier quoting performance
func BenchmarkDialectQuoteIdentifier(b *testing.B) {
	identifiers := []string{"table", "column", "user_id", "database", "schema", "long_table_name_with_underscores"}
	
	for dialectName := range dialectQueries {
		b.Run(dialectName, func(b *testing.B) {
			d := dialect.GetDialect(dialectName)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				for _, identifier := range identifiers {
					d.QuoteIdentifier(identifier)
				}
			}
		})
	}
}

// BenchmarkDialectFeatureSupport benchmarks feature support checking
func BenchmarkDialectFeatureSupport(b *testing.B) {
	features := []dialect.Feature{
		dialect.FeatureCTE,
		dialect.FeatureWindowFunctions,
		dialect.FeatureJSONSupport,
		dialect.FeatureArraySupport,
		dialect.FeatureRecursiveCTE,
		dialect.FeaturePartitioning,
		dialect.FeatureFullTextSearch,
		dialect.FeatureXMLSupport,
		dialect.FeatureUpsert,
		dialect.FeatureReturningClause,
	}
	
	for dialectName := range dialectQueries {
		b.Run(dialectName, func(b *testing.B) {
			d := dialect.GetDialect(dialectName)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				for _, feature := range features {
					d.SupportsFeature(feature)
				}
			}
		})
	}
}

// BenchmarkComplexDialectQueries benchmarks parsing of complex, real-world queries
func BenchmarkComplexDialectQueries(b *testing.B) {
	complexQueries := map[string]string{
		"mysql_complex": "SELECT " +
			"`u`.`user_id`, " +
			"`u`.`username`, " +
			"`p`.`name` AS `product_name`, " +
			"`oi`.`quantity`, " +
			"`oi`.`price`, " +
			"(`oi`.`quantity` * `oi`.`price`) AS `line_total`, " +
			"`o`.`order_date`, " +
			"`o`.`status`, " +
			"CASE " +
			"WHEN `o`.`total_amount` > 1000 THEN 'Premium' " +
			"WHEN `o`.`total_amount` > 500 THEN 'Standard' " +
			"ELSE 'Basic' " +
			"END AS `customer_tier`, " +
			"ROW_NUMBER() OVER (PARTITION BY `u`.`user_id` ORDER BY `o`.`order_date` DESC) AS `order_rank` " +
			"FROM `ecommerce`.`users` `u` " +
			"INNER JOIN `ecommerce`.`orders` `o` ON `u`.`user_id` = `o`.`user_id` " +
			"INNER JOIN `ecommerce`.`order_items` `oi` ON `o`.`order_id` = `oi`.`order_id` " +
			"INNER JOIN `ecommerce`.`products` `p` ON `oi`.`product_id` = `p`.`product_id` " +
			"WHERE `o`.`order_date` >= DATE_SUB(NOW(), INTERVAL 30 DAY) " +
			"AND `u`.`status` = 'active' " +
			"AND `p`.`category` IN ('electronics', 'books', 'clothing') " +
			"GROUP BY `u`.`user_id`, `u`.`username`, `p`.`name`, `oi`.`quantity`, `oi`.`price`, `o`.`order_date`, `o`.`status`, `o`.`total_amount` " +
			"HAVING `line_total` > 50 " +
			"ORDER BY `o`.`order_date` DESC, `line_total` DESC " +
			"LIMIT 1000",

		"postgresql_complex": "SELECT " +
			"\"u\".\"user_id\", " +
			"\"u\".\"username\", " +
			"\"p\".\"name\" AS \"product_name\", " +
			"\"oi\".\"quantity\", " +
			"\"oi\".\"price\", " +
			"(\"oi\".\"quantity\" * \"oi\".\"price\") AS \"line_total\", " +
			"\"o\".\"order_date\", " +
			"\"o\".\"status\", " +
			"CASE " +
			"WHEN \"o\".\"total_amount\" > 1000 THEN 'Premium' " +
			"WHEN \"o\".\"total_amount\" > 500 THEN 'Standard' " +
			"ELSE 'Basic' " +
			"END AS \"customer_tier\" " +
			"FROM \"public\".\"users\" \"u\" " +
			"INNER JOIN \"public\".\"orders\" \"o\" ON \"u\".\"user_id\" = \"o\".\"user_id\" " +
			"INNER JOIN \"public\".\"order_items\" \"oi\" ON \"o\".\"order_id\" = \"oi\".\"order_id\" " +
			"INNER JOIN \"public\".\"products\" \"p\" ON \"oi\".\"product_id\" = \"p\".\"product_id\" " +
			"WHERE \"o\".\"order_date\" >= CURRENT_DATE - INTERVAL '90 days' " +
			"AND \"u\".\"status\" = 'active' " +
			"AND \"p\".\"category\" IN ('electronics', 'books', 'clothing') " +
			"GROUP BY \"u\".\"user_id\", \"u\".\"username\", \"p\".\"name\", \"oi\".\"quantity\", \"oi\".\"price\", \"o\".\"order_date\", \"o\".\"status\", \"o\".\"total_amount\" " +
			"HAVING \"line_total\" > 50 " +
			"ORDER BY \"o\".\"order_date\" DESC, \"line_total\" DESC " +
			"LIMIT 1000",
	}

	for queryName, query := range complexQueries {
		dialectName := queryName[:strings.Index(queryName, "_")]
		b.Run(queryName, func(b *testing.B) {
			d := dialect.GetDialect(dialectName)
			ctx := context.Background()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				p := parser.NewWithDialect(ctx, query, d)
				_, err := p.ParseStatement()
				if err != nil {
					b.Fatalf("Parse failed for %s: %v", queryName, err)
				}
				b.SetBytes(int64(len(query)))
			}
		})
	}
}

// BenchmarkDialectMemoryUsage measures memory allocation for different dialects
func BenchmarkDialectMemoryUsage(b *testing.B) {
	for dialectName, query := range dialectQueries {
		b.Run(dialectName, func(b *testing.B) {
			d := dialect.GetDialect(dialectName)
			ctx := context.Background()
			b.ReportAllocs()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				p := parser.NewWithDialect(ctx, query, d)
				_, err := p.ParseStatement()
				if err != nil {
					b.Fatalf("Parse failed for %s: %v", dialectName, err)
				}
			}
		})
	}
}
