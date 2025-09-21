# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-09-21

### Added
- High-load microservice built with Go
- HTTP API with Gin framework for user and event management
- PostgreSQL database integration with automatic migrations
- Redis caching layer
- Kafka message broker integration with producer/consumer
- Goroutine worker pool for parallel task processing
- Docker containerization with multi-stage build
- Docker Compose for local development
- Kubernetes deployment manifests
- Comprehensive health checks for all services
- GitHub Actions CI/CD pipeline
- End-to-end testing for Docker Compose and Kubernetes
- Local smoke testing scripts (PowerShell and Bash)
- Complete documentation and examples

### Features
- **User Management**: Create, read, update, delete users
- **Event Management**: Create and list events with user association
- **High Performance**: Concurrent processing with goroutines and channels
- **Scalability**: Kubernetes deployment with HPA (Horizontal Pod Autoscaler)
- **Monitoring**: Health endpoints for all integrated services
- **Testing**: Unit tests with race detection and coverage
- **CI/CD**: Automated testing and deployment pipelines

### Technical Stack
- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Message Broker**: Apache Kafka
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **CI/CD**: GitHub Actions

### API Endpoints
- `GET /health` - Health check
- `POST /api/v1/users/` - Create user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `GET /api/v1/users/` - List users
- `POST /api/v1/events/` - Create event
- `GET /api/v1/events/` - List events
- `GET /api/v1/events/:id` - Get event by ID

### Deployment
- **Local**: `docker compose up -d`
- **Kubernetes**: Apply manifests in `k8s/` directory
- **Docker Hub**: `docker.io/oleg2195/highload-microservice:latest`

### Testing
- **Unit Tests**: `go test -race -cover ./...`
- **Local Smoke**: `scripts/smoke.ps1` (Windows) or `scripts/smoke.sh` (Linux/Mac)
- **E2E Compose**: GitHub Actions workflow
- **E2E Kubernetes**: GitHub Actions workflow with Kind cluster
