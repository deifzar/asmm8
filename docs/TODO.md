# ASMM8 TODO & Action Items

**Created:** July 18, 2025
**Last Updated:** November 7, 2025
**Status:** Active Development

## Overview

This document tracks all identified issues, improvements, and action items for the ASMM8 project. Items are categorized by priority and include implementation details, affected files, and success criteria.

## ✅ Recently Completed (November 2025)

### RabbitMQ Manual Acknowledgment Mode
**Status:** ✅ **COMPLETED** (November 7, 2025)
**Priority:** 🔴 CRITICAL
**Issue:** Messages were auto-acknowledged regardless of scan success/failure, causing message loss on crashes
**Impact:** Container crashes in Kubernetes resulted in lost workflow messages
**Effort:** 8 hours

**Implemented Solution:**
- ✅ Manual acknowledgment mode enabled (`autoack: "false"` in configuration)
- ✅ Delivery tag tracking from RabbitMQ consumer through HTTP headers to controller
- ✅ Smart ACK/NACK logic in defer block of `Active()` function
- ✅ ACK: Scan completed (successfully or with handled errors like DB errors, no domains, tool errors)
- ✅ NACK + requeue: Scan crashed, panicked, or was interrupted by SIGTERM
- ✅ NACK no requeue: Handler failed or no handler found (permanent failures)
- ✅ Panic recovery in defer block ensures ACK/NACK always executes
- ✅ Guard clauses prevent ACK/NACK for non-RabbitMQ triggered scans

**Files Modified:**
- `configs/configuration.yaml` - Set autoack to false
- `pkg/orchestrator8/orchestrator8.go` - Added `AckScanCompletion()` and `NackScanMessage()` methods
- `pkg/orchestrator8/orchestrator8.go` - Modified handler to pass delivery tag via HTTP header
- `pkg/amqpM8/pooled_amqp.go` - Removed auto-ACK, implemented smart handler ACK/NACK
- `pkg/controller8/controller8_asmm8.go` - Extract delivery tag, pass to `Active()`
- `pkg/controller8/controller8_asmm8.go` - Defer block with ACK/NACK logic based on completion status
- `pkg/controller8/controller8_asmm8.go` - All early return paths now ACK/NACK appropriately

**Benefits:**
- No message loss during container crashes or Kubernetes pod restarts
- Failed scans automatically requeued for retry
- Completed scans (even with warnings) properly acknowledged
- Workflow never stalls due to unacknowledged messages

---

### External Tool Error Propagation
**Status:** ✅ **COMPLETED** (November 7, 2025)
**Priority:** 🔴 CRITICAL
**Issue:** External tool failures (subfinder, dnsx, alterx) were silently swallowed, not propagated
**Impact:** Scan failures went undetected, `scanFailed` flag never set
**Effort:** 4 hours

**Implemented Solution:**
- ✅ Modified `RunPassiveEnum()` to return `(map[string][]string, error)`
- ✅ Modified `RunActiveEnum()` to return `(map[string][]string, error)`
- ✅ Thread-safe error capture in tool wrappers using `sync.Mutex`
- ✅ Error propagation from subfinder, dnsx, alterx back to controller
- ✅ `Active()` function now sets `scanFailed = true` when tools fail
- ✅ Defer block uses `scanFailed` flag to determine ACK/NACK behavior

**Files Modified:**
- `pkg/passive/passive.go` - Added error return and propagation
- `pkg/passive/subfinder/subfinder.go` - Thread-safe error capture
- `pkg/active/active.go` - Added error return and propagation
- `pkg/active/dnsx/dnsx.go` - Thread-safe error capture
- `pkg/active/alterx/alterx.go` - Thread-safe error capture for both alterx and dnsx-after-alterx
- `pkg/controller8/controller8_asmm8.go` - Handle returned errors, set `scanFailed` flag

**Benefits:**
- External tool failures are now detected and tracked
- Scan status accurately reflects tool failures
- Operators can identify which tools failed via logs
- Failed scans are properly handled with warnings

---

## Critical Issues (Fix Immediately)

### 1. Database Resource Leaks
**Priority:** 🔴 CRITICAL  
**Issue:** Missing `defer rows.Close()` statements causing connection leaks  
**Files:** `pkg/db8/db8_domain8.go`, `pkg/db8/db8_hostname8.go`  
**Impact:** Connection pool exhaustion, system instability  
**Effort:** 1 hour  

**Action Items:**
- [ ] Add `defer query.Close()` to `GetAllDomain()` (line 41)
- [ ] Add `defer query.Close()` to `GetAllEnabled()` (line 68)
- [ ] Add `defer query.Close()` to `GetOneDomain()` (line 106)
- [ ] Review all database query functions for similar issues
- [ ] Add database connection monitoring

**Code Fix:**
```go
func (m *Db8Domain8) GetAllDomain() ([]model8.Domain8, error) {
    query, err := m.Db.Query("SELECT id, name, companyname, enabled FROM cptm8domain")
    if err != nil {
        return []model8.Domain8{}, err
    }
    defer query.Close() // ADD THIS LINE
    // ... rest of function
}
```

### 2. Security: Credentials in Configuration
**Priority:** 🔴 CRITICAL  
**Issue:** Plain text credentials in configuration files  
**Files:** `configuration.yaml`, `pkg/db8/db8.go`, `pkg/orchestrator8/orchestrator8.go`  
**Impact:** Security vulnerability, credential exposure  
**Effort:** 2 hours  

**Action Items:**
- [ ] Move database password to environment variable
- [ ] Move RabbitMQ password to environment variable
- [ ] Update configuration parsing to check environment variables
- [ ] Add example configuration with placeholder values
- [ ] Update documentation for credential management

**Implementation:**
```go
// pkg/db8/db8.go
func (d *Db8) InitDatabase8(l string, port int, sc, db, u string) {
    d.password = os.Getenv("DB_PASSWORD")
    if d.password == "" {
        log.Fatal("DB_PASSWORD environment variable not set")
    }
    // ... rest of function
}
```

### 3. Excessive Fatal() Calls
**Priority:** 🔴 CRITICAL  
**Issue:** Application terminates instead of graceful error handling  
**Files:** `cmd/launch.go`, `pkg/controller8/controller8_asmm8.go`  
**Impact:** Poor user experience, system instability  
**Effort:** 3 hours  

**Action Items:**
- [ ] Replace Fatal() with proper error returns in controllers
- [ ] Implement graceful error handling in API endpoints
- [ ] Add error recovery mechanisms
- [ ] Update error messages to be more descriptive
- [ ] Add error logging without termination

## High Priority Issues

### 4. Temporary File Cleanup
**Priority:** 🟠 HIGH  
**Issue:** Temporary files created without cleanup mechanism  
**Files:** `pkg/active/active.go`  
**Impact:** Disk space exhaustion, performance degradation  
**Effort:** 2 hours  

**Action Items:**
- [ ] Implement cleanup mechanism in ActiveRunner
- [ ] Add defer statements for temp file removal
- [ ] Create temp file tracking system
- [ ] Add cleanup on application shutdown
- [ ] Add monitoring for temp file usage

### 5. Input Validation
**Priority:** 🟠 HIGH  
**Issue:** Limited validation for domain names and user inputs  
**Files:** `pkg/model8/domain8.go`, `pkg/controller8/`  
**Impact:** Security vulnerability, data integrity issues  
**Effort:** 4 hours  

**Action Items:**
- [ ] Add domain name validation (RFC 1123)
- [ ] Implement custom Gin validators
- [ ] Add input sanitization functions
- [ ] Validate all API parameters
- [ ] Add request size limits

### 6. Code Duplication in Database Layer
**Priority:** 🟠 HIGH  
**Issue:** Duplicate code in `GetAllDomain()` and `GetAllEnabled()`  
**Files:** `pkg/db8/db8_domain8.go`  
**Impact:** Maintenance overhead, inconsistency risk  
**Effort:** 2 hours  

**Action Items:**
- [ ] Create consolidated `GetDomains(enabledOnly bool)` function
- [ ] Refactor existing functions to use consolidated version
- [ ] Update all calling code
- [ ] Add unit tests for new function
- [ ] Remove duplicate code

## Medium Priority Issues

### 7. Performance Optimization
**Priority:** 🟡 MEDIUM  
**Issue:** Various performance bottlenecks identified  
**Files:** Multiple  
**Impact:** Slow processing, resource usage  
**Effort:** 1 week  

**Action Items:**
- [ ] Implement database connection pooling
- [ ] Add result streaming for large datasets
- [ ] Optimize database queries (ExistEnabled)
- [ ] Implement async I/O operations
- [ ] Add memory usage optimization

### 8. Error Handling Standardization
**Priority:** 🟡 MEDIUM  
**Issue:** Inconsistent error handling patterns  
**Files:** Multiple  
**Impact:** Debugging difficulty, maintenance issues  
**Effort:** 1 week  

**Action Items:**
- [ ] Define standard error types
- [ ] Create error handling utilities
- [ ] Implement consistent error logging
- [ ] Add error context information
- [ ] Create error recovery strategies

### 9. Monitoring & Observability
**Priority:** 🟡 MEDIUM  
**Issue:** Limited monitoring and metrics  
**Files:** New files needed  
**Impact:** Operational visibility, debugging difficulty  
**Effort:** 1 week  

**Action Items:**
- [ ] Add Prometheus metrics
- [ ] Implement health check endpoints
- [ ] Add performance monitoring
- [ ] Create monitoring dashboard
- [ ] Add alerting rules

### 10. Testing Infrastructure
**Priority:** 🟡 MEDIUM  
**Issue:** No unit or integration tests  
**Files:** Test files needed  
**Impact:** Code quality, reliability  
**Effort:** 2 weeks  

**Action Items:**
- [ ] Create test database setup
- [ ] Add unit tests for database layer
- [ ] Add unit tests for controllers
- [ ] Create integration tests for APIs
- [ ] Add benchmark tests
- [ ] Set up CI/CD pipeline

## Low Priority Issues

### 11. Code Quality Improvements
**Priority:** 🟢 LOW  
**Issue:** Various code quality issues  
**Files:** Multiple  
**Impact:** Maintenance, readability  
**Effort:** 1 week  

**Action Items:**
- [ ] Fix naming inconsistencies (ASSM8 vs ASMM8)
- [ ] Remove commented-out code
- [ ] Add missing documentation
- [ ] Standardize variable names
- [ ] Add code linting

### 12. Configuration Enhancements
**Priority:** 🟢 LOW  
**Issue:** Configuration system improvements  
**Files:** `pkg/configparser/configparser.go`  
**Impact:** Deployment flexibility  
**Effort:** 3 days  

**Action Items:**
- [ ] Add configuration validation
- [ ] Support multiple configuration formats
- [ ] Add environment-specific configs
- [ ] Implement configuration hot-reload
- [ ] Add configuration documentation

## Feature Requests

### 13. API Authentication
**Priority:** 🟡 MEDIUM  
**Issue:** No authentication mechanism  
**Files:** New files needed  
**Impact:** Security, access control  
**Effort:** 1 week  

**Action Items:**
- [ ] Implement JWT authentication
- [ ] Add user management
- [ ] Create role-based access control
- [ ] Add API key management
- [ ] Implement session management

### 14. Result Caching
**Priority:** 🟢 LOW  
**Issue:** No caching mechanism for results  
**Files:** New files needed  
**Impact:** Performance, user experience  
**Effort:** 1 week  

**Action Items:**
- [ ] Implement Redis caching
- [ ] Add cache invalidation logic
- [ ] Create cache configuration
- [ ] Add cache monitoring
- [ ] Implement cache warming

### 15. API Documentation
**Priority:** 🟡 MEDIUM  
**Issue:** No API documentation  
**Files:** New files needed  
**Impact:** Developer experience  
**Effort:** 3 days  

**Action Items:**
- [ ] Add OpenAPI/Swagger documentation
- [ ] Create API usage examples
- [ ] Add endpoint documentation
- [ ] Create client SDKs
- [ ] Add interactive API explorer

## Production Readiness Roadmap

This roadmap outlines the prioritized path to make ASMM8 production-ready, focusing on critical fixes first, then building towards enterprise-grade reliability and features.

### Phase 1: Critical Stability (Weeks 1-2)
**Goal:** Fix critical issues that prevent stable operation
**Deliverables:** Stable API service without crashes or resource leaks

#### Week 1: Resource Management & Security
- [ ] **Database Resource Leaks** - Fix connection leaks (#1)
- [ ] **Credential Security** - Move passwords to environment variables (#2)
- [ ] **Error Handling** - Replace Fatal() calls with graceful error handling (#3)
- [ ] **File Cleanup** - Implement temporary file cleanup mechanism (#4)

**Success Criteria:** 
- Service runs for 24+ hours without crashes
- No connection pool exhaustion
- No credentials in configuration files

#### Week 2: Input Security & Validation  
- [ ] **Input Validation** - Add comprehensive domain/parameter validation (#5)
- [ ] **Code Consolidation** - Eliminate duplicate database code (#6)
- [ ] **Security Hardening** - Add input sanitization for external tools
- [ ] **Basic Monitoring** - Add health check endpoints

**Success Criteria:**
- All API inputs validated and sanitized
- Malicious inputs safely rejected
- Basic health monitoring operational

### Phase 2: Production Foundation (Weeks 3-6)
**Goal:** Build reliable, observable, and testable foundation
**Deliverables:** Production-ready service with monitoring and tests

#### Week 3: Performance & Observability
- [ ] **Connection Pooling** - Implement database connection pooling (#7)
- [ ] **Metrics Collection** - Add Prometheus metrics (#9)
- [ ] **Performance Monitoring** - Add request/response time tracking
- [ ] **Resource Monitoring** - Monitor memory, CPU, and disk usage

#### Week 4: Testing Infrastructure
- [ ] **Unit Test Framework** - Set up testing infrastructure (#10)
- [ ] **Database Tests** - Add unit tests for database layer
- [ ] **Controller Tests** - Add unit tests for business logic
- [ ] **API Tests** - Add basic integration tests

#### Week 5: Error Handling & Logging
- [ ] **Error Standardization** - Implement consistent error handling (#8)
- [ ] **Structured Logging** - Enhance logging with context and levels
- [ ] **Error Recovery** - Add graceful degradation mechanisms
- [ ] **Audit Logging** - Add security audit trail

#### Week 6: Documentation & API Standards
- [ ] **API Documentation** - Complete OpenAPI/Swagger docs (#15)
- [ ] **Code Documentation** - Add comprehensive code comments
- [ ] **Deployment Guide** - Create production deployment documentation
- [ ] **Operations Manual** - Document monitoring and troubleshooting

**Success Criteria:**
- 80%+ test coverage
- Complete monitoring dashboard
- Full API documentation
- Zero-downtime deployment capability

### Phase 3: Enterprise Features (Weeks 7-10)
**Goal:** Add enterprise-grade security and features
**Deliverables:** Secure, authenticated, high-performance service

#### Week 7: Authentication & Authorization
- [ ] **JWT Authentication** - Implement user authentication (#13)
- [ ] **Role-Based Access** - Add RBAC for API endpoints
- [ ] **API Key Management** - Add API key authentication option
- [ ] **Session Management** - Implement secure session handling

#### Week 8: Performance Optimization
- [ ] **Result Caching** - Implement Redis-based caching (#14)
- [ ] **Query Optimization** - Optimize database queries for scale
- [ ] **Async Processing** - Add background job processing
- [ ] **Rate Limiting** - Implement API rate limiting

#### Week 9: Advanced Monitoring
- [ ] **Distributed Tracing** - Add request tracing across services
- [ ] **Alert Management** - Implement alerting rules and notifications
- [ ] **Performance Profiling** - Add runtime performance profiling
- [ ] **Capacity Planning** - Add resource usage analytics

#### Week 10: Production Hardening
- [ ] **CI/CD Pipeline** - Complete automated deployment pipeline
- [ ] **Load Testing** - Perform comprehensive load testing
- [ ] **Security Scanning** - Add automated security scans
- [ ] **Backup Strategy** - Implement data backup and recovery

**Success Criteria:**
- Authentication system operational
- 95%+ uptime under production load
- Complete CI/CD automation
- Security audit passed

### Phase 4: Scale & Optimization (Weeks 11-12)
**Goal:** Optimize for production scale and operational excellence
**Deliverables:** Scalable, enterprise-ready service

#### Week 11: Scalability
- [ ] **Horizontal Scaling** - Add support for multiple service instances
- [ ] **Database Clustering** - Implement database clustering support
- [ ] **Load Balancing** - Add intelligent load balancing
- [ ] **Auto-Scaling** - Implement auto-scaling based on metrics

#### Week 12: Operational Excellence
- [ ] **Disaster Recovery** - Complete disaster recovery procedures
- [ ] **Performance Benchmarks** - Establish performance baselines
- [ ] **Documentation Review** - Complete all operational documentation
- [ ] **Security Certification** - Complete security compliance review

**Success Criteria:**
- Service scales to handle 10x current load
- RTO < 15 minutes, RPO < 5 minutes
- Complete operational runbooks
- Security compliance achieved

## Implementation Schedule

### Week 1: Critical Issues
- [ ] Fix database resource leaks
- [ ] Implement credential management
- [ ] Replace Fatal() calls with proper error handling
- [ ] Add temporary file cleanup

### Week 2: High Priority Security & Validation
- [ ] Add comprehensive input validation
- [ ] Implement external tool input sanitization
- [ ] Add API parameter validation
- [ ] Consolidate database code

### Week 3: Performance & Monitoring
- [ ] Implement connection pooling
- [ ] Implement Context timeout for DB connections and transactions
- [ ] Add basic metrics collection
- [ ] Create health check endpoints
- [ ] Optimize database queries

### Week 4: Testing & Documentation
- [ ] Create unit test framework
- [ ] Add basic test coverage
- [ ] Complete API documentation
- [ ] Add monitoring dashboard

### Month 2: Advanced Features
- [ ] Implement authentication
- [ ] Add comprehensive monitoring
- [ ] Create integration tests
- [ ] Add result caching

### Month 3: Production Readiness
- [ ] Complete test coverage
- [ ] Add performance benchmarks
- [ ] Implement CI/CD pipeline
- [ ] Add deployment automation

## Success Criteria

### Critical Issues Fixed
- [ ] No database connection leaks
- [ ] No credentials in configuration files
- [ ] No Fatal() calls in business logic
- [ ] Temp files properly cleaned up

### Security Hardened
- [ ] All inputs validated
- [ ] External tools secured
- [ ] Authentication implemented
- [ ] Audit logging in place

### Performance Optimized
- [ ] Database queries optimized
- [ ] Connection pooling implemented
- [ ] Memory usage optimized
- [ ] Response times improved

### Production Ready
- [ ] 90%+ test coverage
- [ ] Monitoring dashboard operational
- [ ] CI/CD pipeline functional
- [ ] Documentation complete

## Tracking

### Completed Items
- [x] ~~Create documentation structure~~ (2025-07-18)
- [x] ~~Complete code review~~ (2025-07-18)
- [x] ~~Document security issues~~ (2025-07-18)
- [x] ~~Create performance analysis~~ (2025-07-18)
- [x] ~~Document architecture~~ (2025-07-18)

### In Progress
- [ ] Working on critical database fixes
- [ ] Implementing credential management
- [ ] Adding input validation

### Blocked/Waiting
- [ ] Waiting for external tool security review
- [ ] Waiting for deployment environment setup

## Notes

### Development Environment Setup
- Ensure PostgreSQL is running locally
- Ensure RabbitMQ is running locally
- Install external tools: subfinder, amass, dnsx, alterx, httpx
- Set up Go development environment

### Testing Environment
- Create test database schema
- Mock external tools for testing
- Set up CI/CD environment
- Configure monitoring stack

### Production Considerations
- Plan for zero-downtime deployments
- Implement proper logging levels
- Configure resource limits
- Set up backup strategies

---

**Last Review:** November 7, 2025
**Next Review:** November 14, 2025
**Assigned:** Development Team
**Status:** Active Development