package analyzer

import (
	"context"
	"runtime"
	"sync"

	"github.com/Chahine-tech/sql-parser-go/pkg/parser"
)

// ConcurrentAnalyzer performs analysis of multiple queries concurrently
type ConcurrentAnalyzer struct {
	workerCount int
	cache       map[string]QueryAnalysis
	mu          sync.RWMutex
}

// NewConcurrentAnalyzer creates a new concurrent analyzer
func NewConcurrentAnalyzer(workerCount int) *ConcurrentAnalyzer {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	return &ConcurrentAnalyzer{
		workerCount: workerCount,
		cache:       make(map[string]QueryAnalysis),
	}
}

// AnalysisJob represents a single analysis job
type AnalysisJob struct {
	ID    string
	Query string
	Stmt  parser.Statement
}

// AnalysisResult represents the result of an analysis job
type AnalysisResult struct {
	ID       string
	Analysis QueryAnalysis
	Error    error
}

// AnalyzeConcurrently analyzes multiple queries concurrently
func (ca *ConcurrentAnalyzer) AnalyzeConcurrently(ctx context.Context, jobs []AnalysisJob) []AnalysisResult {
	jobChan := make(chan AnalysisJob, len(jobs))
	resultChan := make(chan AnalysisResult, len(jobs))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < ca.workerCount; i++ {
		wg.Add(1)
		go ca.worker(ctx, &wg, jobChan, resultChan)
	}

	// Send jobs
	go func() {
		defer close(jobChan)
		for _, job := range jobs {
			select {
			case jobChan <- job:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	results := make([]AnalysisResult, 0, len(jobs))
	for i := 0; i < len(jobs); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case <-ctx.Done():
			results = append(results, AnalysisResult{
				Error: ctx.Err(),
			})
		}
	}

	// Wait for workers to finish
	wg.Wait()
	close(resultChan)

	return results
}

// worker processes analysis jobs
func (ca *ConcurrentAnalyzer) worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan AnalysisJob, resultChan chan<- AnalysisResult) {
	defer wg.Done()

	analyzer := New()

	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				return
			}

			// Check cache first
			ca.mu.RLock()
			if cached, exists := ca.cache[job.ID]; exists {
				ca.mu.RUnlock()
				resultChan <- AnalysisResult{
					ID:       job.ID,
					Analysis: cached,
				}
				continue
			}
			ca.mu.RUnlock()

			// Perform analysis
			analysis := analyzer.Analyze(job.Stmt)

			// Cache result
			ca.mu.Lock()
			ca.cache[job.ID] = analysis
			ca.mu.Unlock()

			resultChan <- AnalysisResult{
				ID:       job.ID,
				Analysis: analysis,
			}

		case <-ctx.Done():
			return
		}
	}
}

// GetCacheStats returns cache statistics
func (ca *ConcurrentAnalyzer) GetCacheStats() map[string]interface{} {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	return map[string]interface{}{
		"cache_size":   len(ca.cache),
		"worker_count": ca.workerCount,
	}
}

// ClearCache clears the analysis cache
func (ca *ConcurrentAnalyzer) ClearCache() {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	ca.cache = make(map[string]QueryAnalysis)
}
