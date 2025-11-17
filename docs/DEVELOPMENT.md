# Development Guide

## Prerequisites

### System Requirements
- **Go 1.21.5+** installed (as specified in go.mod)
- **PostgreSQL 12+** database server
- **RabbitMQ 3.x+** message broker
- **External tools:** subfinder v2.9.0, alterx v0.0.6, dnsx v1.2.2, httpx (optional: amass)

### Database Setup
1. Install PostgreSQL and create database:
```sql
CREATE DATABASE cptm8;
CREATE USER cpt_dbuser WITH PASSWORD '!!cpt!!';
GRANT ALL PRIVILEGES ON DATABASE cptm8 TO cpt_dbuser;
```

2. Create required tables:
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
    live BOOLEAN DEFAULT false,
    enabled BOOLEAN DEFAULT true,
    foundfirsttime TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- General scan settings table
CREATE TABLE general_scan_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(255) NOT NULL,
    setting_value TEXT
);
```

### Message Queue Setup
1. Install RabbitMQ server
2. Create user and configure access:
```bash
sudo rabbitmqctl add_user deifzar deifzar85
sudo rabbitmqctl set_user_tags deifzar administrator
sudo rabbitmqctl set_permissions -p / deifzar ".*" ".*" ".*"
```

### External Tools Installation
Install required security tools via Go:
```bash
# Subfinder - Passive subdomain enumeration
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@v2.9.0

# DNSx - DNS resolution and brute-forcing
go install github.com/projectdiscovery/dnsx/cmd/dnsx@v1.2.2

# Alterx - DNS alteration and permutation
go install github.com/projectdiscovery/alterx/cmd/alterx@v0.0.6

# HTTPx - HTTP probing (optional, latest version)
go install github.com/projectdiscovery/httpx/cmd/httpx@latest
```

Verify installations:
```bash
subfinder -version
dnsx -version
alterx -version
httpx -version
```

## Development Environment

### Environment Configuration
Set environment variables in `configuration.yaml`:
```yaml
APP_ENV: DEV    # DEV, TEST, PROD
LOG_LEVEL: "0"  # 0=debug, 1=info, 2=warn, 3=error, 4=fatal, 5=panic
```

### Development Commands

#### Build and Run
```bash
# Build binary
go build -o asmm8 .

# Run from source
go run main.go launch --ip 0.0.0.0 --port 8000

# Run with custom configuration
go run main.go launch --ip 127.0.0.1 --port 8001
```

#### Dependencies
```bash
# Install/update dependencies
go mod tidy
go mod download

# Vendor dependencies (optional)
go mod vendor
```

#### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/controller8/
```

## Development Workflow

### 1. Code Organization
- **`pkg/api8/`** - HTTP API routes and middleware
- **`pkg/controller8/`** - Business logic controllers
- **`pkg/db8/`** - Database interfaces and implementations
- **`pkg/model8/`** - Data structures and models
- **`pkg/amqpM8/`** - RabbitMQ client implementation
- **`pkg/orchestrator8/`** - Message handling orchestration
- **`pkg/passive/`** - Passive enumeration tools (Amass, Subfinder)
- **`pkg/active/`** - Active enumeration tools (Alterx, DNSx, HTTPx)

### 2. Configuration Management
- Main config: `configuration.yaml`
- Uses Viper for configuration loading
- Environment-specific overrides supported
- Database and RabbitMQ connection strings configurable

### 3. Database Development
- Uses standard `database/sql` with `lib/pq` driver (PostgreSQL)
- Repository pattern with interfaces in `pkg/db8/`
- Connection retry logic: 10 attempts with 5-second intervals
- Batch operations for efficient bulk insertions
- **Important:** Always use `defer rows.Close()` after queries to prevent resource leaks

### 4. Message Queue Development
- Advanced connection pooling with auto-recovery (2-10 connections)
- Topic-based routing with RabbitMQ
- Exchanges: `cptm8` (processing), `notification` (alerts)
- Queue: `qasmm8` (max 1 message, reject-publish overflow)
- Routing pattern: `cptm8.asmm8.#`
- Health checks every 30 minutes (configurable)
- Thread-safe operations with automatic reconnection
- **Manual acknowledgment mode** - Set `autoack: "false"` in configuration
- Delivery tag tracking from consumer to controller
- Smart ACK/NACK logic based on scan completion status

#### Manual Acknowledgment Configuration

The application uses **manual acknowledgment mode** for reliable message processing:

**Configuration:** `configs/configuration.yaml`
```yaml
ORCHESTRATORM8:
  asmm8:
    Consumer:
      - "qasmm8"  # queue name
      - "casmm8"  # consumer name prefix
      - "false"   # autoack - IMPORTANT: Set to "false" for manual ACK
```

**How It Works:**
1. RabbitMQ consumer receives message with unique deliveryTag
2. Consumer passes deliveryTag via HTTP header to API endpoint
3. API controller extracts deliveryTag and passes to scan function
4. Defer block in scan function ACKs or NACKs based on completion:
   - **ACK**: Scan completed (successfully or with handled errors)
   - **NACK + requeue**: Scan crashed, panicked, or was interrupted
   - **NACK no requeue**: Handler error or permanent failure

**Benefits:**
- Messages are requeued automatically if container crashes
- No message loss during Kubernetes pod restarts
- Failed scans can be retried automatically
- Completed scans (even with warnings) are acknowledged and removed

**Debugging Manual ACK:**
```bash
# Check RabbitMQ queue status
rabbitmqadmin list queues name messages messages_ready messages_unacknowledged

# Monitor deliveryTag in logs
grep "deliveryTag" log/asmm8.log

# Check ACK/NACK operations
grep -E "(ACK|NACK)" log/asmm8.log
```

### 5. API Development
- Gin web framework
- RESTful endpoints for domain management
- JSON request/response format
- Middleware for logging and error handling

## Debugging

### Logging
- Uses zerolog for structured logging with log rotation
- Log file: `log/asmm8.log` (with lumberjack rotation, 0640 permissions)
- Log levels: trace (-1), debug (0), info (1), warn (2), error (3), fatal (4), panic (5)
- Console + file output (configurable via APP_ENV)
- Singleton logger initialized with sync.Once

### Common Debug Commands
```bash
# Enable debug logging
export LOG_LEVEL=0

# Run with verbose output
go run main.go launch --ip 0.0.0.0 --port 8000 2>&1 | tee debug.log

# Check database connectivity
psql -h localhost -U cpt_dbuser -d cptm8 -c "SELECT 1;"

# Check RabbitMQ connectivity
rabbitmq-diagnostics ping
```

## Code Style and Standards

### Go Best Practices
- Follow Go conventions (gofmt, golint)
- Use interfaces for testability
- Implement proper error handling
- Use context for cancellation
- Implement graceful shutdown

### Project Conventions
- Package names end with `8` (e.g., `controller8`, `db8`)
- Use UUID v5 for unique identifiers
- Configuration keys in UPPERCASE
- Database table names in lowercase
- RabbitMQ routing keys use dot notation

## Testing Strategy

⚠️ **Current Status:** No unit or integration tests are currently implemented.

### Recommended Unit Tests
- Test individual functions and methods
- Mock external dependencies using interfaces
- Use table-driven tests where applicable
- Focus on: database operations, controllers, scanning engines

### Recommended Integration Tests
- Test database operations with test database
- Test RabbitMQ message handling with test broker
- Test API endpoints with httptest
- Mock external tools (subfinder, dnsx, alterx)

### Recommended End-to-End Tests
- Test complete scan workflows
- Test service integration with naabum8
- Test error scenarios and recovery
- Test resource cleanup and connection management

### Example Test Structure
```bash
# Create test files
touch pkg/db8/db8_domain8_test.go
touch pkg/controller8/controller8_asmm8_test.go
touch pkg/passive/passive_test.go

# Run tests
go test ./pkg/db8/... -v
go test ./pkg/controller8/... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Performance Considerations

### Scanning Performance
- Configure `activeThreads` in `configuration.yaml`
- Use `runType: fast` for passive enumeration only
- Use `runType: complete` for full active enumeration

### Database Performance
- Use batch operations for bulk inserts
- Implement connection pooling
- Monitor query performance

### Message Queue Performance
- Configure prefetch count for consumers
- Use durable queues for reliability
- Monitor queue depth and processing rates

## Security Considerations

### Database Security
- Use prepared statements to prevent SQL injection
- Implement proper connection string security
- Use least privilege database accounts

### API Security
- Input validation on all endpoints
- Rate limiting for API calls
- Secure configuration management

### Network Security
- Bind to specific interfaces when needed
- Use secure communication channels
- Implement proper authentication

## Deployment

### Development Deployment
```bash
# Start services
docker-compose up -d postgres rabbitmq

# Run application
go run main.go launch
```

### Production Deployment
- Build static binary with `CGO_ENABLED=0`
- Use systemd service files
- Configure reverse proxy (nginx/traefik)
- Set appropriate resource limits
- Configure monitoring and alerting

## Troubleshooting

### Common Issues
1. **Database connection failures**: Check PostgreSQL service and credentials
2. **RabbitMQ connection issues**: Verify RabbitMQ service and user permissions
3. **Port conflicts**: Ensure configured ports are available
4. **Missing external tools**: Install required enumeration tools

### Debug Checklist
- [ ] Configuration file syntax valid
- [ ] Database accessible and schema created
- [ ] RabbitMQ accessible and exchanges/queues created
- [ ] External tools installed and in PATH
- [ ] Ports not blocked by firewall
- [ ] Log files for error messages