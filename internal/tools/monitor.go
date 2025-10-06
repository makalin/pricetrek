package tools

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// SystemMonitor monitors system resources and performance
type SystemMonitor struct {
	startTime time.Time
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		startTime: time.Now(),
	}
}

// GetSystemStats returns current system statistics
func (sm *SystemMonitor) GetSystemStats() SystemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemStats{
		Uptime:        time.Since(sm.startTime),
		GoRoutines:    runtime.NumGoroutine(),
		MemoryAlloc:   m.Alloc,
		MemoryTotal:   m.TotalAlloc,
		MemorySys:     m.Sys,
		NumGC:         m.NumGC,
		GCPauseTotal:  m.PauseTotalNs,
		LastGC:        time.Unix(0, int64(m.LastGC)),
	}
}

// SystemStats contains system performance statistics
type SystemStats struct {
	Uptime        time.Duration
	GoRoutines    int
	MemoryAlloc   uint64
	MemoryTotal   uint64
	MemorySys     uint64
	NumGC         uint32
	GCPauseTotal  uint64
	LastGC        time.Time
}

// FormatBytes formats bytes into human readable format
func (ss *SystemStats) FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// MonitorLoop runs continuous monitoring
func (sm *SystemMonitor) MonitorLoop(ctx context.Context, interval time.Duration, callback func(SystemStats)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := sm.GetSystemStats()
			callback(stats)
		}
	}
}

// HealthChecker performs health checks
type HealthChecker struct {
	checks []HealthCheck
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string
	Description string
	Check       func() error
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make([]HealthCheck, 0),
	}
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(name, description string, check func() error) {
	hc.checks = append(hc.checks, HealthCheck{
		Name:        name,
		Description: description,
		Check:       check,
	})
}

// RunChecks runs all health checks
func (hc *HealthChecker) RunChecks() []HealthCheckResult {
	results := make([]HealthCheckResult, len(hc.checks))

	for i, check := range hc.checks {
		start := time.Now()
		err := check.Check()
		duration := time.Since(start)

		results[i] = HealthCheckResult{
			Name:        check.Name,
			Description: check.Description,
			Status:      err == nil,
			Error:       err,
			Duration:    duration,
		}
	}

	return results
}

// HealthCheckResult contains the result of a health check
type HealthCheckResult struct {
	Name        string
	Description string
	Status      bool
	Error       error
	Duration    time.Duration
}

// GetOverallStatus returns the overall health status
func (hc *HealthChecker) GetOverallStatus(results []HealthCheckResult) bool {
	for _, result := range results {
		if !result.Status {
			return false
		}
	}
	return true
}