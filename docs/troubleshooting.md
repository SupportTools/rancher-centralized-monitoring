# Troubleshooting Guide

This guide helps diagnose and resolve common issues with the Rancher Centralized Monitoring Relay.

## Common Issues

### 1. Cannot Connect to Rancher API

**Symptoms:**
- Relay fails to start
- Health check returns 503
- Error: "Failed to connect to Rancher API"

**Possible Causes & Solutions:**

#### Invalid API Credentials
```bash
# Check credentials format
echo $RANCHER_API_ACCESS_KEY  # Should be: token-xxxxx
echo $RANCHER_API_SECRET_KEY  # Should be: secret key string

# Test credentials manually
curl -u "${RANCHER_API_ACCESS_KEY}:${RANCHER_API_SECRET_KEY}" \
  "${RANCHER_API_ENDPOINT}/v3/clusters"
```

#### Network Connectivity
```bash
# Test network connectivity
ping rancher.example.com
curl -I https://rancher.example.com

# Check DNS resolution
nslookup rancher.example.com
dig rancher.example.com
```

#### Certificate Issues
```bash
# Test SSL certificate
openssl s_client -connect rancher.example.com:443 -servername rancher.example.com

# If self-signed certs, add CA bundle
curl --cacert /path/to/ca-bundle.crt https://rancher.example.com
```

#### Firewall/Proxy Issues
```bash
# Check proxy settings
echo $HTTP_PROXY
echo $HTTPS_PROXY
echo $NO_PROXY

# Test through proxy
curl --proxy $HTTP_PROXY https://rancher.example.com
```

### 2. Service Proxy Connectivity Failed

**Symptoms:**
- Health check passes but ready check fails
- Error: "Failed to connect to [service] service"
- HTTP 404 or 503 errors

**Possible Causes & Solutions:**

#### Incorrect Service Configuration
```bash
# Verify service exists in remote cluster
kubectl --kubeconfig=/path/to/remote/kubeconfig \
  get services -n cattle-monitoring-system

# Check service ports
kubectl --kubeconfig=/path/to/remote/kubeconfig \
  describe service rancher-monitoring-prometheus -n cattle-monitoring-system
```

#### Wrong Cluster ID
```bash
# List available clusters
curl -u "${RANCHER_API_ACCESS_KEY}:${RANCHER_API_SECRET_KEY}" \
  "${RANCHER_API_ENDPOINT}/v3/clusters" | jq '.data[] | {id, name}'

# Verify cluster exists and is active
curl -u "${RANCHER_API_ACCESS_KEY}:${RANCHER_API_SECRET_KEY}" \
  "${RANCHER_API_ENDPOINT}/v3/clusters/${CLUSTER_ID}"
```

#### Service Not Running
```bash
# Check if service is running in remote cluster
kubectl --kubeconfig=/path/to/remote/kubeconfig \
  get pods -n cattle-monitoring-system -l app=prometheus

# Check service endpoints
kubectl --kubeconfig=/path/to/remote/kubeconfig \
  get endpoints rancher-monitoring-prometheus -n cattle-monitoring-system
```

#### Network Policy Issues
```bash
# Check for network policies blocking access
kubectl --kubeconfig=/path/to/remote/kubeconfig \
  get networkpolicies -A

# Test direct service access from within cluster
kubectl --kubeconfig=/path/to/remote/kubeconfig run test-pod \
  --image=curlimages/curl --rm -it -- \
  curl http://rancher-monitoring-prometheus.cattle-monitoring-system.svc.cluster.local:9090
```

### 3. Deployment Issues

#### Pod Crashes or Fails to Start

**Check pod status:**
```bash
kubectl get pods -l app.kubernetes.io/name=rancher-monitoring-relay
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

**Common crashes:**

**Missing Environment Variables:**
```yaml
# Fix: Ensure all required env vars are set
env:
- name: RANCHER_API_ENDPOINT
  value: "https://rancher.example.com"  # Must not be empty
- name: CLUSTER_ID
  value: "c-m-xxxxx"                   # Must not be empty
```

**Resource Limits:**
```yaml
# Fix: Increase resource limits
resources:
  requests:
    memory: 128Mi
    cpu: 100m
  limits:
    memory: 256Mi  # Increase if OOM killed
    cpu: 200m
```

**Permission Issues:**
```yaml
# Fix: Ensure proper security context
securityContext:
  runAsUser: 1001
  runAsGroup: 1001
  runAsNonRoot: true
  allowPrivilegeEscalation: false
```

#### Image Pull Issues

**Check image availability:**
```bash
# Verify image exists
docker pull supporttools/rancher-monitoring-relay:latest

# Check image pull secrets
kubectl get pods <pod-name> -o yaml | grep -A5 imagePullSecrets
```

**Fix: Add image pull secrets if needed:**
```yaml
imagePullSecrets:
- name: docker-registry-secret
```

#### Service Account Issues

**Check RBAC permissions:**
```bash
kubectl get serviceaccount rancher-monitoring-relay
kubectl get clusterrolebinding | grep rancher-monitoring-relay
```

### 4. Health Check Failures

#### Health Endpoint Returns 503

**Check logs for specific error:**
```bash
kubectl logs -f deployment/rancher-monitoring-relay | grep -i error
```

**Common issues:**
- API credentials expired or revoked
- Rancher server unreachable
- Network connectivity problems

#### Ready Endpoint Always Fails

**Debug service connectivity:**
```bash
# Enable debug logging
helm upgrade my-relay supporttools/rancher-monitoring-relay \
  --set app.debug=true

# Check which services are failing
kubectl logs -f deployment/rancher-monitoring-relay | grep "Failed to connect"
```

**Test service proxy URLs manually:**
```bash
# Port forward to relay
kubectl port-forward svc/rancher-monitoring-relay 9000:9000

# Test proxy URLs
curl "http://localhost:9000/proxy/prometheus/"
curl "http://localhost:9000/proxy/loki/"
```

### 5. Performance Issues

#### High CPU/Memory Usage

**Monitor resource usage:**
```bash
kubectl top pods -l app.kubernetes.io/name=rancher-monitoring-relay
```

**Solutions:**
```yaml
# Increase resource limits
resources:
  limits:
    cpu: 500m     # Increase from 200m
    memory: 512Mi # Increase from 256Mi

# Enable autoscaling
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 70
```

#### Slow Response Times

**Check network latency:**
```bash
# Test latency to Rancher server
curl -w "@curl-format.txt" -o /dev/null -s https://rancher.example.com

# Create curl-format.txt:
cat > curl-format.txt << 'EOF'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF
```

**Optimize configuration:**
```yaml
# Reduce health check frequency for better performance
healthCheck:
  initialDelaySeconds: 10
  periodSeconds: 60        # Increase from 30
  timeoutSeconds: 10       # Increase from 5
  failureThreshold: 5      # Increase from 3
```

### 6. Configuration Issues

#### Helm Values Not Applied

**Check current values:**
```bash
helm get values my-relay
helm get manifest my-relay
```

**Common issues:**
- Values file not found or not applied
- Wrong value path in YAML
- Template rendering errors

**Debug template rendering:**
```bash
helm template my-relay supporttools/rancher-monitoring-relay \
  --values values.yaml --debug
```

#### Environment Variables Not Set

**Check pod environment:**
```bash
kubectl exec deployment/rancher-monitoring-relay -- env | grep -E "RANCHER|CLUSTER"
```

**Verify secret contents:**
```bash
kubectl get secret rancher-api-credentials -o yaml | base64 -d
```

### 7. Monitoring and Observability Issues

#### Metrics Not Being Scraped

**Check ServiceMonitor configuration:**
```bash
kubectl get servicemonitor -l app.kubernetes.io/name=rancher-monitoring-relay
kubectl describe servicemonitor <servicemonitor-name>
```

**Verify Prometheus discovery:**
```bash
# Check Prometheus targets
kubectl port-forward svc/prometheus 9090:9090
# Navigate to http://localhost:9090/targets
```

**Fix ServiceMonitor labels:**
```yaml
monitoring:
  serviceMonitor:
    enabled: true
    labels:
      prometheus: kube-prometheus  # Match your Prometheus selector
      role: alert-rules
```

#### Missing Logs

**Check log output:**
```bash
kubectl logs deployment/rancher-monitoring-relay --tail=100
```

**Enable debug logging:**
```yaml
app:
  debug: true
```

### 8. Security and Permission Issues

#### RBAC Errors

**Check service account permissions:**
```bash
kubectl auth can-i get pods --as=system:serviceaccount:default:rancher-monitoring-relay
kubectl describe clusterrolebinding | grep rancher-monitoring-relay
```

#### Pod Security Policy Violations

**Check PSP violations:**
```bash
kubectl get events --field-selector reason=FailedCreate
```

**Fix security contexts:**
```yaml
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
```

## Diagnostic Commands

### Quick Health Check Script

```bash
#!/bin/bash
# health-check.sh

set -e

NAMESPACE=${1:-default}
DEPLOYMENT=${2:-rancher-monitoring-relay}

echo "=== Rancher Monitoring Relay Health Check ==="

echo "1. Checking deployment status..."
kubectl get deployment $DEPLOYMENT -n $NAMESPACE

echo "2. Checking pod status..."
kubectl get pods -l app.kubernetes.io/name=rancher-monitoring-relay -n $NAMESPACE

echo "3. Checking recent logs..."
kubectl logs deployment/$DEPLOYMENT -n $NAMESPACE --tail=10

echo "4. Testing health endpoint..."
POD=$(kubectl get pod -l app.kubernetes.io/name=rancher-monitoring-relay -n $NAMESPACE -o jsonpath='{.items[0].metadata.name}')
kubectl exec $POD -n $NAMESPACE -- wget -q --spider http://localhost:9000/health && echo "Health: OK" || echo "Health: FAILED"

echo "5. Testing ready endpoint..."
kubectl exec $POD -n $NAMESPACE -- wget -q --spider http://localhost:9000/ready && echo "Ready: OK" || echo "Ready: FAILED"

echo "6. Checking service connectivity..."
kubectl port-forward deployment/$DEPLOYMENT -n $NAMESPACE 9000:9000 &
PF_PID=$!
sleep 2
curl -s http://localhost:9000/version | jq . || echo "Version endpoint failed"
kill $PF_PID

echo "=== Health check complete ==="
```

### Log Analysis Script

```bash
#!/bin/bash
# analyze-logs.sh

NAMESPACE=${1:-default}
DEPLOYMENT=${2:-rancher-monitoring-relay}

echo "=== Log Analysis ==="

echo "1. Error messages:"
kubectl logs deployment/$DEPLOYMENT -n $NAMESPACE | grep -i error | tail -5

echo "2. Connection issues:"
kubectl logs deployment/$DEPLOYMENT -n $NAMESPACE | grep -i "failed to connect" | tail -5

echo "3. Service connectivity:"
kubectl logs deployment/$DEPLOYMENT -n $NAMESPACE | grep -E "(Successfully connected|Failed.*service)" | tail -10

echo "4. Recent health checks:"
kubectl logs deployment/$DEPLOYMENT -n $NAMESPACE | grep -E "(HealthzHandler|ReadyzHandler)" | tail -5
```

### Network Diagnostic Script

```bash
#!/bin/bash
# network-debug.sh

RANCHER_ENDPOINT=$1
CLUSTER_ID=$2

if [[ -z "$RANCHER_ENDPOINT" || -z "$CLUSTER_ID" ]]; then
  echo "Usage: $0 <rancher-endpoint> <cluster-id>"
  exit 1
fi

echo "=== Network Diagnostics ==="

echo "1. DNS resolution:"
nslookup $(echo $RANCHER_ENDPOINT | sed 's|https*://||' | cut -d/ -f1)

echo "2. Connectivity test:"
curl -I $RANCHER_ENDPOINT

echo "3. SSL certificate check:"
echo | openssl s_client -connect $(echo $RANCHER_ENDPOINT | sed 's|https*://||' | cut -d/ -f1):443 -servername $(echo $RANCHER_ENDPOINT | sed 's|https*://||' | cut -d/ -f1) 2>/dev/null | openssl x509 -noout -dates

echo "4. API endpoint test:"
echo "Test with: curl -u 'token-xxx:secret' $RANCHER_ENDPOINT/v3/clusters/$CLUSTER_ID"
```

## Performance Tuning

### Resource Optimization

```yaml
# For small clusters (< 10 services)
resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 128Mi

# For medium clusters (10-50 services)  
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 300m
    memory: 256Mi

# For large clusters (> 50 services)
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### Health Check Tuning

```yaml
# For stable networks
healthCheck:
  initialDelaySeconds: 10
  periodSeconds: 60
  timeoutSeconds: 5
  failureThreshold: 3

# For unstable networks
healthCheck:
  initialDelaySeconds: 20
  periodSeconds: 120
  timeoutSeconds: 15
  failureThreshold: 5
```

## Getting Help

### Enable Debug Logging

```yaml
app:
  debug: true
```

### Collect Diagnostics

```bash
# Create diagnostics bundle
kubectl get all -l app.kubernetes.io/name=rancher-monitoring-relay -o yaml > diagnostics.yaml
kubectl logs deployment/rancher-monitoring-relay --previous > previous-logs.txt
kubectl logs deployment/rancher-monitoring-relay > current-logs.txt
kubectl describe deployment rancher-monitoring-relay >> diagnostics.yaml
kubectl get events --sort-by=.lastTimestamp | tail -20 >> diagnostics.yaml
```

### Support Channels

- **GitHub Issues**: [rancher-centralized-monitoring/issues](https://github.com/supporttools/rancher-centralized-monitoring/issues)
- **Documentation**: [docs.support.tools](https://docs.support.tools)
- **Community**: [community.support.tools](https://community.support.tools)

### Include in Support Request

1. **Environment Information:**
   - Kubernetes version
   - Rancher version
   - Relay version
   - Deployment method (Helm/kubectl/Docker)

2. **Configuration:**
   - Sanitized configuration files
   - Environment variables (without secrets)

3. **Diagnostics:**
   - Recent logs
   - Pod status and events
   - Network connectivity test results

4. **Problem Description:**
   - Expected behavior
   - Actual behavior
   - Steps to reproduce
   - When the issue started

## Recovery Procedures

### Complete Restart

```bash
# Helm deployment
helm upgrade my-relay supporttools/rancher-monitoring-relay --recreate-pods

# Kubectl deployment
kubectl rollout restart deployment/rancher-monitoring-relay
```

### Reset Configuration

```bash
# Delete and reinstall with clean config
helm uninstall my-relay
helm install my-relay supporttools/rancher-monitoring-relay -f clean-values.yaml
```

### Emergency Rollback

```bash
# Rollback to previous version
helm rollback my-relay

# Or specific version
helm rollback my-relay 2
```