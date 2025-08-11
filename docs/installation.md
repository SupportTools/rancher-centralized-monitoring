# Installation Guide

This guide provides detailed instructions for installing the Rancher Centralized Monitoring Relay in various environments.

## Prerequisites

### System Requirements

- **Rancher Server**: Version 2.5+ with remote clusters configured
- **Kubernetes Cluster**: For deployment (Rancher local cluster recommended)
- **Network Access**: Connectivity from deployment cluster to Rancher server
- **Resource Requirements**:
  - CPU: 100m (request), 200m (limit)
  - Memory: 128Mi (request), 256Mi (limit)
  - Storage: Minimal (stateless application)

### Required Credentials

You'll need valid Rancher API credentials with appropriate permissions:

1. **API Access Key**: Generated from Rancher UI
2. **API Secret Key**: Generated from Rancher UI
3. **Cluster ID**: Target remote cluster identifier (format: `c-xxxxxxx`)

### Generating Rancher API Keys

1. Login to your Rancher server UI
2. Click on your profile (top-right corner)
3. Select "API & Keys"
4. Click "Add Key"
5. Set appropriate scope and permissions
6. Save the generated `Access Key` and `Secret Key`

## Installation Methods

## 1. Helm Installation (Recommended)

### Add Helm Repository

```bash
# Add the Support Tools Helm repository
helm repo add supporttools https://charts.support.tools/
helm repo update
```

### Basic Installation

```bash
helm install my-monitoring-relay supporttools/rancher-monitoring-relay \
  --set rancher.apiEndpoint="https://your-rancher-server" \
  --set rancher.clusterId="c-xxxxx" \
  --set rancher.auth.accessKey="token-xxxxx" \
  --set rancher.auth.secretKey="your-secret-key"
```

### Installation with Custom Values

Create a `values.yaml` file:

```yaml
# values.yaml
rancher:
  apiEndpoint: "https://rancher.example.com"
  clusterId: "c-m-abc123xyz"
  clusterName: "production-cluster-1"
  auth:
    accessKey: "token-abc123"
    secretKey: "your-secret-key-here"

monitoring:
  prometheus:
    namespace: "monitoring"
    service: "prometheus-server"
    port: "9090"
  loki:
    namespace: "logging"
    service: "loki"
    port: "3100"

app:
  debug: false
  metricsPort: 9000

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi

# Enable ServiceMonitor for Prometheus scraping
monitoring:
  serviceMonitor:
    enabled: true
    interval: 30s
```

Install with custom values:

```bash
helm install my-monitoring-relay supporttools/rancher-monitoring-relay \
  --values values.yaml
```

### Installation with Existing Secret

For production environments, create secrets separately:

```bash
# Create secret with API credentials
kubectl create secret generic rancher-api-credentials \
  --from-literal=access-key="token-xxxxx" \
  --from-literal=secret-key="your-secret-key"

# Install using existing secret
helm install my-monitoring-relay supporttools/rancher-monitoring-relay \
  --set rancher.apiEndpoint="https://rancher.example.com" \
  --set rancher.clusterId="c-xxxxx" \
  --set rancher.auth.existingSecret="rancher-api-credentials"
```

## 2. Kubernetes YAML Installation

### Create Namespace

```bash
kubectl create namespace rancher-monitoring-relay
```

### Create Secret

```bash
kubectl create secret generic rancher-api-credentials \
  --namespace rancher-monitoring-relay \
  --from-literal=access-key="token-xxxxx" \
  --from-literal=secret-key="your-secret-key"
```

### Deploy Application

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rancher-monitoring-relay
  namespace: rancher-monitoring-relay
  labels:
    app: rancher-monitoring-relay
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rancher-monitoring-relay
  template:
    metadata:
      labels:
        app: rancher-monitoring-relay
    spec:
      containers:
      - name: relay
        image: supporttools/rancher-monitoring-relay:latest
        ports:
        - containerPort: 9000
          name: metrics
        env:
        - name: RANCHER_API_ENDPOINT
          value: "https://rancher.example.com"
        - name: CLUSTER_ID
          value: "c-xxxxx"
        - name: RANCHER_API_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: rancher-api-credentials
              key: access-key
        - name: RANCHER_API_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: rancher-api-credentials
              key: secret-key
        - name: PROMETHEUS_NAMESPACE
          value: "cattle-monitoring-system"
        - name: PROMETHEUS_SERVICE
          value: "rancher-monitoring-prometheus"
        - name: PROMETHEUS_PORT
          value: "9090"
        - name: LOKI_NAMESPACE
          value: "cattle-logging-system"
        - name: LOKI_SERVICE
          value: "rancher-logging-loki"
        - name: LOKI_PORT
          value: "3100"
        livenessProbe:
          httpGet:
            path: /health
            port: 9000
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 9000
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1001
---
apiVersion: v1
kind: Service
metadata:
  name: rancher-monitoring-relay
  namespace: rancher-monitoring-relay
  labels:
    app: rancher-monitoring-relay
spec:
  selector:
    app: rancher-monitoring-relay
  ports:
  - port: 9000
    targetPort: 9000
    name: metrics
  type: ClusterIP
```

Apply the deployment:

```bash
kubectl apply -f deployment.yaml
```

## 3. Docker Installation

### Basic Docker Run

```bash
docker run -d \
  --name rancher-monitoring-relay \
  -p 9000:9000 \
  -e RANCHER_API_ENDPOINT="https://rancher.example.com" \
  -e RANCHER_API_ACCESS_KEY="token-xxxxx" \
  -e RANCHER_API_SECRET_KEY="your-secret-key" \
  -e CLUSTER_ID="c-xxxxx" \
  -e DEBUG=false \
  supporttools/rancher-monitoring-relay:latest
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  rancher-monitoring-relay:
    image: supporttools/rancher-monitoring-relay:latest
    container_name: rancher-monitoring-relay
    restart: unless-stopped
    ports:
      - "9000:9000"
    environment:
      - RANCHER_API_ENDPOINT=https://rancher.example.com
      - RANCHER_API_ACCESS_KEY=token-xxxxx
      - RANCHER_API_SECRET_KEY=your-secret-key
      - CLUSTER_ID=c-xxxxx
      - DEBUG=false
      - PROMETHEUS_NAMESPACE=cattle-monitoring-system
      - PROMETHEUS_SERVICE=rancher-monitoring-prometheus
      - PROMETHEUS_PORT=9090
      - LOKI_NAMESPACE=cattle-logging-system
      - LOKI_SERVICE=rancher-logging-loki
      - LOKI_PORT=3100
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

Start with Docker Compose:

```bash
docker-compose up -d
```

## 4. Binary Installation

### Download Binary

```bash
# Download latest release
wget https://github.com/supporttools/rancher-centralized-monitoring/releases/latest/download/rancher-monitoring-relay-linux-amd64

# Make executable
chmod +x rancher-monitoring-relay-linux-amd64
sudo mv rancher-monitoring-relay-linux-amd64 /usr/local/bin/rancher-monitoring-relay
```

### Create Service

```bash
# Create user
sudo useradd --system --shell /bin/false rancher-monitoring-relay

# Create environment file
sudo tee /etc/rancher-monitoring-relay.env << EOF
RANCHER_API_ENDPOINT=https://rancher.example.com
RANCHER_API_ACCESS_KEY=token-xxxxx
RANCHER_API_SECRET_KEY=your-secret-key
CLUSTER_ID=c-xxxxx
DEBUG=false
METRICS_PORT=9000
EOF

# Create systemd service
sudo tee /etc/systemd/system/rancher-monitoring-relay.service << EOF
[Unit]
Description=Rancher Centralized Monitoring Relay
After=network.target

[Service]
Type=simple
User=rancher-monitoring-relay
Group=rancher-monitoring-relay
ExecStart=/usr/local/bin/rancher-monitoring-relay
EnvironmentFile=/etc/rancher-monitoring-relay.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable rancher-monitoring-relay
sudo systemctl start rancher-monitoring-relay
```

## Multiple Cluster Deployment

To monitor multiple remote clusters, deploy separate instances with different configurations:

```bash
# Cluster 1
helm install cluster1-relay supporttools/rancher-monitoring-relay \
  --set rancher.apiEndpoint="https://rancher.example.com" \
  --set rancher.clusterId="c-cluster1" \
  --set rancher.clusterName="production-east" \
  --set rancher.auth.accessKey="token-xxxxx" \
  --set rancher.auth.secretKey="your-secret-key" \
  --set fullnameOverride="cluster1-monitoring-relay"

# Cluster 2
helm install cluster2-relay supporttools/rancher-monitoring-relay \
  --set rancher.apiEndpoint="https://rancher.example.com" \
  --set rancher.clusterId="c-cluster2" \
  --set rancher.clusterName="production-west" \
  --set rancher.auth.accessKey="token-xxxxx" \
  --set rancher.auth.secretKey="your-secret-key" \
  --set fullnameOverride="cluster2-monitoring-relay"
```

## Post-Installation Verification

### Check Deployment Status

```bash
# Helm deployment
helm status my-monitoring-relay

# Kubernetes deployment
kubectl get pods -n rancher-monitoring-relay
kubectl logs -f deployment/rancher-monitoring-relay -n rancher-monitoring-relay

# Docker
docker logs rancher-monitoring-relay
```

### Test Endpoints

```bash
# Health check
curl http://localhost:9000/health

# Ready check (tests service connectivity)
curl http://localhost:9000/ready

# Version information
curl http://localhost:9000/version

# Metrics
curl http://localhost:9000/metrics
```

### Verify Service Proxy Connectivity

```bash
# Check logs for successful connections
kubectl logs -f deployment/rancher-monitoring-relay -n rancher-monitoring-relay | grep "Successfully connected"
```

## Upgrading

### Helm Upgrade

```bash
# Update repository
helm repo update

# Upgrade with new values
helm upgrade my-monitoring-relay supporttools/rancher-monitoring-relay \
  --values values.yaml

# Upgrade to specific version
helm upgrade my-monitoring-relay supporttools/rancher-monitoring-relay \
  --version 0.3.1
```

### Docker Upgrade

```bash
# Pull new image
docker pull supporttools/rancher-monitoring-relay:latest

# Stop and remove old container
docker stop rancher-monitoring-relay
docker rm rancher-monitoring-relay

# Start new container
docker run -d \
  --name rancher-monitoring-relay \
  [... same parameters as before ...]
  supporttools/rancher-monitoring-relay:latest
```

## Uninstallation

### Helm Uninstall

```bash
helm uninstall my-monitoring-relay
```

### Kubernetes Uninstall

```bash
kubectl delete -f deployment.yaml
kubectl delete namespace rancher-monitoring-relay
```

### Docker Uninstall

```bash
docker stop rancher-monitoring-relay
docker rm rancher-monitoring-relay
docker rmi supporttools/rancher-monitoring-relay
```

## Next Steps

- [Configuration Guide](configuration.md) - Configure the relay for your environment
- [Usage Examples](usage.md) - Real-world usage scenarios
- [Troubleshooting](troubleshooting.md) - Common issues and solutions