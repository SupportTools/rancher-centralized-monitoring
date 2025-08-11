# Security Guide

This guide provides comprehensive security information for deploying and operating the Rancher Centralized Monitoring Relay in production environments.

## Security Overview

The Rancher Centralized Monitoring Relay is built with security as a primary concern, implementing defense-in-depth principles and following industry best practices for secure software development and deployment.

### Security Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Centralized   │    │  Security Relay │    │  Remote Cluster │
│   Monitoring    │    │  (Hardened)     │    │   (Isolated)    │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Prometheus  │◄┼────┼►│   Relay     │◄┼────┼►│ Prometheus  │ │
│ │ Grafana     │ │    │ │ - Non-root  │ │    │ │ Loki        │ │
│ │             │ │    │ │ - Read-only │ │    │ │ Services    │ │
│ └─────────────┘ │    │ │ - Minimal   │ │    │ └─────────────┘ │
│                 │    │ │ - Secure    │ │    │                 │
└─────────────────┘    │ └─────────────┘ │    └─────────────────┘
                       └─────────────────┘
                            ^       ^
                            │       │
                       TLS/HTTPS   API Auth
                       Encrypted   + RBAC
```

## Security Features

### 🛡️ Container Security

#### Minimal Attack Surface
- **Base Image**: Built on `scratch` - no OS, shell, or unnecessary tools
- **Single Binary**: Only the relay binary is included
- **No Package Manager**: No apt, yum, or other package managers
- **No Shell**: No bash, sh, or other shells available

#### Secure Runtime
- **Non-Root User**: Runs as UID 1001 (non-privileged)
- **Read-Only Filesystem**: Root filesystem mounted read-only
- **No Privilege Escalation**: `allowPrivilegeEscalation: false`
- **Dropped Capabilities**: All Linux capabilities dropped

#### Container Hardening
```dockerfile
FROM scratch
WORKDIR /root/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/rancher-monitoring-relay /bin/
ENTRYPOINT ["/bin/rancher-monitoring-relay"]
```

### 🔐 Authentication & Authorization

#### Rancher API Authentication
- **API Key Based**: Uses Rancher API access/secret key pairs
- **HTTPS Only**: All communication encrypted in transit
- **Certificate Validation**: Full SSL/TLS certificate validation
- **Token Rotation**: Supports regular credential rotation

#### Kubernetes RBAC
```yaml
# Minimal required permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: rancher-monitoring-relay
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get"]
```

### 🔒 Network Security

#### Secure Communications
- **TLS 1.2+**: Minimum TLS version for all connections
- **Certificate Validation**: No insecure certificate skipping
- **Encrypted Transit**: All data encrypted in flight
- **No Plaintext**: No plaintext protocols or credentials

#### Network Policies
```yaml
# Restrictive network policy example
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: rancher-monitoring-relay
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: rancher-monitoring-relay
  policyTypes:
  - Ingress
  - Egress
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443  # HTTPS to Rancher only
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: prometheus
    ports:
    - protocol: TCP
      port: 9000  # Metrics endpoint
```

### 📊 Security Monitoring

#### Built-in Security Metrics
The relay exposes security-relevant metrics via the `/metrics` endpoint:

```prometheus
# Connection security metrics
rancher_api_connection_errors_total
rancher_api_auth_failures_total
rancher_api_certificate_errors_total

# Service proxy security metrics
service_proxy_connection_errors_total
service_proxy_timeout_errors_total
service_proxy_auth_failures_total

# Application security metrics
http_requests_duration_seconds
http_request_size_bytes
http_response_size_bytes
```

#### Security Logging
All security-relevant events are logged:
- Authentication attempts and failures
- Connection errors and timeouts
- Certificate validation issues
- Unusual access patterns
- Configuration changes

## Threat Model

### 🎯 Assets Protected
1. **Rancher API Credentials**: Access keys and secrets
2. **Monitoring Data**: Metrics and logs from remote clusters
3. **Network Access**: Service proxy connectivity
4. **System Resources**: CPU, memory, network

### ⚠️ Threats Mitigated

#### Credential Exposure
- **Threat**: API keys exposed in logs, config, or environment
- **Mitigation**: Kubernetes secrets, no logging of credentials
- **Detection**: GitLeaks, secret scanning in CI/CD

#### Man-in-the-Middle Attacks
- **Threat**: Interception of communication to Rancher
- **Mitigation**: TLS 1.2+, certificate validation
- **Detection**: TLS handshake monitoring

#### Privilege Escalation
- **Threat**: Container breakout or privilege escalation
- **Mitigation**: Non-root user, read-only filesystem, dropped capabilities
- **Detection**: Runtime security monitoring

#### Denial of Service
- **Threat**: Resource exhaustion or service disruption
- **Mitigation**: Resource limits, timeout configurations
- **Detection**: Resource usage monitoring

#### Supply Chain Attacks
- **Threat**: Compromised dependencies or base images
- **Mitigation**: Dependency scanning, SBOM generation
- **Detection**: Vulnerability scanners in CI/CD

## Security Scanning

### 🔍 Automated Security Pipeline

Our CI/CD pipeline includes comprehensive security scanning:

```yaml
# Security scanning in GitHub Actions
- name: Security Scanning
  jobs:
    - GitLeaks (secrets detection)
    - CodeQL (static analysis)
    - Gosec (Go security analyzer)
    - Trivy (vulnerability scanning)
    - Grype (container scanning)
    - Checkov (IaC security)
```

#### Scan Coverage Matrix

| Scan Type | Tool | Frequency | Severity Threshold |
|-----------|------|-----------|-------------------|
| Secret Detection | GitLeaks | Every commit | All |
| Static Analysis | CodeQL | Every commit | Medium+ |
| Go Security | Gosec | Every commit | Medium+ |
| Dependencies | Trivy | Every build | High+ |
| Container Image | Grype | Every build | High+ |
| Infrastructure | Checkov | Every commit | Medium+ |
| License Compliance | FOSSA | Weekly | All |
| SBOM Generation | Anchore | Every build | N/A |

### 🛠️ Local Security Testing

Run security scans locally:

```bash
# Run all security scans
make security

# Individual scans
gosec ./...                              # Go security analyzer
docker run --rm -v $(pwd):/src aquasec/trivy:latest fs /src  # Filesystem scan
docker run --rm -v $(pwd):/src gitleaks/gitleaks:latest detect /src  # Secret scan
```

## Security Configuration

### 🔧 Production Deployment Security

#### Kubernetes Security Context
```yaml
# Recommended security context
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1001
  runAsGroup: 1001
  seccompProfile:
    type: RuntimeDefault
  seLinuxOptions:
    level: "s0:c123,c456"

podSecurityContext:
  fsGroup: 1001
  runAsGroup: 1001
  runAsNonRoot: true
  runAsUser: 1001
  seccompProfile:
    type: RuntimeDefault
  supplementalGroups: []
```

#### Pod Security Standards
The relay is compatible with the **Restricted** Pod Security Standard:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring-system
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

#### Resource Limits
```yaml
# Prevent resource exhaustion attacks
resources:
  limits:
    cpu: 500m
    memory: 512Mi
    ephemeral-storage: 100Mi
  requests:
    cpu: 100m
    memory: 128Mi
    ephemeral-storage: 50Mi
```

### 🔑 Secrets Management

#### Production Secrets Configuration
```yaml
# Use Kubernetes secrets (never inline values)
apiVersion: v1
kind: Secret
metadata:
  name: rancher-api-credentials
  annotations:
    kubernetes.io/managed-by: "external-secrets-operator"
type: Opaque
data:
  access-key: <base64-encoded-access-key>
  secret-key: <base64-encoded-secret-key>
```

#### External Secrets Integration
```yaml
# Example with External Secrets Operator
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "https://vault.company.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "rancher-monitoring-relay"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: rancher-credentials
spec:
  refreshInterval: 15m
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: rancher-api-credentials
  data:
  - secretKey: access-key
    remoteRef:
      key: rancher/monitoring
      property: access-key
  - secretKey: secret-key
    remoteRef:
      key: rancher/monitoring
      property: secret-key
```

### 🛡️ Network Security

#### Service Mesh Integration
```yaml
# Istio service mesh configuration
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: rancher-monitoring-relay
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: rancher-monitoring-relay
  mtls:
    mode: STRICT

---
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: rancher-monitoring-relay
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: rancher-monitoring-relay
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/monitoring/sa/prometheus"]
    to:
    - operation:
        methods: ["GET"]
        paths: ["/metrics", "/health"]
```

#### Ingress Security
```yaml
# Secure ingress configuration
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: rancher-monitoring-relay
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/backend-protocol: "HTTP"
    nginx.ingress.kubernetes.io/whitelist-source-range: "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - monitoring-relay.company.com
    secretName: monitoring-relay-tls
  rules:
  - host: monitoring-relay.company.com
    http:
      paths:
      - path: /metrics
        pathType: Prefix
        backend:
          service:
            name: rancher-monitoring-relay
            port:
              number: 9000
```

## Security Monitoring & Alerting

### 📈 Security Metrics

Monitor these key security metrics:

```prometheus
# Authentication failures
rate(rancher_api_auth_failures_total[5m]) > 0.1

# TLS certificate errors
rate(rancher_api_certificate_errors_total[5m]) > 0

# Unusual connection patterns
rate(http_requests_total[5m]) > 100

# Resource exhaustion
container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.8
```

### 🚨 Security Alerts

```yaml
# Prometheus alerting rules
groups:
- name: rancher-monitoring-relay-security
  rules:
  - alert: AuthenticationFailure
    expr: rate(rancher_api_auth_failures_total[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High authentication failure rate"
      description: "Rancher API authentication failures exceed threshold"

  - alert: TLSCertificateError
    expr: rate(rancher_api_certificate_errors_total[5m]) > 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "TLS certificate validation errors"
      description: "Certificate validation failures detected"

  - alert: UnusualTrafficPattern
    expr: rate(http_requests_total[5m]) > 100
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Unusual traffic pattern detected"
      description: "HTTP request rate is unusually high"

  - alert: ResourceExhaustion
    expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High memory usage"
      description: "Memory usage exceeds 90% of limit"
```

## Incident Response

### 🚨 Security Incident Types

#### Credential Compromise
1. **Immediate Actions:**
   - Rotate affected API keys
   - Review access logs
   - Update secrets in Kubernetes
   - Restart affected pods

2. **Investigation:**
   - Check git history for exposed secrets
   - Review container logs
   - Analyze network traffic
   - Identify scope of compromise

#### Vulnerability Exploit
1. **Immediate Actions:**
   - Isolate affected systems
   - Apply security patches
   - Update container images
   - Review security logs

2. **Recovery:**
   - Deploy patched version
   - Validate security posture
   - Monitor for reoccurrence
   - Document lessons learned

### 📋 Incident Response Checklist

- [ ] Identify and contain the incident
- [ ] Assess the scope and impact
- [ ] Collect forensic evidence
- [ ] Notify relevant stakeholders
- [ ] Apply immediate fixes
- [ ] Monitor for persistence
- [ ] Conduct post-incident review
- [ ] Update security measures

## Compliance

### 📜 Security Standards

The relay is designed to meet these security standards:

- **CIS Kubernetes Benchmark**: Container and orchestration security
- **OWASP Application Security**: Web application security practices
- **NIST Cybersecurity Framework**: Comprehensive security program
- **SOC 2 Type II**: Security and availability controls
- **ISO 27001**: Information security management

### 🏛️ Compliance Features

| Requirement | Implementation | Verification |
|-------------|----------------|--------------|
| Data Encryption | TLS 1.2+ for all connections | Certificate validation |
| Access Control | RBAC + API key authentication | Access logging |
| Audit Logging | Comprehensive security logs | Log analysis |
| Vulnerability Management | Automated scanning in CI/CD | Scan reports |
| Incident Response | Defined procedures and runbooks | Incident exercises |
| Supply Chain Security | SBOM + dependency scanning | Vulnerability reports |

## Security Best Practices

### 🔒 Deployment Best Practices

1. **Use Latest Versions**: Always deploy the latest stable version
2. **Apply Security Contexts**: Use restrictive security contexts
3. **Enable Network Policies**: Implement least-privilege networking
4. **Set Resource Limits**: Prevent resource exhaustion
5. **Use Secrets**: Never put credentials in environment variables
6. **Enable TLS**: Ensure all communication is encrypted
7. **Monitor Continuously**: Implement comprehensive monitoring

### 🛠️ Operational Best Practices

1. **Regular Updates**: Keep dependencies and base images updated
2. **Rotate Credentials**: Regularly rotate API keys and certificates
3. **Backup Configurations**: Maintain secure configuration backups
4. **Test Security**: Regularly test security controls
5. **Train Teams**: Ensure teams understand security procedures
6. **Document Everything**: Maintain up-to-date security documentation

### 🔍 Monitoring Best Practices

1. **Security Metrics**: Monitor security-relevant metrics
2. **Alert Tuning**: Minimize false positives in security alerts
3. **Log Aggregation**: Centralize security logs for analysis
4. **Threat Detection**: Implement automated threat detection
5. **Incident Response**: Have clear incident response procedures

## Security Contacts

- **Security Team**: [security@support.tools](mailto:security@support.tools)
- **Emergency Contact**: Available 24/7 for critical security incidents
- **Bug Bounty**: Responsible disclosure program available
- **Security Advisories**: Subscribe for security updates

## Resources

- [SECURITY.md](../SECURITY.md) - Security policy and vulnerability reporting
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [OWASP Kubernetes Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Kubernetes_Security_Cheat_Sheet.html)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

---

**Last Updated**: August 2025  
**Next Review**: November 2025