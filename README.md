# Item Sync Service

A Go service for synchronizing external API data with automatic background jobs, retry logic, and idempotent operations.

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ for local development

### Using Docker (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd item-sync

# Copy environment file
cp .env.example .env

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app
```

The service will be available at http://localhost:8080

### Local Development

```bash
# Install dependencies
go mod download

# Set up environment
cp .env.example .env

# Start MySQL and Redis
docker-compose up -d mysql redis

# Start the service
go run main.go
```

## API Endpoints

### Health Check
```bash
GET /health
```

### Sync Items
```bash
POST /sync
{
  "force_sync": false,
  "api_source": "pokemon",
  "operation": "list",
  "params": {
    "limit": 20,
    "offset": 0
  }
}
```

### List Items
```bash
GET /items?limit=20&offset=0&api_source=pokemon
```

### Get Item Detail
```bash
GET /items/:id
```

## Background Jobs

The service automatically runs sync jobs every 15 minutes:

- Pokemon Sync: Fetches all Pokemon data with pagination
- OpenWeather Sync: Fetches weather data for predefined cities  
- Job Tracking: All executions logged with metrics and error handling

### Monitoring Jobs

```bash
# View worker logs
docker-compose logs -f app | grep -E "(worker|sync|job)"

# Check job history in database
SELECT * FROM sync_jobs ORDER BY started_at DESC LIMIT 10;
```

## Configuration

### Environment Variables

Key configuration options (see `.env.example` for complete list):

```bash
# API Selection
API_API_TYPE=pokemon              # or "openweather"

# Worker Configuration
WORKER_ENABLED=true               # Enable background jobs
WORKER_SYNC_INTERVAL=15m          # Sync every 15 minutes
WORKER_JOB_TIMEOUT=10m            # Job timeout

# Retry and Circuit Breaker
RETRY_MAX_RETRIES=5               # Max retry attempts
RETRY_BACKOFF_FACTOR=2.0          # Exponential backoff
RETRY_CIRCUIT_THRESHOLD=5         # Circuit breaker threshold
RETRY_CIRCUIT_TIMEOUT=60s         # Circuit breaker timeout

# Database
DATABASE_HOST=localhost
DATABASE_PORT=3306
DATABASE_USER=root
DATABASE_DATABASE=item_sync

# Cache
REDIS_HOST=localhost
REDIS_PORT=6379
```

### Supported API Types

#### Pokemon API
- URL: https://pokeapi.co/api/v2
- Features: Full pagination support, ID extraction from URLs

#### OpenWeather API
- URL: https://api.openweathermap.org/data/2.5
- Auth: API key required (configured in code)
- Features: City-based weather data

## Development

### Project Structure
```
├── internal/
│   ├── infrastructure/   # Server, database, worker setup
│   ├── item/             # Core business logic
│   │   ├── entity/       # Data models
│   │   ├── usecase/      # Business logic
│   │   ├── handler/      # HTTP handlers
│   │   ├── repository/   # Data access
│   │   ├── jobs/         # Background jobs
│   │   └── strategy/     # Sync strategies
│   └── errors/           # Custom error types
├── pkg/
│   ├── api/              # External API clients
│   ├── worker/           # Job scheduler
│   ├── retry/            # Retry logic
│   ├── circuit/          # Circuit breaker
│   └── logger/           # Logging utilities
├── migrations/           # Database migrations
└── config/               # Configuration management
```

### Running Tests
```bash
go test ./...
```

## Assumptions Made

### Data Consistency
- External APIs provide consistent data structures
- Pokemon API pagination is reliable and complete
- Content changes can be detected via hash comparison

### Business Requirements
- 15-minute sync interval is sufficient for data freshness
- Idempotent operations prevent duplicate processing
- Failed items don't block successful ones
- Job history retention is handled externally

### API Behavior
- Pokemon API is stable and doesn't require authentication
- OpenWeather API requires key management in application code
- Rate limiting is handled by retry logic and circuit breakers

### Scale Assumptions
- Single-instance deployment initially
- MySQL can handle the data volume and concurrent access
- Redis cache fits in memory
- Background jobs complete within timeout windows

## Trade-offs and Improvements

### With More Time, I Would Add:

#### Production Readiness
- Distributed Locking: Prevent multiple instances from running the same job
- Metrics and Monitoring: Prometheus metrics, health check endpoints
- Graceful Shutdown: Proper signal handling for job completion
- Configuration Validation: Startup-time config validation

#### Performance and Scale
- Connection Pooling: Optimized database connection management  
- Batch Processing: Process multiple items per transaction
- Async Processing: Message queues for high-throughput scenarios
- Horizontal Scaling: Kubernetes-ready deployment with multiple replicas

#### Reliability and Observability
- Structured Logging: JSON logs with correlation IDs
- Distributed Tracing: Request tracing across services
- Alerting: Failed job notifications and error rate monitoring
- Circuit Breaker Metrics: API health monitoring and automatic recovery

#### Security and Compliance
- API Key Management: Secure credential storage (Vault, K8s secrets)
- Rate Limiting: API-specific rate limiting and backpressure
- Input Validation: Comprehensive request validation and sanitization
- Audit Logging: Compliance-ready audit trails

#### Testing and Quality
- Integration Tests: End-to-end API and database testing
- Load Testing: Performance benchmarks and capacity planning
- Chaos Engineering: Resilience testing with failure injection
- Test Coverage: Comprehensive unit and integration test coverage

#### Advanced Features
- Multi-tenancy: Support multiple API configurations per tenant
- Real-time Updates: WebSocket or SSE for live data updates
- Data Transformation: Pluggable data transformation pipelines
- API Versioning: Support for multiple API versions and migration

#### Operational Excellence
- Blue-Green Deployment: Zero-downtime deployments
- Rollback Capability: Quick rollback on deployment issues
- Feature Flags: Runtime configuration changes without deployment
- Disaster Recovery: Backup strategies and recovery procedures

### Current Trade-offs Made:

1. Single Instance: No distributed coordination (acceptable for initial deployment)
2. In-Memory Circuit Breaker: State lost on restart (vs. Redis-backed state)
3. Hardcoded API Config: External APIs configured in code (vs. dynamic configuration)
4. Manual Scaling: Requires configuration changes (vs. auto-scaling)