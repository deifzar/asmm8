# ASMM8 - Asset Surface Management Mate

<div align="center">

**Production-grade Go microservice for automated subdomain enumeration and reconnaissance.**

[![Go Version](https://img.shields.io/badge/Go-1.21.5+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?logo=docker)](dockerfile)
[![Status](https://img.shields.io/badge/status-pre--production-orange)](https://github.com/yourusername/asmm8)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

[Features](#features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Architecture](#architecture) • [API](#api-reference)

</div>

---

## Overview

ASMM8 (Asset Surface Management Mate) is a production-grade Go microservice designed for automated subdomain enumeration and asset reconnaissance. It provides a robust REST API for managing domains and orchestrating both passive and active subdomain discovery workflows.

**Built for:**
- Security researchers and penetration testers
- Red teams conducting reconnaissance
- Bug bounty hunters expanding attack surface
- Security operations teams managing asset inventories

### Key Features

- **Dual Enumeration Modes**: Passive (API-based) and active (DNS brute-forcing) subdomain discovery
- **Asynchronous Processing**: RabbitMQ-based message queuing with advanced connection pooling
- **Persistent Storage**: PostgreSQL integration with optimized batch operations
- **Production Ready**: Docker containerization, health checks, and graceful shutdown
- **Scalable Architecture**: Interface-based design with dependency injection
- **Comprehensive Tooling**: Integration with industry-standard tools (Subfinder, DNSx, Alterx, HTTPx)

---

## Quick Start

### Prerequisites

- **Go** 1.21.5 or higher
- **PostgreSQL** 12+ (for data persistence)
- **RabbitMQ** 3.8+ (for message queuing)
- **Docker** (optional, for containerized deployment)

### Installation

#### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/ASMM8.git
cd ASMM8

# Install Go dependencies
go mod download

# Install external security tools
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@v2.9.0
go install github.com/projectdiscovery/dnsx/cmd/dnsx@v1.2.2
go install github.com/projectdiscovery/alterx/cmd/alterx@v0.0.6

# Build the binary
go build -o asmm8 .

# Run the service
./asmm8 launch --ip 0.0.0.0 --port 8000
```

#### Option 2: Docker

```bash
# Build the Docker image
docker build -t asmm8:latest .

# Run the container
docker run -d \
  -p 8000:8000 \
  -e DATABASE_HOST=your-db-host \
  -e RABBITMQ_HOST=your-rabbitmq-host \
  --name asmm8 \
  asmm8:latest
```

### Configuration

1. Copy the example configuration:
```bash
cp configs/configuration.yaml.example configs/configuration.yaml
```

2. Edit `configs/configuration.yaml` with your settings:
```yaml
APP_ENV: PROD
LOG_LEVEL: "1"  # 0=debug, 1=info, 2=warn, 3=error

ASMM8:
  runType: complete  # "fast" (passive only) or "complete" (passive + active)
  activeWordList: wordlist/subdomainslite.txt
  activeThreads: 100

Database:
  location: localhost
  port: 5432
  database: cptm8
  username: cpt_dbuser
  password: ${DB_PASSWORD}  # Use environment variables for secrets

RabbitMQ:
  location: localhost
  port: 5672
  username: ${RABBITMQ_USER}
  password: ${RABBITMQ_PASS}
```

3. Set environment variables for sensitive data:
```bash
export DB_PASSWORD="your-secure-password"
export RABBITMQ_USER="your-rabbitmq-user"
export RABBITMQ_PASS="your-rabbitmq-password"
```

### Database Setup

```sql
-- Create database
CREATE DATABASE cptm8;

-- Create tables
CREATE TABLE cptm8domain (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    companyname VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cptm8hostname (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID REFERENCES cptm8domain(id),
    hostname VARCHAR(255) NOT NULL,
    live BOOLEAN DEFAULT false,
    enabled BOOLEAN DEFAULT true,
    foundfirsttime TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_hostname_domain ON cptm8hostname(domain_id);
CREATE INDEX idx_hostname_live ON cptm8hostname(live);
```

---

## Usage

### Basic Workflow

1. **Add a domain to track:**
```bash
curl -X POST http://localhost:8000/domain \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com",
    "companyname": "Example Corp",
    "enabled": true
  }'
```

2. **Run a complete scan (passive + active):**
```bash
curl "http://localhost:8000/scan?domains=example.com"
```

3. **Run passive enumeration only:**
```bash
curl "http://localhost:8000/scan/passive?domains=example.com"
```

4. **Check subdomain liveness (HTTP/HTTPS probing):**
```bash
curl "http://localhost:8000/scan/check?domains=example.com"
```

5. **Retrieve discovered hostnames:**
```bash
curl "http://localhost:8000/domain/{domain-uuid}/hostname"
```

### Scanning Modes

#### Fast Mode (Passive Only)
```yaml
ASMM8:
  runType: fast
```
- Uses external APIs (SecurityTrails, VirusTotal, etc.)
- No DNS brute-forcing
- Faster execution (typically 1-5 minutes)
- Lower resource usage

#### Complete Mode (Passive + Active)
```yaml
ASMM8:
  runType: complete
```
- Passive enumeration + DNS brute-forcing + permutations
- More comprehensive results
- Longer execution time (5-30 minutes depending on wordlist)
- Higher resource usage

---

## API Reference

### Domain Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/domain` | List all domains |
| `POST` | `/domain` | Create new domain |
| `GET` | `/domain/:id` | Get domain by ID |
| `PUT` | `/domain/:id` | Update domain |
| `DELETE` | `/domain/:id` | Delete domain |

### Hostname Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/domain/:id/hostname` | List hostnames for domain |
| `POST` | `/domain/:id/hostname` | Add hostname |
| `GET` | `/domain/:id/hostname/:hostnameid` | Get specific hostname |
| `PUT` | `/domain/:id/hostname/:hostnameid` | Update hostname |
| `DELETE` | `/domain/:id/hostname/:hostnameid` | Delete hostname |

### Scanning Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/scan?domains=d1,d2` | Full scan (passive + active) |
| `GET` | `/scan/passive?domains=d1` | Passive enumeration only |
| `GET` | `/scan/active?domains=d1` | Active enumeration only |
| `GET` | `/scan/check?domains=d1` | HTTP/HTTPS liveness check |

### Health Check Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Liveness probe (always 200 OK) |
| `GET` | `/ready` | Readiness probe (checks DB + RabbitMQ) |

**Example Response:**
```json
{
  "status": "ready",
  "database": "connected",
  "rabbitmq": "connected",
  "timestamp": "2025-11-16T12:34:56Z"
}
```

---

## Architecture

### High-Level Overview

```
┌─────────────┐
│   Client    │───┐
│   (HTTP)    │   │
└─────────────┘   │
                  │    ┌─────────────┐    ┌─────────────┐
┌─────────────┐   ├───▶│  ASMM8 API  │───▶│  Database   │
│  RabbitMQ   │   │    │  (Gin/Go)   │    │ (PostgreSQL)│
│  (Message   │───┘    └─────────────┘    └─────────────┘
│   Queue)    │                 │
└─────────────┘                 │
                 ┌──────────────┼──────────────┐
                 ▼              ▼              ▼
         ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
         │  Passive    │ │   Active    │ │   Liveness  │
         │ Enumeration │ │ Enumeration │ │   Check     │
         │ (Subfinder) │ │(DNSx/Alterx)│ │   (HTTPx)   │
         └─────────────┘ └─────────────┘ └─────────────┘
```

### Scanning Workflow

**Complete Scan Pipeline:**

```
1. PASSIVE ENUMERATION
   ├─→ Subfinder (API-based discovery)
   └─→ Result aggregation → Deduplication → Database storage

2. ACTIVE ENUMERATION
   ├─→ DNSx Brute Force (wordlist-based)
   ├─→ Alterx Permutations (subdomain variations)
   ├─→ DNSx Resolution (verify permutations)
   └─→ Deduplication → Database storage

3. LIVENESS VERIFICATION
   └─→ HTTPx Probing (HTTP/HTTPS detection)
       └─→ Update live status in database

4. ORCHESTRATION
   └─→ Publish results to downstream services via RabbitMQ
```

### Package Structure

```
asmm8/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Base command setup
│   ├── launch.go          # API service launcher
│   └── version.go         # Version information
├── pkg/                    # 13 packages, 42 Go files
|   ├── active/            # Active scan utilities
│   ├── amqpM8/            # RabbitMQ connection pooling (5 files)
│   ├── api8/              # HTTP API routes and initialization
│   ├── cleanup8/          # Temporary file cleanup utilities
│   ├── configparser/      # Configuration management (Viper)
│   ├── controller8/       # Business logic controllers
│   ├── db8/               # Database access layer (6 modules)
│   ├── log8/              # Structured logging (zerolog)
│   ├── model8/            # Data models and domain entities (13 files)
│   ├── notification8/     # Discord notifications
│   ├── orchestrator8/     # Service orchestration
│   ├── passive/           # Passive scan utilities
│   └── utils/             # Utility functions
├── configs/               # Configuration files
├── docs/                  # Comprehensive documentation
└── main.go                # Application entry point
```

### Key Components

- **[API Layer](pkg/api8/)** - Gin-based REST API (132 lines)
- **[Controllers](pkg/controller8/)** - Business logic (775 lines)
- **[Database Layer](pkg/db8/)** - PostgreSQL repository pattern (730 lines)
- **[Message Queue](pkg/amqpM8/)** - RabbitMQ with connection pooling (1,777 lines)
- **[Orchestrator](pkg/orchestrator8/)** - Service coordination (259 lines)
- **[Passive Scanner](pkg/passive/)** - API-based enumeration (177 lines)
- **[Active Scanner](pkg/active/)** - DNS brute-forcing (246 lines)

For detailed architecture documentation, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## Advanced Features

### RabbitMQ Integration

**Advanced Connection Pooling:**
- Configurable pool size (default: 2-10 connections)
- Automatic connection recovery on failures
- Periodic health checks (30-minute intervals)
- Manual message acknowledgment with smart ACK/NACK logic
- Delivery tag tracking for message lifecycle management

**Message Flow:**
```
RabbitMQ → Consumer → Extract deliveryTag → HTTP Request
    ↓
LaunchScan() → Extract from header → Active()
    ↓
Defer ACK/NACK → Success: ACK | Failure: NACK+requeue
```

### Error Handling

**Resilient Design:**
- Panic recovery with defer blocks
- Automatic requeue on scan failures
- Error notifications via RabbitMQ
- Graceful degradation on tool failures

### Concurrent Processing

- Goroutines for parallel subdomain discovery
- Channel-based communication between tools
- Configurable thread counts (default: 100 DNS threads)
- Optimized batch database insertions

---

## Documentation

Comprehensive documentation is available in the [docs/](docs/) directory:

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Detailed system architecture (14 KB) |
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Development setup and guidelines (6 KB) |
| [SECURITY.md](docs/SECURITY.md) | Security best practices (9 KB) |
| [PERFORMANCE.md](docs/PERFORMANCE.md) | Performance optimization (15 KB) |
| [TODO.md](docs/TODO.md) | Known issues and roadmap (16 KB) |
| [CODE_REVIEW.md](docs/CODE_REVIEW.md) | Code review checklist (6 KB) |

---

## External Tools

ASMM8 integrates with industry-standard reconnaissance tools:

| Tool | Version | Purpose | Size |
|------|---------|---------|------|
| [Subfinder](https://github.com/projectdiscovery/subfinder) | v2.9.0 | Passive subdomain discovery | 41 MB |
| [DNSx](https://github.com/projectdiscovery/dnsx) | v1.2.2 | DNS resolution and brute-forcing | 36 MB |
| [Alterx](https://github.com/projectdiscovery/alterx) | v0.0.6 | Subdomain permutation generation | 27 MB |
| [HTTPx](https://github.com/projectdiscovery/httpx) | Latest | HTTP/HTTPS probing | - |

All tools are automatically installed during the Docker build process.

---

## Performance

**Typical Performance Metrics:**

- **Passive Scan**: 1-5 minutes for 1,000-10,000 subdomains
- **Active Scan**: 5-30 minutes (depends on wordlist size)
- **Database Operations**: Batch insertions for optimal performance
- **Concurrent Processing**: Up to 100 parallel DNS resolutions

**Resource Requirements:**

- **CPU**: 2+ cores recommended
- **Memory**: 2 GB minimum, 4 GB recommended
- **Storage**: 10 GB for application + logs + temporary files
- **Network**: Stable internet connection for API-based discovery

For optimization tips, see [docs/PERFORMANCE.md](docs/PERFORMANCE.md).

---

## Security Considerations

### Current Limitations

- No authentication on API endpoints (planned for v2.0)
- Database credentials in configuration file (use environment variables)
- Limited input validation (basic Gin binding only)

### Recommendations

1. **Use environment variables** for all secrets
2. **Deploy behind API gateway** with authentication
3. **Enable TLS/SSL** for production deployments
4. **Implement rate limiting** to prevent abuse
5. **Run as non-root user** in containers (already configured)

See [docs/SECURITY.md](docs/SECURITY.md) for comprehensive security guidelines.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow Go best practices and existing code style
4. Add tests for new functionality
5. Update documentation as needed
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development guidelines.

---

## Roadmap

### Version 1.x (Current)
- [x] Core subdomain enumeration functionality
- [x] PostgreSQL persistence
- [x] RabbitMQ message queuing
- [x] Docker containerization
- [x] Health check endpoints

### Version 2.0 (Planned)
- [ ] JWT-based authentication
- [ ] Rate limiting and request throttling
- [ ] Unit and integration tests (target: 80% coverage)
- [ ] Kubernetes deployment manifests
- [ ] Prometheus metrics integration
- [ ] GraphQL API option
- [ ] Web dashboard for visualization

See [docs/TODO.md](docs/TODO.md) for the complete roadmap and known issues.

---

## Troubleshooting

### Common Issues

**1. Database connection failures**
```bash
# Check PostgreSQL is running
systemctl status postgresql

# Verify connection settings in configs/configuration.yaml
# Ensure database and tables are created
```

**2. RabbitMQ connection errors**
```bash
# Check RabbitMQ status
systemctl status rabbitmq-server

# Verify credentials and port in configuration
# Check exchange and queue creation
```

**3. External tools not found**
```bash
# Ensure tools are in PATH
which subfinder dnsx alterx httpx

# Reinstall if needed
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@v2.9.0
```

**4. Permission errors**
```bash
# Ensure log directory is writable
chmod 755 log/ app/log/

# Check file permissions for config files
chmod 640 configs/configuration.yaml
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [ProjectDiscovery](https://github.com/projectdiscovery) for excellent security tools
- [Gin Web Framework](https://github.com/gin-gonic/gin) for the HTTP router
- [RabbitMQ](https://www.rabbitmq.com/) for reliable message queuing
- [PostgreSQL](https://www.postgresql.org/) for robust data persistence

---

## Contact

For questions, issues, or feature requests, please open an issue on GitHub.

**Project Link:** [https://github.com/yourusername/ASMM8](https://github.com/yourusername/ASMM8)

---

<div align="center">

**Built with ❤️ for the security community**

[⬆ Back to Top](#asmm8---asset-surface-management-mate)

</div>
