package logger

import "strings"

// Supported log formats for SQL Server
type LogFormat int

const (
	FormatUnknown LogFormat = iota
	FormatProfiler
	FormatExtendedEvents
	FormatQueryStore
	FormatErrorLog
	FormatPerformanceCounter
)

func (lf LogFormat) String() string {
	switch lf {
	case FormatProfiler:
		return "SQL Server Profiler"
	case FormatExtendedEvents:
		return "Extended Events"
	case FormatQueryStore:
		return "Query Store"
	case FormatErrorLog:
		return "Error Log"
	case FormatPerformanceCounter:
		return "Performance Counter"
	default:
		return "Unknown"
	}
}

// LogFormatDetector helps identify the format of log files
type LogFormatDetector struct{}

func NewLogFormatDetector() *LogFormatDetector {
	return &LogFormatDetector{}
}

func (d *LogFormatDetector) DetectFormat(sample string) LogFormat {
	// Check for Extended Events XML format
	if containsAny(sample, []string{"<event name=", "xml version=", "<Events>"}) {
		return FormatExtendedEvents
	}

	// Check for Query Store JSON format
	if containsAny(sample, []string{"query_sql_text", "avg_duration", "execution_count"}) {
		return FormatQueryStore
	}

	// Check for Profiler format (timestamp followed by event)
	if containsAny(sample, []string{"SQL:BatchCompleted", "RPC:Completed", "LoginEvent"}) {
		return FormatProfiler
	}

	// Check for Error Log format
	if containsAny(sample, []string{"spid", "Login succeeded", "Error:", "Warning:"}) {
		return FormatErrorLog
	}

	// Check for Performance Counter format
	if containsAny(sample, []string{"Duration:", "CPU:", "Reads:", "Writes:"}) {
		return FormatPerformanceCounter
	}

	return FormatUnknown
}

func containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if len(text) >= len(pattern) {
			for i := 0; i <= len(text)-len(pattern); i++ {
				if text[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}
	return false
}

// LogMetrics provides statistics about parsed logs
type LogMetrics struct {
	TotalEntries       int            `json:"total_entries"`
	QueryTypes         map[string]int `json:"query_types"`
	AvgDuration        float64        `json:"avg_duration_ms"`
	MaxDuration        int64          `json:"max_duration_ms"`
	MinDuration        int64          `json:"min_duration_ms"`
	TotalReads         int64          `json:"total_reads"`
	TotalWrites        int64          `json:"total_writes"`
	DatabaseCounts     map[string]int `json:"database_counts"`
	HourlyDistribution map[int]int    `json:"hourly_distribution"`
}

func CalculateMetrics(entries []LogEntry) LogMetrics {
	metrics := LogMetrics{
		QueryTypes:         make(map[string]int),
		DatabaseCounts:     make(map[string]int),
		HourlyDistribution: make(map[int]int),
	}

	if len(entries) == 0 {
		return metrics
	}

	metrics.TotalEntries = len(entries)
	metrics.MinDuration = entries[0].Duration
	metrics.MaxDuration = entries[0].Duration

	var totalDuration int64

	for _, entry := range entries {
		// Query type analysis
		queryType := getQueryType(entry.Query)
		metrics.QueryTypes[queryType]++

		// Duration analysis
		totalDuration += entry.Duration
		if entry.Duration > metrics.MaxDuration {
			metrics.MaxDuration = entry.Duration
		}
		if entry.Duration < metrics.MinDuration {
			metrics.MinDuration = entry.Duration
		}

		// I/O analysis
		metrics.TotalReads += entry.Reads
		metrics.TotalWrites += entry.Writes

		// Database analysis
		if entry.Database != "" {
			metrics.DatabaseCounts[entry.Database]++
		}

		// Hourly distribution
		hour := entry.Timestamp.Hour()
		metrics.HourlyDistribution[hour]++
	}

	if metrics.TotalEntries > 0 {
		metrics.AvgDuration = float64(totalDuration) / float64(metrics.TotalEntries)
	}

	return metrics
}

func getQueryType(query string) string {
	if len(query) == 0 {
		return "UNKNOWN"
	}

	query = strings.ToUpper(strings.TrimSpace(query))

	if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(query, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(query, "DROP") {
		return "DROP"
	} else if strings.HasPrefix(query, "ALTER") {
		return "ALTER"
	} else if strings.HasPrefix(query, "EXEC") || strings.HasPrefix(query, "EXECUTE") {
		return "EXECUTE"
	}

	return "OTHER"
}
