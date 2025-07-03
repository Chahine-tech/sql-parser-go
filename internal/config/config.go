package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Parser settings
	Parser ParserConfig `json:"parser" yaml:"parser"`

	// Analyzer settings
	Analyzer AnalyzerConfig `json:"analyzer" yaml:"analyzer"`

	// Logger settings
	Logger LoggerConfig `json:"logger" yaml:"logger"`

	// Output settings
	Output OutputConfig `json:"output" yaml:"output"`
}

type ParserConfig struct {
	// Enable strict mode for parsing
	StrictMode bool `json:"strict_mode" yaml:"strict_mode"`

	// Maximum query size to parse (in characters)
	MaxQuerySize int `json:"max_query_size" yaml:"max_query_size"`

	// SQL dialect (sqlserver, mysql, postgresql, etc.)
	Dialect string `json:"dialect" yaml:"dialect"`
}

type AnalyzerConfig struct {
	// Enable optimization suggestions
	EnableOptimizations bool `json:"enable_optimizations" yaml:"enable_optimizations"`

	// Complexity threshold for warnings
	ComplexityThreshold int `json:"complexity_threshold" yaml:"complexity_threshold"`

	// Enable detailed column analysis
	DetailedAnalysis bool `json:"detailed_analysis" yaml:"detailed_analysis"`
}

type LoggerConfig struct {
	// Default log format if auto-detection fails
	DefaultFormat string `json:"default_format" yaml:"default_format"`

	// Maximum log file size to process (in MB)
	MaxFileSizeMB int `json:"max_file_size_mb" yaml:"max_file_size_mb"`

	// Filter settings
	Filters FilterConfig `json:"filters" yaml:"filters"`
}

type FilterConfig struct {
	// Minimum duration in milliseconds
	MinDurationMs int64 `json:"min_duration_ms" yaml:"min_duration_ms"`

	// Maximum duration in milliseconds
	MaxDurationMs int64 `json:"max_duration_ms" yaml:"max_duration_ms"`

	// Include only specific query types
	QueryTypes []string `json:"query_types" yaml:"query_types"`

	// Exclude system queries
	ExcludeSystem bool `json:"exclude_system" yaml:"exclude_system"`
}

type OutputConfig struct {
	// Output format (json, table, csv)
	Format string `json:"format" yaml:"format"`

	// Pretty print JSON output
	PrettyJSON bool `json:"pretty_json" yaml:"pretty_json"`

	// Include timestamps in output
	IncludeTimestamps bool `json:"include_timestamps" yaml:"include_timestamps"`

	// Output file path (empty for stdout)
	OutputFile string `json:"output_file" yaml:"output_file"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Parser: ParserConfig{
			StrictMode:   false,
			MaxQuerySize: 1000000, // 1MB
			Dialect:      "sqlserver",
		},
		Analyzer: AnalyzerConfig{
			EnableOptimizations: true,
			ComplexityThreshold: 10,
			DetailedAnalysis:    true,
		},
		Logger: LoggerConfig{
			DefaultFormat: "profiler",
			MaxFileSizeMB: 100,
			Filters: FilterConfig{
				MinDurationMs: 0,
				MaxDurationMs: 0,          // No limit
				QueryTypes:    []string{}, // All types
				ExcludeSystem: true,
			},
		},
		Output: OutputConfig{
			Format:            "json",
			PrettyJSON:        true,
			IncludeTimestamps: true,
			OutputFile:        "", // stdout
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(filename string) (*Config, error) {
	config := DefaultConfig()

	if filename == "" {
		return config, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	ext := filepath.Ext(filename)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, config)
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, filename string) error {
	var data []byte
	var err error

	ext := filepath.Ext(filename)
	switch ext {
	case ".json":
		if config.Output.PrettyJSON {
			data, err = json.MarshalIndent(config, "", "  ")
		} else {
			data, err = json.Marshal(config)
		}
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Parser.MaxQuerySize <= 0 {
		return fmt.Errorf("parser.max_query_size must be positive")
	}

	if c.Analyzer.ComplexityThreshold < 0 {
		return fmt.Errorf("analyzer.complexity_threshold must be non-negative")
	}

	if c.Logger.MaxFileSizeMB <= 0 {
		return fmt.Errorf("logger.max_file_size_mb must be positive")
	}

	validFormats := map[string]bool{
		"json":  true,
		"table": true,
		"csv":   true,
	}

	if !validFormats[c.Output.Format] {
		return fmt.Errorf("invalid output format: %s", c.Output.Format)
	}

	validDialects := map[string]bool{
		"sqlserver":  true,
		"mysql":      true,
		"postgresql": true,
		"oracle":     true,
		"sqlite":     true,
	}

	if !validDialects[c.Parser.Dialect] {
		return fmt.Errorf("invalid SQL dialect: %s", c.Parser.Dialect)
	}

	return nil
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "sqlparser.yaml"
	}
	return filepath.Join(homeDir, ".sqlparser", "config.yaml")
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	return os.MkdirAll(configDir, 0755)
}
