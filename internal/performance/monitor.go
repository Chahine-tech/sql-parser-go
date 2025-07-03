package performance

import (
	"fmt"
	"runtime"
	"time"
)

// MemoryStats holds memory usage statistics
type MemoryStats struct {
	AllocBytes      uint64        `json:"alloc_bytes"`
	TotalAllocBytes uint64        `json:"total_alloc_bytes"`
	SysBytes        uint64        `json:"sys_bytes"`
	NumGC           uint32        `json:"num_gc"`
	GCPauseTotal    time.Duration `json:"gc_pause_total_ns"`
}

// GetMemoryStats returns current memory statistics
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		SysBytes:        m.Sys,
		NumGC:           m.NumGC,
		GCPauseTotal:    time.Duration(m.PauseTotalNs),
	}
}

// FormatBytes formats bytes to human readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PerformanceMonitor tracks performance metrics
type PerformanceMonitor struct {
	startTime   time.Time
	startMemory MemoryStats
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime:   time.Now(),
		startMemory: GetMemoryStats(),
	}
}

// GetMetrics returns performance metrics since monitor creation
func (pm *PerformanceMonitor) GetMetrics() map[string]interface{} {
	currentTime := time.Now()
	currentMemory := GetMemoryStats()

	return map[string]interface{}{
		"duration_ms":        currentTime.Sub(pm.startTime).Milliseconds(),
		"memory_used_bytes":  currentMemory.AllocBytes,
		"memory_used_human":  FormatBytes(currentMemory.AllocBytes),
		"memory_delta_bytes": int64(currentMemory.AllocBytes) - int64(pm.startMemory.AllocBytes),
		"gc_count_delta":     currentMemory.NumGC - pm.startMemory.NumGC,
		"gc_pause_total_ms":  currentMemory.GCPauseTotal.Milliseconds(),
	}
}

// ForceGC forces garbage collection and returns memory stats
func ForceGC() MemoryStats {
	runtime.GC()
	runtime.GC() // Double GC to ensure cleanup
	return GetMemoryStats()
}
