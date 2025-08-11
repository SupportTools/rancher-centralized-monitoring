# Rancher Centralized Monitoring Helm Chart

This Helm chart deploys the Rancher Centralized Monitoring relay agent on Kubernetes. The relay agent runs as a deployment in the Rancher local cluster and acts as a proxy to access Prometheus and Loki monitoring services in remote clusters through Rancher's service proxy functionality.

## Overview

Each deployment instance acts as a relay for a specific remote cluster, making monitoring data from remote clusters accessible to centralized monitoring infrastructure without requiring direct cluster access.

## Installation

```bash
helm install my-monitoring-relay ./charts/rancher-centralized-monitoring
```

Or using the Helm repository:

```bash
helm repo add supporttools https://charts.support.tools/
helm install my-monitoring-relay supporttools/rancher-centralized-monitoring
```

## Configuration

The following table lists the configurable parameters of the rancher-centralized-monitoring chart and their default values.

### Core Settings

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `replicaCount` | Number of replica pods | `1` |
| `image.repository` | Container image repository | `supporttools/rancher-centralized-monitoring` |
| `image.tag` | Container image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Rancher Configuration

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `rancher.apiEndpoint` | Rancher API endpoint URL (required) | `""` |
| `rancher.clusterId` | Target cluster ID to monitor (required) | `""` |
| `rancher.clusterName` | Human-readable cluster name | `""` |
| `rancher.auth.existingSecret` | Name of existing secret with API credentials | `""` |
| `rancher.auth.accessKeySecretKey` | Key name for access key in existing secret | `access-key` |
| `rancher.auth.secretKeySecretKey` | Key name for secret key in existing secret | `secret-key` |
| `rancher.auth.accessKey` | Rancher API access key (not recommended for production) | `""` |
| `rancher.auth.secretKey` | Rancher API secret key (not recommended for production) | `""` |

### Monitoring Services Configuration

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `monitoring.prometheus.namespace` | Prometheus service namespace | `cattle-monitoring-system` |
| `monitoring.prometheus.service` | Prometheus service name | `rancher-monitoring-prometheus` |
| `monitoring.prometheus.port` | Prometheus service port | `9090` |
| `monitoring.loki.namespace` | Loki service namespace | `cattle-logging-system` |
| `monitoring.loki.service` | Loki service name | `rancher-logging-loki` |
| `monitoring.loki.port` | Loki service port | `3100` |
| `monitoring.remote.namespace` | Custom service namespace | `""` |
| `monitoring.remote.service` | Custom service name | `""` |
| `monitoring.remote.port` | Custom service port | `""` |

### Application Settings

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `app.debug` | Enable debug logging | `false` |
| `app.metricsPort` | Metrics endpoint port | `9000` |

### Service Configuration

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `9000` |
| `service.targetPort` | Target container port | `9000` |

### Resources and Scaling

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |
| `resources.limits.cpu` | CPU limit | `200m` |
| `resources.limits.memory` | Memory limit | `256Mi` |
| `autoscaling.enabled` | Enable horizontal pod autoscaling | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `1` |
| `autoscaling.maxReplicas` | Maximum replicas | `3` |

### Observability

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `observability.serviceMonitor.enabled` | Enable Prometheus ServiceMonitor | `false` |
| `observability.serviceMonitor.interval` | Scrape interval | `30s` |
| `observability.serviceMonitor.scrapeTimeout` | Scrape timeout | `10s` |
| `healthCheck.enabled` | Enable health checks | `true` |
| `healthCheck.path` | Health check endpoint path | `/health` |

## Usage Examples

### Basic Configuration

```yaml
rancher:
  apiEndpoint: "https://rancher.example.com/v3"
  clusterId: "c-m-12345678"
  clusterName: "production-cluster"
  auth:
    accessKey: "token-12345"
    secretKey: "secret-67890"

app:
  debug: false
```

### Production Configuration with Existing Secret

Create a secret with your Rancher API credentials:

```bash
kubectl create secret generic rancher-credentials \
  --from-literal=access-key=token-12345 \
  --from-literal=secret-key=secret-67890
```

Then reference it in your values:

```yaml
rancher:
  apiEndpoint: "https://rancher.example.com/v3"
  clusterId: "c-m-12345678"
  clusterName: "production-cluster"
  auth:
    existingSecret: "rancher-credentials"

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

observability:
  serviceMonitor:
    enabled: true
    interval: 15s
```

### Custom Monitoring Service Configuration

```yaml
rancher:
  apiEndpoint: "https://rancher.example.com/v3"
  clusterId: "c-m-12345678"

monitoring:
  prometheus:
    namespace: "custom-monitoring"
    service: "prometheus-server"
    port: "9090"
  loki:
    namespace: "custom-logging"
    service: "loki"
    port: "3100"
```

### Multiple Remote Services

```yaml
rancher:
  apiEndpoint: "https://rancher.example.com/v3"
  clusterId: "c-m-12345678"

monitoring:
  remote:
    namespace: "custom-namespace"
    service: "custom-service"
    port: "8080"
```

## Architecture

- **Deployment**: Runs in the Rancher local cluster
- **Service Proxy**: Uses Rancher's built-in service proxy to access remote cluster services
- **One-to-One**: Each deployment instance handles one remote cluster
- **URL Pattern**: `{RANCHER_API_ENDPOINT}/k8s/clusters/{CLUSTER_ID}/api/v1/namespaces/{namespace}/services/{service}:{port}/proxy/`

## Monitoring and Observability

The relay exposes several endpoints for monitoring and health checks:

- `/metrics` - Prometheus metrics endpoint
- `/health` - Health check endpoint for liveness and readiness probes
- Application logs with structured logging using logrus

### ServiceMonitor

Enable the ServiceMonitor to have Prometheus automatically scrape metrics:

```yaml
observability:
  serviceMonitor:
    enabled: true
    labels:
      app: monitoring
```

## Security

The chart follows security best practices:

- Non-root container execution
- Read-only root filesystem
- Dropped capabilities
- Security contexts enforced
- Secret management for API credentials

## Troubleshooting

### Debug Mode

Enable debug logging to troubleshoot connectivity issues:

```yaml
app:
  debug: true
```

### Connection Issues

Verify Rancher API connectivity:
1. Check that `rancher.apiEndpoint` is correct and accessible
2. Verify API credentials have appropriate permissions
3. Ensure the target `clusterId` exists and is accessible
4. Check that monitoring services exist in the target cluster

### Common Issues

- **403 Forbidden**: API credentials lack sufficient permissions
- **404 Not Found**: Cluster ID or service names are incorrect
- **Connection Timeout**: Network connectivity issues or incorrect endpoints