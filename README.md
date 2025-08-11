# Rancher Centralized Monitoring Relay

[![CI/CD](https://github.com/supporttools/rancher-centralized-monitoring/workflows/CI%2FCD%20-%20v0.3/badge.svg)](https://github.com/supporttools/rancher-centralized-monitoring/actions)
[![Docker Image](https://img.shields.io/docker/pulls/supporttools/rancher-monitoring-relay.svg)](https://hub.docker.com/r/supporttools/rancher-monitoring-relay)
[![Helm Chart](https://img.shields.io/badge/helm-chart-blue.svg)](https://charts.support.tools/)

A Go-based monitoring relay agent that uses Rancher's service proxy functionality to access Prometheus and Loki server pods in remote clusters. The relay makes monitoring data from remote clusters accessible to centralized monitoring infrastructure.

## 🎯 Overview

The Rancher Centralized Monitoring Relay is designed to run as a deployment in the Rancher local cluster. Each deployment instance acts as a relay for a specific remote cluster, providing a bridge between centralized monitoring systems and remote cluster monitoring services that would otherwise be inaccessible from outside the cluster.

### Key Features

- **🔗 Service Proxy Integration**: Uses Rancher's built-in service proxy to reach remote monitoring services
- **📊 Multi-Service Support**: Configurable for Prometheus, Loki, and custom remote services
- **🔒 Secure Authentication**: Uses Rancher API keys for secure access
- **📈 Health Monitoring**: Built-in health checks and metrics endpoints
- **⚙️ Flexible Configuration**: Environment variable based configuration
- **🚀 Production Ready**: Includes Helm charts, CI/CD pipeline, and comprehensive monitoring

### Architecture

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│  Centralized        │    │   Rancher Local     │    │   Remote Cluster    │
│  Monitoring         │────│   Cluster           │────│   c-xxxxx           │
│  (Prometheus/Grafana)│    │                     │    │                     │
│                     │    │ ┌─────────────────┐ │    │ ┌─────────────────┐ │
│                     │    │ │ Monitoring      │ │    │ │ Prometheus      │ │
│                     │    │ │ Relay Agent     │ │    │ │ Loki            │ │
│                     │    │ │                 │ │    │ │ Custom Services │ │
└─────────────────────┘    │ └─────────────────┘ │    │ └─────────────────┘ │
                           └─────────────────────┘    └─────────────────────┘
```

## 🚀 Quick Start

### Using Helm (Recommended)

```bash
# Add the Support Tools Helm repository
helm repo add supporttools https://charts.support.tools/
helm repo update

# Install the monitoring relay
helm install my-relay supporttools/rancher-monitoring-relay \
  --set rancher.apiEndpoint="https://your-rancher-server" \
  --set rancher.clusterId="c-xxxxx" \
  --set rancher.auth.accessKey="token-xxxxx" \
  --set rancher.auth.secretKey="your-secret-key"
```

### Using Docker

```bash
docker run -d \
  -e RANCHER_API_ENDPOINT="https://your-rancher-server" \
  -e RANCHER_API_ACCESS_KEY="token-xxxxx" \
  -e RANCHER_API_SECRET_KEY="your-secret-key" \
  -e CLUSTER_ID="c-xxxxx" \
  -p 9000:9000 \
  supporttools/rancher-monitoring-relay:latest
```

## 📋 Prerequisites

- Rancher server with remote clusters
- Valid Rancher API credentials with cluster access permissions
- Kubernetes cluster for deployment (Rancher local cluster recommended)
- Prometheus/Loki or other monitoring services running in remote clusters

## 📚 Documentation

- **[Installation Guide](docs/installation.md)** - Detailed installation instructions
- **[Configuration Guide](docs/configuration.md)** - Complete configuration reference
- **[Usage Examples](docs/usage.md)** - Real-world usage scenarios
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

## 🔧 Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `RANCHER_API_ENDPOINT` | Rancher server URL | `https://rancher.example.com` |
| `RANCHER_API_ACCESS_KEY` | API access key | `token-xxxxx` |
| `RANCHER_API_SECRET_KEY` | API secret key | `your-secret-key` |
| `CLUSTER_ID` | Target cluster ID | `c-m-xxxxxxx` |

### Service Configuration

```bash
# Prometheus (defaults shown)
PROMETHEUS_NAMESPACE=cattle-monitoring-system
PROMETHEUS_SERVICE=rancher-monitoring-prometheus
PROMETHEUS_PORT=9090

# Loki (defaults shown)
LOKI_NAMESPACE=cattle-logging-system
LOKI_SERVICE=loki
LOKI_PORT=3100

# Custom service
REMOTE_NAMESPACE=monitoring
REMOTE_SERVICE=custom-service
REMOTE_PORT=8080
```

## 🏥 Health Checks

The relay provides several HTTP endpoints for monitoring:

- `GET /health` - Basic Rancher API connectivity check
- `GET /ready` - Comprehensive service connectivity check
- `GET /version` - Version and build information
- `GET /metrics` - Prometheus metrics

## 🔨 Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/supporttools/rancher-centralized-monitoring
cd rancher-centralized-monitoring

# Install dependencies
go mod download

# Build the application
go build -o rancher-monitoring-relay main.go

# Run tests
make test

# Build Docker image
make build TAG=latest
```

### Local Development

```bash
# Install development tools
make install-tools

# Run full CI pipeline locally
make ci

# Start development server
export RANCHER_API_ENDPOINT="https://your-rancher"
export RANCHER_API_ACCESS_KEY="token-xxxxx"
export RANCHER_API_SECRET_KEY="your-secret"
export CLUSTER_ID="c-xxxxx"
export DEBUG=true

go run main.go
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [https://docs.support.tools](https://docs.support.tools)
- **Issues**: [GitHub Issues](https://github.com/supporttools/rancher-centralized-monitoring/issues)

---

Made with ❤️ by [Support Tools](https://support.tools)
