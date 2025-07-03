package analyzer

type QueryAnalysis struct {
	Tables     []TableInfo     `json:"tables"`
	Columns    []ColumnInfo    `json:"columns"`
	Joins      []JoinInfo      `json:"joins"`
	Conditions []ConditionInfo `json:"conditions"`
	QueryType  string          `json:"query_type"`
	Complexity int             `json:"complexity"`
	// Performance metrics
	Performance *PerformanceMetrics `json:"performance,omitempty"`
}

type TableInfo struct {
	Schema string `json:"schema,omitempty"`
	Name   string `json:"name"`
	Alias  string `json:"alias,omitempty"`
	Usage  string `json:"usage"` // SELECT, UPDATE, DELETE, INSERT
}

type ColumnInfo struct {
	Table string `json:"table,omitempty"`
	Name  string `json:"name"`
	Usage string `json:"usage"` // SELECT, WHERE, JOIN, ORDER_BY, GROUP_BY
}

type JoinInfo struct {
	Type       string `json:"type"`
	LeftTable  string `json:"left_table"`
	RightTable string `json:"right_table"`
	Condition  string `json:"condition"`
}

type ConditionInfo struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	Table    string `json:"table,omitempty"`
}

type OptimizationSuggestion struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Line        int    `json:"line,omitempty"`
}

type PerformanceMetrics struct {
	EstimatedRows        int64    `json:"estimated_rows"`
	EstimatedCost        float64  `json:"estimated_cost"`
	JoinComplexity       int      `json:"join_complexity"`
	IndexRecommendations []string `json:"index_recommendations"`
	RiskLevel            string   `json:"risk_level"` // LOW, MEDIUM, HIGH
}
