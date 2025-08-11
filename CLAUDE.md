# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based monitoring relay agent designed to run as a deployment in the Rancher local cluster. Each deployment instance acts as a relay for a specific remote cluster, using Rancher's service proxy functionality to access Prometheus and Loki server pods in that remote cluster. The relay makes monitoring data from remote clusters accessible to centralized monitoring infrastructure.

## Architecture

- **Entry point**: `main.go` - handles startup, configuration validation, and Rancher API connectivity test
- **Configuration**: `pkg/config/` - environment variable management with type conversion utilities
- **Logging**: `pkg/logging/` - structured logging with logrus, configurable debug mode and caller information
- **Models**: `pkg/models/` - data structures including Status model with GORM tags for database operations

## Development Commands

### Build and Run
```bash
# Build the application
go build -o rancher-monitoring main.go

# Run directly
go run main.go

# Run with dependencies
go mod tidy && go run main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./pkg/config
```

### Dependencies
```bash
# Download dependencies
go mod download

# Clean up dependencies
go mod tidy

# View dependency graph
go mod graph
```

## Required Environment Variables

The application requires these environment variables to function:
- `RANCHER_API_ENDPOINT` - Rancher API base URL
- `RANCHER_API_ACCESS_KEY` - API access key for authentication
- `RANCHER_API_SECRET_KEY` - API secret key for authentication  
- `CLUSTER_ID` - Target cluster identifier

Optional variables:
- `DEBUG=true` - Enable debug logging and detailed caller information
- `METRICS_PORT` - Port for metrics endpoint (default: 9000)
- `CLUSTER_NAME` - Human-readable cluster name

## Remote Endpoint Configuration

### Prometheus Configuration (with defaults)
- `PROMETHEUS_NAMESPACE` - Namespace for Prometheus service (default: `cattle-monitoring-system`)
- `PROMETHEUS_SERVICE` - Prometheus service name (default: `rancher-monitoring-prometheus`)
- `PROMETHEUS_PORT` - Prometheus service port (default: `9090`)

### Loki Configuration (with defaults)
- `LOKI_NAMESPACE` - Namespace for Loki service (default: `cattle-logging-system`)
- `LOKI_SERVICE` - Loki service name (default: `rancher-logging-loki`)
- `LOKI_PORT` - Loki service port (default: `3100`)

### Generic Remote Endpoint Configuration
- `REMOTE_NAMESPACE` - Custom namespace for other services
- `REMOTE_SERVICE` - Custom service name
- `REMOTE_PORT` - Custom service port

## Key Dependencies

- `github.com/sirupsen/logrus` - Structured logging
- `github.com/rancher/*` - Rancher API client libraries
- GORM tags in models suggest database persistence (though database setup not yet implemented in main.go)

## Deployment Architecture

- Runs as a Kubernetes deployment in the Rancher local cluster
- One deployment instance per remote cluster to be monitored
- Each instance is configured with specific `CLUSTER_ID` for its target remote cluster
- Acts as a relay/proxy between centralized monitoring and remote cluster monitoring services

## Development Notes

- Uses Rancher's service proxy to reach Prometheus and Loki pods in remote clusters without requiring direct cluster access
- Designed for deployment in Rancher local cluster with each instance handling one remote cluster
- The Status model in `pkg/models/status.go` appears to be for backup/database monitoring rather than general Rancher monitoring, suggesting this may be a template or evolving codebase
- Debug mode affects both log level and whether filename/line information is included in logs
- HTTP client has 10-second timeout for Rancher API calls
- Service proxy URLs typically follow pattern: `{RANCHER_API_ENDPOINT}/k8s/clusters/{CLUSTER_ID}/api/v1/namespaces/{namespace}/services/{service}:{port}/proxy/`