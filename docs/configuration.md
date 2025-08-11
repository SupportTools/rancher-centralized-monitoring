# Configuration Guide

This guide covers all configuration options for the Rancher Centralized Monitoring Relay.

## Environment Variables

### Core Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `RANCHER_API_ENDPOINT` | ✅ | - | Rancher server API endpoint URL |
| `RANCHER_API_ACCESS_KEY` | ✅ | - | Rancher API access key (token-xxxxx) |
| `RANCHER_API_SECRET_KEY` | ✅ | - | Rancher API secret key |
| `CLUSTER_ID` | ✅ | - | Target remote cluster ID (c-xxxxxxx) |
| `CLUSTER_NAME` | ❌ | "" | Human-readable cluster name for logging |
| `DEBUG` | ❌ | false | Enable debug logging |
| `METRICS_PORT` | ❌ | 9000 | HTTP server port for metrics/health endpoints |

### Prometheus Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PROMETHEUS_NAMESPACE` | ❌ | cattle-monitoring-system | Namespace containing Prometheus service |
| `PROMETHEUS_SERVICE` | ❌ | rancher-monitoring-prometheus | Prometheus service name |
| `PROMETHEUS_PORT` | ❌ | 9090 | Prometheus service port |

### Loki Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LOKI_NAMESPACE` | ❌ | cattle-logging-system | Namespace containing Loki service |
| `LOKI_SERVICE` | ❌ | rancher-logging-loki | Loki service name |
| `LOKI_PORT` | ❌ | 3100 | Loki service port |

### Custom Remote Service Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REMOTE_NAMESPACE` | ❌ | "" | Namespace for custom service |
| `REMOTE_SERVICE` | ❌ | "" | Custom service name |
| `REMOTE_PORT` | ❌ | "" | Custom service port |

## Configuration Examples

### Basic Configuration

For a standard Rancher setup with default monitoring stack:

```bash
export RANCHER_API_ENDPOINT="https://rancher.example.com"
export RANCHER_API_ACCESS_KEY="token-abc123"
export RANCHER_API_SECRET_KEY="your-secret-key"
export CLUSTER_ID="c-m-xyz789"
export CLUSTER_NAME="production-cluster"
```

### Custom Monitoring Stack

For clusters with custom Prometheus/Loki deployments:

```bash
# Core configuration
export RANCHER_API_ENDPOINT="https://rancher.example.com"
export RANCHER_API_ACCESS_KEY="token-abc123"
export RANCHER_API_SECRET_KEY="your-secret-key"
export CLUSTER_ID="c-m-xyz789"

# Custom Prometheus
export PROMETHEUS_NAMESPACE="monitoring"
export PROMETHEUS_SERVICE="prometheus-server"
export PROMETHEUS_PORT="9090"

# Custom Loki
export LOKI_NAMESPACE="logging"
export LOKI_SERVICE="loki"
export LOKI_PORT="3100"
```

### Custom Service Relay

To relay a custom service (not Prometheus/Loki):

```bash
# Core configuration
export RANCHER_API_ENDPOINT="https://rancher.example.com"
export RANCHER_API_ACCESS_KEY="token-abc123"
export RANCHER_API_SECRET_KEY="your-secret-key"
export CLUSTER_ID="c-m-xyz789"

# Custom service
export REMOTE_NAMESPACE="monitoring"
export REMOTE_SERVICE="alertmanager"
export REMOTE_PORT="9093"
```

### Debug Configuration

For troubleshooting and development:

```bash
# Enable debug logging
export DEBUG=true

# Custom metrics port
export METRICS_PORT=8080

# All other config...
export RANCHER_API_ENDPOINT="https://rancher.example.com"
# ... etc
```

## Helm Chart Configuration

### values.yaml Structure

```yaml
# Rancher connection settings
rancher:
  apiEndpoint: "https://rancher.example.com"
  clusterId: "c-m-abc123xyz"
  clusterName: "production-cluster-1"
  auth:
    # Option 1: Use existing Kubernetes secret (recommended)
    existingSecret: "rancher-api-credentials"
    accessKeySecretKey: "access-key"
    secretKeySecretKey: "secret-key"
    
    # Option 2: Inline credentials (not recommended for production)
    accessKey: "token-abc123"
    secretKey: "your-secret-key"

# Service monitoring configuration
monitoring:
  prometheus:
    namespace: "cattle-monitoring-system"
    service: "rancher-monitoring-prometheus" 
    port: "9090"
  loki:
    namespace: "cattle-logging-system"
    service: "rancher-logging-loki"
    port: "3100"
  remote:
    namespace: ""
    service: ""
    port: ""

# Application settings
app:
  debug: false
  metricsPort: 9000

# Kubernetes deployment settings
image:
  repository: supporttools/rancher-monitoring-relay
  tag: "latest"
  pullPolicy: IfNotPresent

replicaCount: 1

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi

# Service account
serviceAccount:
  create: true
  annotations: {}
  name: ""

# Pod security
podSecurityContext:
  fsGroup: 1001
  runAsGroup: 1001
  runAsNonRoot: true
  runAsUser: 1001

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1001

# Service
service:
  type: ClusterIP
  port: 9000
  targetPort: 9000

# Health checks
healthCheck:
  enabled: true
  path: /health
  initialDelaySeconds: 10
  periodSeconds: 30
  timeoutSeconds: 5
  successThreshold: 1
  failureThreshold: 3

# Monitoring integration
monitoring:
  serviceMonitor:
    enabled: false
    interval: 30s
    scrapeTimeout: 10s
    labels: {}
    annotations: {}

# Autoscaling (if needed)
autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 3
  targetCPUUtilizationPercentage: 80

# Node selection
nodeSelector: {}
tolerations: []
affinity: {}

# Ingress (if external access needed)
ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts:
    - host: monitoring-relay.example.com
      paths:
        - path: /
          pathType: Prefix
  tls: []
```

### Production Configuration Example

```yaml
# production-values.yaml
rancher:
  apiEndpoint: "https://rancher.company.com"
  clusterId: "c-m-prod1234"
  clusterName: "production-cluster-east"
  auth:
    existingSecret: "rancher-prod-credentials"

monitoring:
  prometheus:
    namespace: "cattle-monitoring-system"
    service: "rancher-monitoring-prometheus"
    port: "9090"
  loki:
    namespace: "cattle-logging-system"  
    service: "rancher-logging-loki"
    port: "3100"

app:
  debug: false
  metricsPort: 9000

image:
  tag: "0.3.5"
  pullPolicy: Always

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

# Enable Prometheus monitoring
monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s
    labels:
      monitoring: "prometheus"

# Pod anti-affinity for HA
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchLabels:
            app.kubernetes.io/name: rancher-monitoring-relay
        topologyKey: kubernetes.io/hostname

# Multiple replicas for HA
replicaCount: 2

# Enable autoscaling
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 70
```

## Service Proxy URL Patterns

The relay constructs service proxy URLs following this pattern:

```
{RANCHER_API_ENDPOINT}/k8s/clusters/{CLUSTER_ID}/api/v1/namespaces/{NAMESPACE}/services/{SERVICE}:{PORT}/proxy/
```

### Examples

For a cluster `c-m-abc123` with Rancher at `https://rancher.example.com`:

**Prometheus:**
```
https://rancher.example.com/k8s/clusters/c-m-abc123/api/v1/namespaces/cattle-monitoring-system/services/rancher-monitoring-prometheus:9090/proxy/
```

**Loki:**
```  
https://rancher.example.com/k8s/clusters/c-m-abc123/api/v1/namespaces/cattle-logging-system/services/rancher-logging-loki:3100/proxy/
```

**Custom Service:**
```
https://rancher.example.com/k8s/clusters/c-m-abc123/api/v1/namespaces/monitoring/services/alertmanager:9093/proxy/
```

## Advanced Configuration

### Multi-Cluster Setup

For monitoring multiple clusters, deploy separate instances with unique configurations:

```yaml
# cluster1-values.yaml
rancher:
  clusterId: "c-m-cluster1"
  clusterName: "production-east"
fullnameOverride: "cluster1-monitoring-relay"

---
# cluster2-values.yaml  
rancher:
  clusterId: "c-m-cluster2"
  clusterName: "production-west"
fullnameOverride: "cluster2-monitoring-relay"
```

Deploy each:

```bash
helm install cluster1 supporttools/rancher-monitoring-relay -f cluster1-values.yaml
helm install cluster2 supporttools/rancher-monitoring-relay -f cluster2-values.yaml
```

### Custom Health Check Endpoints

The relay provides health endpoints that can be customized:

| Endpoint | Purpose | HTTP Method |
|----------|---------|-------------|
| `/health` | Basic Rancher API connectivity | GET |
| `/ready` | Service connectivity via proxy | GET |
| `/version` | Build and version information | GET |
| `/metrics` | Prometheus metrics | GET |

### Prometheus ServiceMonitor

Enable automatic Prometheus scraping:

```yaml
monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
    labels:
      monitoring: "prometheus"
      team: "platform"
    annotations:
      prometheus.io/scrape: "true"
```

### Resource Limits and Requests

Adjust based on your cluster size and monitoring load:

```yaml
resources:
  requests:
    cpu: 100m      # Minimum CPU required
    memory: 128Mi  # Minimum memory required
  limits:
    cpu: 500m      # Maximum CPU allowed
    memory: 512Mi  # Maximum memory allowed
```

### Security Context

The relay runs with minimal privileges:

```yaml
securityContext:
  allowPrivilegeEscalation: false  # Prevent privilege escalation
  capabilities:
    drop:
    - ALL                          # Drop all capabilities
  readOnlyRootFilesystem: true     # Read-only root filesystem
  runAsNonRoot: true               # Run as non-root user
  runAsUser: 1001                  # Specific user ID
```

## Environment-Specific Configurations

### Development Environment

```yaml
app:
  debug: true
  metricsPort: 9000

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 256Mi

replicaCount: 1

monitoring:
  serviceMonitor:
    enabled: false
```

### Staging Environment

```yaml
app:
  debug: false
  metricsPort: 9000

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 300m
    memory: 384Mi

replicaCount: 1

monitoring:
  serviceMonitor:
    enabled: true
    interval: 60s
```

### Production Environment

```yaml
app:
  debug: false
  metricsPort: 9000

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

replicaCount: 2

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5

monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchLabels:
            app.kubernetes.io/name: rancher-monitoring-relay
        topologyKey: kubernetes.io/hostname
```

## Validation

After configuration, validate your setup:

```bash
# Check configuration values
helm get values my-monitoring-relay

# Test endpoints
kubectl port-forward svc/my-monitoring-relay 9000:9000
curl http://localhost:9000/health
curl http://localhost:9000/ready
curl http://localhost:9000/version

# Check logs
kubectl logs -f deployment/my-monitoring-relay
```

## Next Steps

- [Usage Examples](usage.md) - Real-world usage scenarios
- [Troubleshooting](troubleshooting.md) - Common issues and solutions