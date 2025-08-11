# Usage Examples

This guide provides real-world usage scenarios for the Rancher Centralized Monitoring Relay.

## Common Use Cases

### 1. Centralized Prometheus Monitoring

Monitor multiple remote clusters from a central Prometheus instance.

#### Scenario
You have:
- Central monitoring cluster with Prometheus/Grafana
- 5 remote production clusters with their own Prometheus instances
- Need to scrape metrics from all remote clusters centrally

#### Solution

Deploy monitoring relays for each remote cluster:

```bash
# Deploy relay for each cluster
for cluster in cluster1 cluster2 cluster3 cluster4 cluster5; do
  helm install ${cluster}-relay supporttools/rancher-monitoring-relay \
    --set rancher.apiEndpoint="https://rancher.company.com" \
    --set rancher.clusterId="c-m-${cluster}" \
    --set rancher.clusterName="${cluster}" \
    --set rancher.auth.existingSecret="rancher-credentials" \
    --set fullnameOverride="${cluster}-monitoring-relay"
done
```

Configure central Prometheus to scrape through the relays:

```yaml
# prometheus.yml
global:
  scrape_interval: 30s

scrape_configs:
  # Cluster 1 Prometheus via relay
  - job_name: 'cluster1-prometheus'
    static_configs:
    - targets: ['cluster1-monitoring-relay:9000']
    metrics_path: /proxy/prometheus/api/v1/query
    params:
      query: ['up']
    
  # Cluster 2 Prometheus via relay  
  - job_name: 'cluster2-prometheus'
    static_configs:
    - targets: ['cluster2-monitoring-relay:9000']
    metrics_path: /proxy/prometheus/api/v1/query
    params:
      query: ['up']
```

### 2. Centralized Log Aggregation with Loki

Aggregate logs from remote Loki instances.

#### Scenario
- Central logging cluster with Grafana
- Multiple remote clusters with Loki deployments
- Need centralized log querying and alerting

#### Solution

```yaml
# loki-federation-values.yaml
rancher:
  apiEndpoint: "https://rancher.company.com"
  clusterId: "c-m-prod-east"
  clusterName: "production-east"
  auth:
    existingSecret: "rancher-credentials"

monitoring:
  loki:
    namespace: "cattle-logging-system"
    service: "rancher-logging-loki"
    port: "3100"
  # Disable Prometheus monitoring for this relay
  prometheus:
    namespace: ""
    
fullnameOverride: "prod-east-loki-relay"

# Add labels for identification
podAnnotations:
  cluster: "prod-east"
  service: "loki"
```

Deploy and configure Grafana to query through the relay:

```bash
helm install prod-east-loki supporttools/rancher-monitoring-relay \
  -f loki-federation-values.yaml
```

Add as Loki data source in Grafana:
```
URL: http://prod-east-loki-relay:9000/proxy/loki
```

### 3. Custom Service Monitoring

Monitor custom applications in remote clusters.

#### Scenario
- Custom application with metrics endpoint on port 8080
- Application deployed in `apps` namespace as service `myapp-metrics`
- Need to scrape metrics from central monitoring

#### Solution

```yaml
# custom-app-values.yaml
rancher:
  apiEndpoint: "https://rancher.company.com"
  clusterId: "c-m-app-cluster"
  clusterName: "application-cluster"
  auth:
    existingSecret: "rancher-credentials"

monitoring:
  # Disable default services
  prometheus:
    namespace: ""
  loki:
    namespace: ""
  # Configure custom service
  remote:
    namespace: "apps"
    service: "myapp-metrics"
    port: "8080"

fullnameOverride: "myapp-metrics-relay"
```

Deploy the relay:

```bash
helm install myapp-relay supporttools/rancher-monitoring-relay \
  -f custom-app-values.yaml
```

The service will be available at:
```
http://myapp-metrics-relay:9000/proxy/
```

### 4. Multi-Environment Monitoring

Monitor dev, staging, and production environments separately.

#### Scenario
- 3 environments: dev, staging, production
- Each environment has its own Rancher-managed cluster
- Need environment-specific monitoring with proper labeling

#### Solution

Create environment-specific configurations:

```yaml
# dev-values.yaml
rancher:
  clusterId: "c-m-dev123"
  clusterName: "development"
  
app:
  debug: true
  
resources:
  requests:
    cpu: 50m
    memory: 64Mi

podAnnotations:
  environment: "development"
  team: "platform"

fullnameOverride: "dev-monitoring-relay"
```

```yaml
# staging-values.yaml  
rancher:
  clusterId: "c-m-stage456"
  clusterName: "staging"

app:
  debug: false
  
resources:
  requests:
    cpu: 100m
    memory: 128Mi

podAnnotations:
  environment: "staging"
  team: "platform"

fullnameOverride: "staging-monitoring-relay"
```

```yaml
# prod-values.yaml
rancher:
  clusterId: "c-m-prod789"
  clusterName: "production"

app:
  debug: false
  
resources:
  requests:
    cpu: 200m
    memory: 256Mi

replicaCount: 2

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5

podAnnotations:
  environment: "production" 
  team: "platform"

monitoring:
  serviceMonitor:
    enabled: true

fullnameOverride: "prod-monitoring-relay"
```

Deploy all environments:

```bash
helm install dev-relay supporttools/rancher-monitoring-relay -f dev-values.yaml
helm install staging-relay supporttools/rancher-monitoring-relay -f staging-values.yaml  
helm install prod-relay supporttools/rancher-monitoring-relay -f prod-values.yaml
```

### 5. High Availability Setup

Deploy monitoring relay with high availability.

#### Scenario
- Critical production monitoring
- Need redundancy and failover
- Multiple availability zones

#### Solution

```yaml
# ha-values.yaml
rancher:
  apiEndpoint: "https://rancher.company.com"
  clusterId: "c-m-critical-prod"
  clusterName: "critical-production"
  auth:
    existingSecret: "rancher-credentials"

# Multiple replicas
replicaCount: 3

# Enable autoscaling
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 60

# Spread across availability zones
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app.kubernetes.io/name: rancher-monitoring-relay
      topologyKey: topology.kubernetes.io/zone
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: node-role.kubernetes.io/monitoring
          operator: In
          values: ["true"]

# Resource limits for stable performance
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# Enable monitoring
monitoring:
  serviceMonitor:
    enabled: true
    interval: 15s
    scrapeTimeout: 10s

# Pod disruption budget
podDisruptionBudget:
  enabled: true
  minAvailable: 2
```

Deploy with HA configuration:

```bash
helm install critical-relay supporttools/rancher-monitoring-relay \
  -f ha-values.yaml \
  --namespace monitoring-system \
  --create-namespace
```

### 6. Service Discovery Integration

Integrate with service discovery systems.

#### Scenario
- Using Consul for service discovery
- Need to register monitoring relay endpoints
- Automatic service health checking

#### Solution

```yaml
# service-discovery-values.yaml
rancher:
  apiEndpoint: "https://rancher.company.com"  
  clusterId: "c-m-services"
  clusterName: "services-cluster"
  auth:
    existingSecret: "rancher-credentials"

# Service annotations for Consul registration
service:
  annotations:
    consul.hashicorp.com/service-name: "rancher-monitoring-relay"
    consul.hashicorp.com/service-tags: "monitoring,relay,rancher"
    consul.hashicorp.com/service-port: "9000"

# Pod annotations for Consul Connect
podAnnotations:
  consul.hashicorp.com/connect-inject: "true"
  consul.hashicorp.com/connect-service: "rancher-monitoring-relay"

# Health check configuration for Consul
healthCheck:
  enabled: true
  path: /health
  initialDelaySeconds: 10
  periodSeconds: 30
```

### 7. GitOps Integration

Deploy using GitOps with ArgoCD or Flux.

#### ArgoCD Application

```yaml
# argocd-app.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: rancher-monitoring-relay
  namespace: argocd
spec:
  project: platform
  source:
    repoURL: https://charts.support.tools/
    chart: rancher-monitoring-relay
    targetRevision: "0.3.5"
    helm:
      valueFiles:
      - values-production.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: monitoring-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
```

#### Flux HelmRelease

```yaml
# flux-helmrelease.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: rancher-monitoring-relay
  namespace: monitoring-system
spec:
  interval: 10m
  chart:
    spec:
      chart: rancher-monitoring-relay
      version: "0.3.5"
      sourceRef:
        kind: HelmRepository
        name: supporttools
        namespace: flux-system
  valuesFrom:
  - kind: Secret
    name: rancher-monitoring-config
  values:
    rancher:
      apiEndpoint: "https://rancher.company.com"
      clusterId: "c-m-flux-cluster"
```

### 8. Monitoring Multiple Services per Cluster

Monitor both Prometheus and Loki in the same cluster.

#### Scenario
- Remote cluster has both Prometheus and Loki
- Need single relay to access both services
- Different endpoints for each service

#### Solution

The relay automatically detects and monitors both services:

```yaml
# multi-service-values.yaml
rancher:
  apiEndpoint: "https://rancher.company.com"
  clusterId: "c-m-full-stack"
  clusterName: "full-stack-cluster"
  auth:
    existingSecret: "rancher-credentials"

monitoring:
  # Both services will be monitored
  prometheus:
    namespace: "cattle-monitoring-system"
    service: "rancher-monitoring-prometheus"
    port: "9090"
  loki:
    namespace: "cattle-logging-system"
    service: "rancher-logging-loki"
    port: "3100"

# Enable comprehensive health checks
healthCheck:
  enabled: true
  path: /ready  # This checks both services
  initialDelaySeconds: 15
  periodSeconds: 30
  failureThreshold: 3

fullnameOverride: "full-stack-monitoring-relay"
```

Access services via the relay:
- Prometheus: `http://full-stack-monitoring-relay:9000/proxy/prometheus/`
- Loki: `http://full-stack-monitoring-relay:9000/proxy/loki/`

### 9. Custom Metrics and Alerting

Export custom metrics about relay performance.

#### Scenario
- Need metrics about relay performance
- Want alerts on connectivity issues
- Monitor service proxy response times

#### Solution

The relay provides built-in metrics at `/metrics`:

```yaml
# monitoring-values.yaml
rancher:
  apiEndpoint: "https://rancher.company.com"
  clusterId: "c-m-monitored"
  clusterName: "monitored-cluster"
  auth:
    existingSecret: "rancher-credentials"

# Enable ServiceMonitor for Prometheus scraping
monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s
    scrapeTimeout: 10s
    labels:
      monitoring: "prometheus"
      alert: "critical"
    annotations:
      prometheus.io/scrape: "true"
      prometheus.io/path: "/metrics"
      prometheus.io/port: "9000"

# Add custom labels for alerting
podLabels:
  monitoring: "enabled"
  cluster: "monitored-cluster"
  service: "rancher-relay"
```

Create Prometheus alerts:

```yaml
# alerts.yaml
groups:
- name: rancher-monitoring-relay
  rules:
  - alert: RancherRelayDown
    expr: up{job="rancher-monitoring-relay"} == 0
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Rancher monitoring relay is down"
      description: "Relay {{ $labels.instance }} has been down for more than 2 minutes"
      
  - alert: RancherRelayServiceUnhealthy
    expr: probe_success{job="rancher-monitoring-relay"} == 0
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "Rancher relay service connectivity issues"
      description: "Relay {{ $labels.instance }} cannot reach remote services"
```

### 10. Development and Testing

Use the relay for development and testing scenarios.

#### Local Development

```bash
# Start relay locally for testing
export RANCHER_API_ENDPOINT="https://rancher-dev.company.com"
export RANCHER_API_ACCESS_KEY="token-dev123"
export RANCHER_API_SECRET_KEY="dev-secret-key"
export CLUSTER_ID="c-m-dev-cluster"
export DEBUG=true
export METRICS_PORT=9000

go run main.go
```

#### Testing with Port Forward

```bash
# Deploy to test cluster
helm install test-relay supporttools/rancher-monitoring-relay \
  --set rancher.apiEndpoint="https://rancher-test.company.com" \
  --set rancher.clusterId="c-m-test123" \
  --set rancher.auth.accessKey="token-test" \
  --set rancher.auth.secretKey="test-secret"

# Port forward for local access
kubectl port-forward svc/test-relay 9000:9000

# Test endpoints
curl http://localhost:9000/health
curl http://localhost:9000/ready
curl http://localhost:9000/version
curl http://localhost:9000/metrics
```

#### Integration Testing

```bash
# Create test script
cat << 'EOF' > test-relay.sh
#!/bin/bash
set -e

echo "Testing Rancher Monitoring Relay..."

# Test health endpoint
echo "Checking health..."
curl -f http://localhost:9000/health || exit 1

# Test ready endpoint
echo "Checking readiness..."
curl -f http://localhost:9000/ready || exit 1

# Test version endpoint
echo "Checking version..."
curl -f http://localhost:9000/version | jq . || exit 1

# Test metrics endpoint
echo "Checking metrics..."
curl -f http://localhost:9000/metrics | head -5

echo "All tests passed!"
EOF

chmod +x test-relay.sh
./test-relay.sh
```

## Best Practices

### 1. Resource Management

- Set appropriate resource requests and limits
- Use horizontal pod autoscaling for high-load scenarios
- Monitor resource usage with metrics

### 2. Security

- Use Kubernetes secrets for API credentials
- Enable pod security contexts
- Regularly rotate API keys
- Use network policies to restrict access

### 3. Monitoring

- Enable ServiceMonitor for Prometheus scraping
- Set up alerts for relay health and connectivity
- Monitor service proxy response times
- Use health checks for automatic recovery

### 4. High Availability

- Deploy multiple replicas for critical environments
- Use pod anti-affinity to spread across nodes
- Configure pod disruption budgets
- Implement proper load balancing

### 5. Configuration Management

- Use Helm values files for environment-specific settings
- Store secrets separately from configuration
- Version control your configurations
- Use GitOps for deployment automation

## Next Steps

- [Troubleshooting Guide](troubleshooting.md) - Common issues and solutions
- [Configuration Guide](configuration.md) - Complete configuration reference