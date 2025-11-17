# ASMM8 Performance Optimization Guide

**Date:** July 18, 2025
**Last Updated:** November 7, 2025
**Focus:** Performance bottlenecks and optimization strategies

## Overview

This document analyzes the current performance characteristics of ASMM8 and provides specific recommendations for optimization across different system components.

## Current Performance Analysis

### 1. Database Performance

#### Issues Identified

**Resource Leaks (HIGH IMPACT)**
- **Location:** `pkg/db8/db8_domain8.go`
- **Issue:** Missing `defer rows.Close()` statements
- **Impact:** Connection pool exhaustion, memory leaks

```go
// PROBLEMATIC CODE
func (m *Db8Domain8) GetAllDomain() ([]model8.Domain8, error) {
    query, err := m.Db.Query("SELECT id, name, companyname, enabled FROM cptm8domain")
    if err != nil {
        return []model8.Domain8{}, err
    }
    // Missing: defer query.Close()
    
    var domains []model8.Domain8
    for query.Next() {
        // Process results
    }
    return domains, nil
}
```

**Query Inefficiencies (MEDIUM IMPACT)**
- **Location:** `pkg/db8/db8_domain8.go:94-103`
- **Issue:** `ExistEnabled()` fetches unnecessary data
- **Impact:** Increased I/O and memory usage

```go
// INEFFICIENT
func (m *Db8Domain8) ExistEnabled() bool {
    err := m.Db.QueryRow("SELECT id, name, companyname, enabled FROM cptm8domain WHERE enabled = true").Scan()
    return err == nil
}

// OPTIMIZED
func (m *Db8Domain8) ExistEnabled() bool {
    var exists bool
    err := m.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM cptm8domain WHERE enabled = true)").Scan(&exists)
    return err == nil && exists
}
```

**Code Duplication (MEDIUM IMPACT)**
- **Location:** `pkg/db8/db8_domain8.go:40-92`
- **Issue:** Duplicate query logic in `GetAllDomain()` and `GetAllEnabled()`
- **Impact:** Maintenance overhead, potential inconsistencies

### 2. Scanning Performance

#### Concurrency Analysis

**Strengths:**
- Good use of goroutines for parallel processing
- Channel-based communication between components
- Separate input/output processing

**Areas for Improvement:**

**Temporary File Management (HIGH IMPACT)**
- **Location:** `pkg/active/active.go:42-54`
- **Issue:** Temp files created without cleanup
- **Impact:** Disk space exhaustion, I/O bottlenecks

```go
// PROBLEMATIC CODE
for _, domain := range r.SeedDomains {
    tempFile := "./tmp/tempDomain-" + domain + ".txt"
    utils.WriteTempFile(tempFile, results.Hostnames[domain])
    // No cleanup mechanism
}
```

**Memory Usage (MEDIUM IMPACT)**
- **Location:** Scanning modules
- **Issue:** All results stored in memory
- **Impact:** High memory usage with large datasets

```go
// CURRENT: Store all results in memory
results.Hostnames = make(map[string][]string)
for _, domain := range r.SeedDomains {
    results.Hostnames[domain] = append(results.Hostnames[domain], newResults...)
}
```

### 3. I/O Performance

#### File Operations
- **Issue:** Synchronous file operations
- **Impact:** Blocking I/O during large result processing

#### Network Operations
- **Issue:** No connection pooling for external services
- **Impact:** Connection overhead for repeated requests

## Performance Optimization Recommendations

### 1. Database Optimizations

#### Immediate Fixes

**Fix Resource Leaks:**
```go
func (m *Db8Domain8) GetAllDomain() ([]model8.Domain8, error) {
    query, err := m.Db.Query("SELECT id, name, companyname, enabled FROM cptm8domain")
    if err != nil {
        return []model8.Domain8{}, err
    }
    defer query.Close() // FIX: Add defer statement
    
    var domains []model8.Domain8
    for query.Next() {
        // Process results
    }
    return domains, nil
}
```

**Optimize Queries:**
```go
func (m *Db8Domain8) ExistEnabled() bool {
    var exists bool
    err := m.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM cptm8domain WHERE enabled = true LIMIT 1)").Scan(&exists)
    return err == nil && exists
}
```

**Consolidate Duplicate Code:**
```go
func (m *Db8Domain8) GetDomains(enabledOnly bool) ([]model8.Domain8, error) {
    query := "SELECT id, name, companyname, enabled FROM cptm8domain"
    var args []interface{}
    
    if enabledOnly {
        query += " WHERE enabled = $1"
        args = append(args, true)
    }
    
    rows, err := m.Db.Query(query, args...)
    if err != nil {
        return []model8.Domain8{}, err
    }
    defer rows.Close()
    
    var domains []model8.Domain8
    for rows.Next() {
        var domain model8.Domain8
        err := rows.Scan(&domain.Id, &domain.Name, &domain.Companyname, &domain.Enabled)
        if err != nil {
            return nil, err
        }
        domains = append(domains, domain)
    }
    
    return domains, nil
}
```

#### Connection Pool Configuration
```go
// pkg/db8/db8.go - Add connection pool settings
func (d *Db8) OpenConnection() (*sql.DB, error) {
    db, err := sql.Open("postgres", d.GetConnectionString())
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(25)                 // Maximum open connections
    db.SetMaxIdleConns(25)                 // Maximum idle connections
    db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
    db.SetConnMaxIdleTime(5 * time.Minute) // Idle connection timeout
    
    return db, nil
}
```

### 2. Scanning Optimizations

#### Implement Streaming for Large Datasets
```go
// pkg/model8/results8.go - Add streaming support
type ResultStream struct {
    Domain   string
    Hostname string
    Source   string
}

func (r *PassiveRunner) RunPassiveEnumStream(ctx context.Context, results chan<- ResultStream) error {
    defer close(results)
    
    for _, domain := range r.SeedDomains {
        // Stream results as they come
        go func(d string) {
            // Process domain and stream results
            for hostname := range processedResults {
                select {
                case results <- ResultStream{Domain: d, Hostname: hostname, Source: "passive"}:
                case <-ctx.Done():
                    return
                }
            }
        }(domain)
    }
    
    return nil
}
```

#### Implement Proper Temp File Cleanup
```go
// pkg/active/active.go - Add cleanup mechanism
import (
    "os"
    "path/filepath"
)

type ActiveRunner struct {
    SeedDomains []string
    Results     int
    Subdomains  map[string][]string
    tempFiles   []string // Track temp files
}

func (r *ActiveRunner) RunActiveEnum(wordlist string, threads int, prevResults map[string][]string) map[string][]string {
    // Ensure cleanup on exit
    defer r.cleanup()
    
    // ... existing code ...
    
    for _, domain := range r.SeedDomains {
        tempFile := "./tmp/tempDomain-" + domain + ".txt"
        r.tempFiles = append(r.tempFiles, tempFile) // Track file
        utils.WriteTempFile(tempFile, results.Hostnames[domain])
    }
    
    // ... rest of function ...
}

func (r *ActiveRunner) cleanup() {
    for _, file := range r.tempFiles {
        if err := os.Remove(file); err != nil {
            log8.BaseLogger.Warn().Msgf("Failed to cleanup temp file %s: %v", file, err)
        }
    }
    r.tempFiles = nil
}
```

#### Add Context-Based Cancellation
```go
// pkg/passive/passive.go - Add context support
func (r *PassiveRunner) RunPassiveEnum(ctx context.Context, prevResults map[string][]string) map[string][]string {
    var wg sync.WaitGroup
    var results model8.Result8
    results.Hostnames = make(map[string][]string)
    
    for _, domain := range r.SeedDomains {
        select {
        case <-ctx.Done():
            return results.Hostnames
        default:
            wg.Add(2)
            sf_results := make(chan string)
            go subfinder.RunSubfinderInWithContext(ctx, domain, sf_results, &wg)
            go subfinder.RunSubfinderOutWithContext(ctx, domain, sf_results, &results, &wg)
        }
    }
    
    wg.Wait()
    return results.Hostnames
}
```

### 3. Memory Optimization

#### Implement Result Batching
```go
// pkg/utils/batch.go - New file
type BatchProcessor struct {
    batchSize int
    processor func([]string) error
}

func NewBatchProcessor(size int, processor func([]string) error) *BatchProcessor {
    return &BatchProcessor{
        batchSize: size,
        processor: processor,
    }
}

func (bp *BatchProcessor) Process(items []string) error {
    for i := 0; i < len(items); i += bp.batchSize {
        end := i + bp.batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        if err := bp.processor(batch); err != nil {
            return err
        }
    }
    return nil
}
```

#### Optimize Data Structures
```go
// pkg/model8/results8.go - Optimize for memory usage
type HostnameEntry struct {
    Domain   string    `json:"domain"`
    Hostname string    `json:"hostname"`
    Source   string    `json:"source"`
    LastSeen time.Time `json:"last_seen"`
}

// Use sync.Pool for object reuse
var hostnamePool = sync.Pool{
    New: func() interface{} {
        return &HostnameEntry{}
    },
}

func GetHostnameEntry() *HostnameEntry {
    return hostnamePool.Get().(*HostnameEntry)
}

func PutHostnameEntry(entry *HostnameEntry) {
    entry.Domain = ""
    entry.Hostname = ""
    entry.Source = ""
    entry.LastSeen = time.Time{}
    hostnamePool.Put(entry)
}
```

### 4. I/O Optimization

#### Async File Operations
```go
// pkg/utils/async_io.go - New file
import (
    "context"
    "io"
    "os"
)

type AsyncFileWriter struct {
    writes chan writeRequest
    done   chan struct{}
}

type writeRequest struct {
    filename string
    data     []byte
    response chan error
}

func NewAsyncFileWriter(workers int) *AsyncFileWriter {
    afw := &AsyncFileWriter{
        writes: make(chan writeRequest, 100),
        done:   make(chan struct{}),
    }
    
    for i := 0; i < workers; i++ {
        go afw.worker()
    }
    
    return afw
}

func (afw *AsyncFileWriter) worker() {
    for req := range afw.writes {
        err := os.WriteFile(req.filename, req.data, 0644)
        req.response <- err
    }
}

func (afw *AsyncFileWriter) WriteFile(filename string, data []byte) error {
    response := make(chan error, 1)
    afw.writes <- writeRequest{
        filename: filename,
        data:     data,
        response: response,
    }
    return <-response
}
```

### 5. Monitoring and Metrics

#### Add Performance Metrics
```go
// pkg/metrics/metrics.go - New file
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    scanDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "asmm8_scan_duration_seconds",
            Help: "Duration of scanning operations",
        },
        []string{"scan_type", "domain"},
    )
    
    hostnamesFound = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "asmm8_hostnames_found_total",
            Help: "Total number of hostnames found",
        },
        []string{"domain", "source"},
    )
    
    dbOperations = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "asmm8_db_operation_duration_seconds",
            Help: "Duration of database operations",
        },
        []string{"operation"},
    )
)

func RecordScanDuration(scanType, domain string, duration time.Duration) {
    scanDuration.WithLabelValues(scanType, domain).Observe(duration.Seconds())
}

func IncrementHostnamesFound(domain, source string) {
    hostnamesFound.WithLabelValues(domain, source).Inc()
}

func RecordDBOperation(operation string, duration time.Duration) {
    dbOperations.WithLabelValues(operation).Observe(duration.Seconds())
}
```

## Performance Testing

### 1. Benchmarking
```go
// pkg/db8/db8_domain8_test.go - Add benchmarks
func BenchmarkGetAllDomain(b *testing.B) {
    // Setup test database
    db := setupTestDB()
    defer db.Close()
    
    domain8 := NewDb8Domain8(db)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := domain8.GetAllDomain()
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkExistEnabled(b *testing.B) {
    db := setupTestDB()
    defer db.Close()
    
    domain8 := NewDb8Domain8(db)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = domain8.ExistEnabled()
    }
}
```

### 2. Load Testing
```bash
# Use tools like ab or wrk for API load testing
ab -n 1000 -c 10 http://localhost:8000/api/domains

# Database load testing
pgbench -i -s 10 cptm8
pgbench -c 10 -j 2 -t 1000 cptm8
```

### 3. Profiling
```go
// main.go - Add profiling support
import (
    _ "net/http/pprof"
    "net/http"
)

func main() {
    // Start profiling server
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // ... rest of main function
}
```

## Monitoring Dashboard

### Key Metrics to Monitor
1. **Database Performance**
   - Connection pool usage
   - Query execution times
   - Lock contention

2. **Scanning Performance**
   - Scan completion times
   - Hostnames discovered per scan
   - Error rates

3. **System Resources**
   - Memory usage
   - CPU utilization
   - Disk I/O
   - Network I/O

4. **Application Metrics**
   - API response times
   - Queue lengths
   - Active goroutines

### Sample Grafana Queries
```promql
# Average scan duration by type
rate(asmm8_scan_duration_seconds_sum[5m]) / rate(asmm8_scan_duration_seconds_count[5m])

# Database operation percentiles
histogram_quantile(0.95, rate(asmm8_db_operation_duration_seconds_bucket[5m]))

# Hostnames found rate
rate(asmm8_hostnames_found_total[5m])
```

## Performance Improvement Roadmap

### Phase 1 (Immediate - 1 week)
- [ ] Fix database resource leaks
- [ ] Implement temp file cleanup
- [ ] Add connection pool configuration
- [ ] Optimize database queries

### Phase 2 (Short-term - 1 month)
- [ ] Implement streaming for large datasets
- [ ] Add context-based cancellation
- [ ] Implement async I/O operations
- [ ] Add basic metrics collection

### Phase 3 (Medium-term - 3 months)
- [ ] Implement result batching
- [ ] Add comprehensive monitoring
- [ ] Implement caching strategies
- [ ] Add load balancing for scanning

### Phase 4 (Long-term - 6 months)
- [ ] Implement distributed scanning
- [ ] Add advanced caching (Redis)
- [ ] Implement auto-scaling
- [ ] Add predictive performance analysis

## Conclusion

The ASMM8 system has good foundational performance characteristics with effective use of Go's concurrency features. However, addressing the identified issues - particularly database resource management and temporary file cleanup - will significantly improve overall system performance and reliability.

The recommended optimizations should be implemented incrementally, starting with the immediate fixes and progressing through the phases outlined in the roadmap. Regular monitoring and profiling will help identify additional optimization opportunities as the system scales.