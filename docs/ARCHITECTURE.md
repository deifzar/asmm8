# ASMM8 Architecture Documentation

**Date:** July 18, 2025
**Last Updated:** November 7, 2025
**Version:** 1.1

## Overview

ASMM8 (Asset Surface Management Mate) is a Go-based microservice designed for automated subdomain enumeration and reconnaissance. It follows a modular architecture with clear separation of concerns and leverages modern Go patterns for concurrency and scalability.

## System Architecture

### High-Level Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Load Balancer │    │   Other Services│
│                 │    │                 │    │   (naabum8)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                                 ▼
                    ┌─────────────────┐
                    │   ASMM8 API     │
                    │   (Gin Router)  │
                    └─────────────────┘
                                 │
                 ┌───────────────┼───────────────┐
                 │               │               │
                 ▼               ▼               ▼
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │   Controllers   │ │   Orchestrator  │ │   Notification  │
    │   (Business     │ │   (RabbitMQ)    │ │   System        │
    │   Logic)        │ └─────────────────┘ └─────────────────┘
    └─────────────────┘          │
                 │               │
                 ▼               ▼
    ┌─────────────────┐ ┌─────────────────┐
    │   Database      │ │   Scanning      │
    │   Layer (DB8)   │ │   Engines       │
    └─────────────────┘ └─────────────────┘
                                 │
                 ┌───────────────┼───────────────┐
                 │               │               │
                 ▼               ▼               ▼
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │   Passive       │ │   Active        │ │   External      │
    │   Enumeration   │ │   Enumeration   │ │   Tools         │
    │   (Subfinder,   │ │   (DNSx,        │ │   (amass,       │
    │   Amass)        │ │   Alterx)       │ │   httpx)        │
    └─────────────────┘ └─────────────────┘ └─────────────────┘
```

## Component Architecture

### 1. API Layer (`pkg/api8/`)

**Purpose:** HTTP API server and routing  
**Framework:** Gin Web Framework  
**Responsibilities:**
- HTTP request handling
- Route definition and middleware
- Request/response serialization
- Database connection management

**Key Files:**
- `api8.go` - Main API server implementation
- Route handlers defined in controllers

**Design Patterns:**
- Dependency Injection for database and config
- Middleware pattern for cross-cutting concerns
- RESTful API design

**Code Size:** 132 lines

### 2. Controller Layer (`pkg/controller8/`)

**Purpose:** Business logic and request processing  
**Pattern:** Interface-based design with dependency injection  
**Responsibilities:**
- Request validation
- Business logic execution
- Response formatting
- Error handling

**Key Components:**
- `controller8_asmm8.go` (442 lines) - Main scan orchestration
- `controller8_domain8.go` (157 lines) - Domain management
- `controller8_hostname8.go` (176 lines) - Hostname operations with batch support

**Total Code Size:** 775 lines across 6 files (including interfaces)

**Interface Design:**
```go
type Controller8ASMM8Interface interface {
    LaunchScan(c *gin.Context)
    LaunchActive(c *gin.Context)
    LaunchPassive(c *gin.Context)
    // ... other methods
}
```

### 3. Database Layer (`pkg/db8/`)

**Purpose:** Data persistence and database operations  
**Technology:** PostgreSQL with SQLX  
**Pattern:** Repository pattern with interfaces  

**Key Components:**
- `db8.go` (78 lines) - Connection management with retry logic (10 attempts, 5s intervals)
- `db8_domain8.go` (190 lines) - Domain operations
- `db8_hostname8.go` (367 lines) - Hostname operations with batch insertions
- `db8_generalscansettings.go` - Configuration persistence

**Total Code Size:** 730 lines across 7 files (including interfaces)

**Connection Features:**
- PostgreSQL driver: `github.com/lib/pq`
- 10 retry attempts with 5-second intervals between retries
- 30-second connection timeout
- Schema-specific connection support

**Database Schema:**
```sql
-- Domains table
CREATE TABLE cptm8domain (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    companyname VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Hostnames table  
CREATE TABLE cptm8hostname (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID REFERENCES cptm8domain(id),
    hostname VARCHAR(255) NOT NULL,
    source VARCHAR(50) NOT NULL,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4. Scanning Engine Architecture

#### Passive Enumeration (`pkg/passive/`)
**Purpose:** Subdomain discovery using external APIs and databases  
**Tools:** Subfinder, Amass  
**Pattern:** Producer-Consumer with channels  

**Workflow:**
```
Domain Input → Subfinder → Channel → Result Aggregation → Database Storage
              ↓
           Amass → Channel → Result Aggregation → Database Storage
```

**Implementation Pattern:**
```go
func (r *PassiveRunner) RunPassiveEnum(prevResults map[string][]string) map[string][]string {
    var wg sync.WaitGroup
    for _, domain := range r.SeedDomains {
        wg.Add(2)
        results := make(chan string)
        go subfinder.RunSubfinderIn(domain, results, &wg)
        go subfinder.RunSubfinderOut(domain, results, &results, &wg)
    }
    wg.Wait()
    return results.Hostnames
}
```

#### Active Enumeration (`pkg/active/`)
**Purpose:** Subdomain discovery through DNS brute-forcing and permutation  
**Tools:** DNSx, Alterx, HTTPx  
**Pattern:** Pipeline processing with temporary file coordination  

**Workflow:**
```
Previous Results → DNS Brute Force → Temp Files → Alterx Permutation → Live Check → Results
                       ↓                           ↓                     ↓
                    DNSx Tool           Alterx Tool             HTTPx Tool
```

### 5. Message Queue Architecture (`pkg/orchestrator8/`, `pkg/amqpM8/`)

**Purpose:** Inter-service communication and task coordination
**Technology:** RabbitMQ with AMQP 0.9.1 (`github.com/rabbitmq/amqp091-go v1.10.0`)
**Pattern:** Topic-based routing with exchanges and advanced connection pooling

**Total Code Size:** 2,036 lines
- `pkg/amqpM8` - 1,777 lines (5 files)
  - `pooled_amqp.go` (842 lines) - **Largest file in codebase**
  - `connection_pool.go` (406 lines) - Pool management
  - `shared_state.go` (241 lines) - Thread-safe state
  - `pool_manager.go` (190 lines) - Pool creation
  - `initialization.go` (98 lines) - Initialization
- `pkg/orchestrator8` - 259 lines (2 files)

**Advanced Features:**
- Connection pooling with configurable limits (default: 2-10 connections)
- Automatic connection recovery on failures
- Periodic health checks (default: 30 minutes)
- Consumer auto-recovery with lifecycle management
- Thread-safe operations using sync.Mutex
- Configurable retry logic (3 attempts, 2s delay by default)
- **Manual acknowledgment mode** with smart ACK/NACK logic
- Delivery tag tracking for message lifecycle management
- Automatic requeue on scan failures with exponential backoff  

**Exchange Configuration:**
```yaml
Exchanges:
  cptm8: "topic"        # Main processing exchange
  notification: "topic"  # Notification exchange

Queues:
  qasmm8:               # ASMM8 processing queue
    - exchange: cptm8
    - routing_key: "cptm8.asmm8.#"
    - max_length: 1
    - overflow: reject-publish
    - consumer: casmm8 (non-exclusive)
```

**Message Flow with Manual Acknowledgment:**
```
RabbitMQ Message → Consumer (pooled_amqp.go)
                        ↓
                   Extract deliveryTag
                        ↓
                HTTP Request (with X-RabbitMQ-Delivery-Tag header)
                        ↓
                   LaunchScan() → Extract deliveryTag from header
                        ↓
                   Active() → Defer with ACK/NACK logic
                        ↓
        ┌───────────────┴───────────────┐
        │                               │
    Scan Completes              Scan Fails (crash/SIGTERM)
        │                               │
    ACK message                    NACK + requeue
        ↓                               ↓
    Remove from queue              Return to queue for retry
        ↓
    Publish to naabum8
```

**ACK/NACK Decision Logic:**
- **ACK (completed=true)**: Scan finished successfully OR scan failed early (DB error, no domains, tool error)
- **NACK + requeue (completed=false)**: Scan crashed, panicked, or was interrupted by SIGTERM
- **NACK no requeue**: Handler failed or no handler found (permanent failures)

**Delivery Tag Tracking:**
1. RabbitMQ assigns unique deliveryTag to each message
2. Consumer extracts deliveryTag and passes via HTTP header
3. Controller extracts deliveryTag and passes to Active() function
4. Defer block in Active() uses deliveryTag to ACK/NACK after scan completes/fails
5. Guard clause ensures only RabbitMQ-triggered scans are acknowledged

### 6. Configuration Management (`pkg/configparser/`)

**Purpose:** Application configuration and environment management  
**Technology:** Viper with YAML configuration  
**Pattern:** Centralized configuration with environment overrides  

**Configuration Structure:**
```yaml
APP_ENV: DEV|TEST|PROD
LOG_LEVEL: "0-5"
ASMM8:
  runType: fast|complete
  activeWordList: filename
  activeThreads: integer
Database:
  location: hostname
  port: integer
  # ... other database settings
RabbitMQ:
  location: hostname
  port: integer
  # ... other messaging settings
```

### 7. Logging Architecture (`pkg/log8/`)

**Purpose:** Structured logging and observability
**Technology:** Zerolog (`github.com/rs/zerolog v1.33.0`) with log rotation
**Pattern:** Singleton logger with sync.Once initialization
**Code Size:** 99 lines

**Features:**
- Dual output: Console + file ([log/asmm8.log](../log/asmm8.log))
- Log rotation with `lumberjack.v2`
- File permissions: 0640 (readable by log aggregation tools)
- Configurable via LOG_LEVEL in configuration.yaml  

**Log Levels:**
- `-1` Trace (development debugging)
- `0` Debug (detailed information)
- `1` Info (general information)
- `2` Warn (warning conditions)
- `3` Error (error conditions)
- `4` Fatal (critical errors)
- `5` Panic (system panic)

## Data Flow Architecture

### 1. Scan Initiation Flow
```
HTTP Request → API Layer → Controller → Domain Validation → Tool Installation
                                                                     ↓
RabbitMQ Queue Check → Orchestrator → Scan Type Decision → Passive/Active Engine
                                                                     ↓
External Tools → Result Collection → Database Storage → Notification
```

### 2. Result Processing Flow
```
External Tool Output → Channel → Result Aggregation → Deduplication → Database → API Response
                                                                            ↓
                                                                    Notification System
```

### 3. Inter-Service Communication
```
ASMM8 Service → RabbitMQ Exchange → Routing Key → Target Service Queue → Consumer
                                                                               ↓
                                                                      naabum8 Service
```

## Security Architecture

### 1. Authentication & Authorization
**Current State:** Not implemented  
**Recommended:** JWT-based authentication with role-based access control

### 2. Input Validation
**Current State:** Basic Gin binding validation  
**Recommended:** Enhanced validation with custom validators

### 3. Data Protection
**Current State:** Database credentials in configuration  
**Recommended:** Environment variables and secret management

## Scalability Architecture

### 1. Horizontal Scaling
**Current Limitations:**
- Single instance processing
- No load balancing
- Shared temporary file storage

**Recommended Improvements:**
- Kubernetes deployment
- Load balancer integration
- Distributed temporary storage

### 2. Vertical Scaling
**Current Optimizations:**
- Concurrent processing with goroutines
- Channel-based communication
- Configurable thread counts

**Recommended Improvements:**
- Connection pooling
- Result streaming
- Memory optimization

## Error Handling Architecture

### 1. Error Propagation
**Pattern:** Explicit error returns with context  
**Current Issues:** Excessive use of `Fatal()` calls  
**Recommended:** Graceful error handling with recovery

### 2. Error Categories
- **System Errors:** Database connection failures, tool installation
- **Business Errors:** Invalid domains, no targets in scope
- **External Errors:** Tool execution failures, network issues

### 3. Error Recovery
**Current State:** Limited recovery mechanisms  
**Recommended:** Circuit breaker pattern, retry logic, fallback strategies

## Testing Architecture

### 1. Unit Testing
**Current State:** Not implemented  
**Recommended Structure:**
```
pkg/
├── db8/
│   ├── db8_test.go
│   └── db8_domain8_test.go
├── controller8/
│   └── controller8_test.go
└── passive/
    └── passive_test.go
```

### 2. Integration Testing
**Recommended:** Test database setup, external tool mocking, API endpoint testing

### 3. Performance Testing
**Recommended:** Benchmark tests, load testing, stress testing

## Deployment Architecture

### 1. Container Strategy
**Current:** Basic Dockerfile  
**Recommended:** Multi-stage builds, security hardening, resource limits

### 2. Configuration Management
**Current:** YAML files  
**Recommended:** ConfigMaps, Secrets, environment-specific configurations

### 3. Monitoring & Observability
**Current:** Basic logging  
**Recommended:** Metrics collection, tracing, alerting

## Future Architecture Considerations

### 1. Microservices Evolution
- Service decomposition
- API gateway integration
- Service mesh implementation

### 2. Event-Driven Architecture
- Event sourcing
- CQRS pattern
- Async processing

### 3. Cloud-Native Features
- Auto-scaling
- Service discovery
- Health checking

## Design Patterns Used

### 1. Creational Patterns
- **Factory Pattern:** Database connection creation
- **Builder Pattern:** Configuration building with Viper

### 2. Structural Patterns
- **Adapter Pattern:** External tool integration
- **Facade Pattern:** API layer abstraction

### 3. Behavioral Patterns
- **Observer Pattern:** RabbitMQ message handling
- **Strategy Pattern:** Passive vs Active scanning
- **Template Method:** Scanning workflow

## Architecture Best Practices

### 1. Followed Practices
- Interface-based design
- Dependency injection
- Separation of concerns
- Concurrent processing

### 2. Areas for Improvement
- Error handling standardization
- Resource management
- Security integration
- Testing coverage

### 3. Recommended Additions
- Circuit breaker pattern
- Rate limiting
- Caching strategies
- Event sourcing

## Conclusion

The ASMM8 architecture demonstrates solid Go development practices with good separation of concerns and effective use of concurrency. The modular design allows for easy extension and maintenance. However, improvements in error handling, resource management, and security would enhance the overall robustness of the system.

The architecture is well-suited for the current requirements but would benefit from the recommended enhancements for production deployment and scalability.