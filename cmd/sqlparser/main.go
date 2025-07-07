package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Chahine-tech/sql-parser-go/internal/config"
	"github.com/Chahine-tech/sql-parser-go/internal/performance"
	"github.com/Chahine-tech/sql-parser-go/pkg/analyzer"
	"github.com/Chahine-tech/sql-parser-go/pkg/dialect"
	"github.com/Chahine-tech/sql-parser-go/pkg/logger"
	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

func main() {
	var (
		queryFile    = flag.String("query", "", "File containing the SQL query")
		queryText    = flag.String("sql", "", "SQL query string")
		logFile      = flag.String("log", "", "SQL Server log file")
		outputFormat = flag.String("output", "json", "Output format (json, table)")
		verbose      = flag.Bool("verbose", false, "Verbose mode")
		configFile   = flag.String("config", "", "Configuration file path")
		dialectFlag  = flag.String("dialect", "", "SQL dialect (mysql, postgresql, sqlserver, sqlite, oracle)")
		showHelp     = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *showHelp {
		showUsage()
		return
	}

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	if *outputFormat != "json" {
		cfg.Output.Format = *outputFormat
	}

	// Override dialect from command line if provided
	if *dialectFlag != "" {
		cfg.Parser.Dialect = *dialectFlag
	}

	if *queryFile != "" {
		if err := analyzeQueryFile(*queryFile, cfg, *verbose); err != nil {
			fmt.Printf("Error analyzing query file: %v\n", err)
			os.Exit(1)
		}
	} else if *queryText != "" {
		if err := analyzeQueryString(*queryText, cfg, *verbose); err != nil {
			fmt.Printf("Error analyzing query: %v\n", err)
			os.Exit(1)
		}
	} else if *logFile != "" {
		if err := parseLogFile(*logFile, cfg, *verbose); err != nil {
			fmt.Printf("Error parsing log file: %v\n", err)
			os.Exit(1)
		}
	} else {
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("SQL Parser Go - Multi-Dialect SQL Query Analysis Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  sqlparser -query file.sql          Analyze SQL query from file")
	fmt.Println("  sqlparser -sql \"SELECT * FROM...\"   Analyze SQL query from string")
	fmt.Println("  sqlparser -log logfile.log          Parse SQL Server log file")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -output FORMAT    Output format: json, table (default: json)")
	fmt.Println("  -dialect DIALECT  SQL dialect: mysql, postgresql, sqlserver, sqlite, oracle (default: sqlserver)")
	fmt.Println("  -verbose          Enable verbose output")
	fmt.Println("  -config FILE      Configuration file path")
	fmt.Println("  -help             Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  sqlparser -query complex_query.sql -output json -dialect mysql")
	fmt.Println("  sqlparser -sql \"SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id\" -dialect postgresql")
	fmt.Println("  sqlparser -log sqlserver.log -output table -verbose")
}

func analyzeQueryFile(filename string, cfg *config.Config, verbose bool) error {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	return analyzeQueryString(string(content), cfg, verbose)
}

func analyzeQueryString(sql string, cfg *config.Config, verbose bool) error {
	monitor := performance.NewPerformanceMonitor()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if verbose {
		fmt.Printf("Analyzing SQL query...\n")
		fmt.Printf("Query: %s\n", sql)
		fmt.Printf("Dialect: %s\n\n", cfg.Parser.Dialect)
	}

	// Get the dialect
	d := dialect.GetDialect(cfg.Parser.Dialect)

	// Create parser with dialect
	p := parser.NewWithDialect(ctx, sql, d)
	stmt, err := p.ParseStatement()
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	if verbose {
		fmt.Printf("Parsed statement type: %s\n", stmt.Type())

		if metrics := p.GetParseMetrics(); metrics != nil {
			fmt.Printf("Parser metrics: %v\n", metrics)
		}
	}

	a := analyzer.New()
	analysis := a.Analyze(stmt)

	var suggestions []analyzer.OptimizationSuggestion
	if cfg.Analyzer.EnableOptimizations {
		if selectStmt, ok := stmt.(*parser.SelectStatement); ok {
			suggestions = a.SuggestOptimizations(selectStmt)
		}
	}

	if verbose {
		metrics := monitor.GetMetrics()
		fmt.Printf("Performance metrics: %v\n", metrics)
	}

	return outputAnalysis(analysis, suggestions, cfg)
}

func parseLogFile(filename string, cfg *config.Config, verbose bool) error {
	if verbose {
		fmt.Printf("Parsing log file: %s\n", filename)
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	// Parse the log
	logParser := logger.NewSQLServerLogParser()
	entries, err := logParser.ParseLog(file)
	if err != nil {
		return fmt.Errorf("failed to parse log: %v", err)
	}

	if verbose {
		fmt.Printf("Parsed %d log entries\n", len(entries))
	}

	// Apply filters
	filteredEntries := logger.FilterEntries(entries, logger.FilterCriteria{
		MinDuration: 0, // Will use config values in real implementation
		// Add more filter criteria based on config
	})

	if verbose {
		fmt.Printf("After filtering: %d entries\n", len(filteredEntries))
	}

	// Calculate metrics
	metrics := logger.CalculateMetrics(filteredEntries)

	// Output results
	return outputLogAnalysis(filteredEntries, metrics, cfg)
}

func outputAnalysis(analysis analyzer.QueryAnalysis, suggestions []analyzer.OptimizationSuggestion, cfg *config.Config) error {
	switch cfg.Output.Format {
	case "json":
		return outputJSON(map[string]interface{}{
			"analysis":    analysis,
			"suggestions": suggestions,
		}, cfg.Output.PrettyJSON)
	case "table":
		return outputTable(analysis, suggestions)
	case "csv":
		return outputCSV(analysis, suggestions)
	default:
		return fmt.Errorf("unsupported output format: %s", cfg.Output.Format)
	}
}

func outputLogAnalysis(entries []logger.LogEntry, metrics logger.LogMetrics, cfg *config.Config) error {
	switch cfg.Output.Format {
	case "json":
		return outputJSON(map[string]interface{}{
			"entries": entries,
			"metrics": metrics,
		}, cfg.Output.PrettyJSON)
	case "table":
		return outputLogTable(entries, metrics)
	default:
		return fmt.Errorf("unsupported output format: %s", cfg.Output.Format)
	}
}

func outputJSON(data interface{}, pretty bool) error {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	fmt.Println(string(output))
	return nil
}

func outputTable(analysis analyzer.QueryAnalysis, suggestions []analyzer.OptimizationSuggestion) error {
	fmt.Println("=== SQL Query Analysis ===")
	fmt.Printf("Query Type: %s\n", analysis.QueryType)
	fmt.Printf("Complexity: %d\n\n", analysis.Complexity)

	// Tables
	if len(analysis.Tables) > 0 {
		fmt.Println("Tables:")
		fmt.Printf("%-20s %-10s %-10s %s\n", "Name", "Schema", "Alias", "Usage")
		fmt.Println(strings.Repeat("-", 60))
		for _, table := range analysis.Tables {
			fmt.Printf("%-20s %-10s %-10s %s\n", table.Name, table.Schema, table.Alias, table.Usage)
		}
		fmt.Println()
	}

	// Columns
	if len(analysis.Columns) > 0 {
		fmt.Println("Columns:")
		fmt.Printf("%-20s %-10s %s\n", "Name", "Table", "Usage")
		fmt.Println(strings.Repeat("-", 40))
		for _, column := range analysis.Columns {
			fmt.Printf("%-20s %-10s %s\n", column.Name, column.Table, column.Usage)
		}
		fmt.Println()
	}

	// Joins
	if len(analysis.Joins) > 0 {
		fmt.Println("Joins:")
		fmt.Printf("%-10s %-15s %-15s %s\n", "Type", "Left Table", "Right Table", "Condition")
		fmt.Println(strings.Repeat("-", 60))
		for _, join := range analysis.Joins {
			fmt.Printf("%-10s %-15s %-15s %s\n", join.Type, join.LeftTable, join.RightTable, join.Condition)
		}
		fmt.Println()
	}

	// Optimization Suggestions
	if len(suggestions) > 0 {
		fmt.Println("Optimization Suggestions:")
		for i, suggestion := range suggestions {
			fmt.Printf("%d. [%s] %s\n", i+1, suggestion.Severity, suggestion.Description)
		}
		fmt.Println()
	}

	return nil
}

func outputLogTable(entries []logger.LogEntry, metrics logger.LogMetrics) error {
	fmt.Println("=== SQL Server Log Analysis ===")
	fmt.Printf("Total Entries: %d\n", metrics.TotalEntries)
	fmt.Printf("Average Duration: %.2f ms\n", metrics.AvgDuration)
	fmt.Printf("Max Duration: %d ms\n", metrics.MaxDuration)
	fmt.Printf("Min Duration: %d ms\n\n", metrics.MinDuration)

	// Query Types Distribution
	if len(metrics.QueryTypes) > 0 {
		fmt.Println("Query Types:")
		for queryType, count := range metrics.QueryTypes {
			fmt.Printf("  %s: %d\n", queryType, count)
		}
		fmt.Println()
	}

	// Show first few entries
	if len(entries) > 0 {
		fmt.Println("Recent Entries:")
		fmt.Printf("%-20s %-8s %-8s %-15s %s\n", "Timestamp", "Duration", "Reads", "Database", "Query")
		fmt.Println(strings.Repeat("-", 80))

		limit := 10
		if len(entries) < limit {
			limit = len(entries)
		}

		for i := 0; i < limit; i++ {
			entry := entries[i]
			query := entry.Query
			if len(query) > 40 {
				query = query[:37] + "..."
			}
			fmt.Printf("%-20s %-8d %-8d %-15s %s\n",
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				entry.Duration,
				entry.Reads,
				entry.Database,
				query)
		}
	}

	return nil
}

func outputCSV(analysis analyzer.QueryAnalysis, suggestions []analyzer.OptimizationSuggestion) error {
	fmt.Println("Query Type,Complexity,Table Count,Column Count,Join Count,Suggestions")
	fmt.Printf("%s,%d,%d,%d,%d,%d\n",
		analysis.QueryType,
		analysis.Complexity,
		len(analysis.Tables),
		len(analysis.Columns),
		len(analysis.Joins),
		len(suggestions))

	// Tables CSV
	if len(analysis.Tables) > 0 {
		fmt.Println("\nTables:")
		fmt.Println("Name,Schema,Alias,Usage")
		for _, table := range analysis.Tables {
			fmt.Printf("%s,%s,%s,%s\n", table.Name, table.Schema, table.Alias, table.Usage)
		}
	}

	// Columns CSV
	if len(analysis.Columns) > 0 {
		fmt.Println("\nColumns:")
		fmt.Println("Name,Table,Usage")
		for _, column := range analysis.Columns {
			fmt.Printf("%s,%s,%s\n", column.Name, column.Table, column.Usage)
		}
	}

	// Suggestions CSV
	if len(suggestions) > 0 {
		fmt.Println("\nSuggestions:")
		fmt.Println("Type,Description,Severity")
		for _, suggestion := range suggestions {
			fmt.Printf("%s,\"%s\",%s\n", suggestion.Type, suggestion.Description, suggestion.Severity)
		}
	}

	return nil
}
