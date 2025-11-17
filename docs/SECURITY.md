# ASMM8 Security Considerations

**Date:** July 18, 2025
**Last Updated:** November 7, 2025
**Risk Assessment:** Medium to High

## Overview

This document outlines security considerations, vulnerabilities, and recommended security improvements for the ASMM8 Asset Surface Management system.

## Current Security Issues

### 1. Credential Management (HIGH RISK)

**Issue:** Database and RabbitMQ credentials stored in plain text configuration files.

**Files Affected:**
- `configuration.yaml` (lines 58, 63)
- `pkg/db8/db8.go` (password handling)
- `pkg/orchestrator8/orchestrator8.go` (RabbitMQ credentials)

**Current Implementation:**
```yaml
# configuration.yaml
Database:
  username: "cpt_dbuser"
  password: "!!cpt!!"        # Plain text password
RabbitMQ:
  username: "deifzar"
  password: "deifzar85"      # Plain text password
```

**Risk:** Credentials exposure, unauthorized access to database and message queue.

### 2. External Tool Execution (MEDIUM RISK)

**Issue:** Direct execution of external tools without proper input validation or sanitization.

**Files Affected:**
- `pkg/passive/subfinder/subfinder.go` (92 lines)
- `pkg/active/dnsx/dnsx.go` (86 lines)
- `pkg/active/alterx/alterx.go` (100 lines)
- `pkg/passive/httpx/httpx.go` (92 lines)
- `pkg/active/httpx/httpx.go` (94 lines)

**External Tools Used:**
- **subfinder** v2.9.0 - Passive subdomain enumeration
- **dnsx** v1.2.2 - DNS resolution and brute-forcing
- **alterx** v0.0.6 - DNS alteration/permutation generation
- **httpx** - HTTP probing

**Current Implementation:**
```go
// pkg/passive/subfinder/subfinder.go
cmd := exec.Command("subfinder", "-d", seedDomain, "-silent", "-all",
    "-config", "./configs/subfinderconfig.yaml",
    "-pc", "./configs/subfinderprovider-config.yaml")
```

**Risk:** Command injection, arbitrary code execution if domain input is not properly sanitized before passing to external commands.

### 3. Input Validation (MEDIUM RISK)

**Issue:** Limited input validation for domain names and API parameters.

**Files Affected:**
- `pkg/controller8/controller8_domain8.go`
- `pkg/model8/domain8.go`
- API endpoints

**Current Implementation:**
```go
// pkg/model8/domain8.go:13
type PostDomain8 struct {
    Name        string `json:"name" binding:"required"` // No domain validation
    Companyname string `json:"companyname" binding:"required"`
    Enabled     bool   `json:"enabled" binding:"boolean"`
}
```

**Risk:** Injection attacks, malformed input processing.

### 4. SQL Injection Protection (LOW RISK)

**Issue:** While prepared statements are used, consistency could be improved.

**Files Affected:**
- `pkg/db8/db8_domain8.go`
- `pkg/db8/db8_hostname8.go`

**Current Status:** Generally well-protected with prepared statements, but some queries could be improved.

### 5. Authentication & Authorization (NOT IMPLEMENTED)

**Issue:** No authentication or authorization mechanisms in place.

**Files Affected:**
- `pkg/api8/api8.go`
- All API endpoints

**Risk:** Unauthorized access to scanning capabilities and data.

## Security Recommendations

### Immediate Actions (High Priority)

#### 1. Implement Environment-Based Credential Management
```bash
# Environment variables
export DB_PASSWORD="secure_password"
export RABBITMQ_PASSWORD="secure_password"
export JWT_SECRET="random_secret_key"
```

```go
// pkg/db8/db8.go - Updated implementation
func (d *Db8) InitDatabase8(l string, port int, sc, db, u string) {
    d.location = l
    d.port = port
    d.schema = sc
    d.database = db
    d.username = u
    d.password = os.Getenv("DB_PASSWORD") // Use environment variable
}
```

#### 2. Add Input Validation
```go
// pkg/model8/domain8.go - Enhanced validation
type PostDomain8 struct {
    Name        string `json:"name" binding:"required,hostname"`
    Companyname string `json:"companyname" binding:"required,min=1,max=100"`
    Enabled     bool   `json:"enabled" binding:"boolean"`
}

// Add custom validation function
func IsValidDomain(domain string) bool {
    // Implement RFC 1123 hostname validation
    // Check for malicious patterns
    return true
}
```

#### 3. Sanitize External Tool Input
```go
// pkg/passive/subfinder/subfinder.go - Enhanced security
import (
    "regexp"
    "strings"
)

func sanitizeDomain(domain string) (string, error) {
    // Remove potentially dangerous characters
    domainRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
    if !domainRegex.MatchString(domain) {
        return "", errors.New("invalid domain format")
    }
    
    // Additional checks for length, format, etc.
    if len(domain) > 253 {
        return "", errors.New("domain too long")
    }
    
    return domain, nil
}

func RunSubfinderIn(seedDomain string, results chan<- string, wg *sync.WaitGroup) {
    defer wg.Done()
    
    // Sanitize input
    cleanDomain, err := sanitizeDomain(seedDomain)
    if err != nil {
        log8.BaseLogger.Error().Msgf("Invalid domain: %s", err)
        close(results)
        return
    }
    
    // Use sanitized domain
    cmd := exec.Command("subfinder", "-d", cleanDomain, "-silent", "-all")
    // ... rest of implementation
}
```

### Medium Priority Actions

#### 1. Implement API Authentication
```go
// pkg/api8/middleware.go - New file
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        // Validate JWT token
        if !isValidToken(token) {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

#### 2. Add Rate Limiting
```go
// pkg/api8/middleware.go - Rate limiting
import "golang.org/x/time/rate"

func RateLimitMiddleware(requests int, duration time.Duration) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(duration), requests)
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

#### 3. Implement Request Logging
```go
// pkg/api8/middleware.go - Security logging
func SecurityLogMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Log request details
        log8.BaseLogger.Info().
            Str("method", c.Request.Method).
            Str("path", c.Request.URL.Path).
            Str("ip", c.ClientIP()).
            Str("user_agent", c.Request.UserAgent()).
            Msg("API request")
        
        c.Next()
        
        // Log response details
        log8.BaseLogger.Info().
            Int("status", c.Writer.Status()).
            Dur("duration", time.Since(start)).
            Msg("API response")
    }
}
```

### Long-term Security Improvements

#### 1. Secret Management System
- Implement HashiCorp Vault or AWS Secrets Manager
- Rotate credentials regularly
- Use encrypted configuration files

#### 2. Network Security
- Implement TLS/SSL for all communications
- Use VPN or private networks for internal services
- Implement firewall rules

#### 3. Container Security
```dockerfile
# Dockerfile - Security improvements
FROM golang:1.21-alpine AS builder

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# ... build process ...

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy binary and switch to non-root user
COPY --from=builder /app/asmm8 .
COPY --from=builder /etc/passwd /etc/passwd
USER appuser

CMD ["./asmm8"]
```

#### 4. Audit Logging
```go
// pkg/log8/audit.go - New file
type AuditEvent struct {
    Timestamp time.Time
    UserID    string
    Action    string
    Resource  string
    Success   bool
    Details   map[string]interface{}
}

func LogAuditEvent(event AuditEvent) {
    log8.BaseLogger.Info().
        Time("timestamp", event.Timestamp).
        Str("user_id", event.UserID).
        Str("action", event.Action).
        Str("resource", event.Resource).
        Bool("success", event.Success).
        Interface("details", event.Details).
        Msg("audit_event")
}
```

## Security Testing Recommendations

### 1. Static Analysis
```bash
# Install and run gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
gosec ./...
```

### 2. Dependency Scanning
```bash
# Install and run nancy
go install github.com/sonatypeoss/nancy@latest
nancy sleuth
```

### 3. Container Scanning
```bash
# Use tools like Trivy or Clair
trivy image asmm8:latest
```

## Compliance Considerations

### 1. Data Protection
- Implement data retention policies
- Add data encryption at rest
- Ensure GDPR compliance for EU data

### 2. Access Control
- Implement role-based access control (RBAC)
- Add multi-factor authentication (MFA)
- Regular access reviews

### 3. Monitoring & Alerting
- Security event monitoring
- Intrusion detection
- Automated incident response

## Security Incident Response

### 1. Incident Classification
- **Critical:** Unauthorized access, data breach
- **High:** Service disruption, credential compromise
- **Medium:** Failed authentication attempts, suspicious activity
- **Low:** Configuration issues, minor vulnerabilities

### 2. Response Procedures
1. **Detection:** Automated monitoring and alerting
2. **Analysis:** Incident investigation and classification
3. **Containment:** Isolate affected systems
4. **Recovery:** Restore services and data
5. **Documentation:** Post-incident review and improvements

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://github.com/OWASP/Go-SCP)
- [Container Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [Database Security Guidelines](https://www.postgresql.org/docs/current/security.html)