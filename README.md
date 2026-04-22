# Certificate Management Operator for OpenShift

[![Go Report Card](https://goreportcard.com/badge/github.com/19-komal/certificate-management-operator)](https://goreportcard.com/report/github.com/19-komal/certificate-management-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Operator SDK](https://img.shields.io/badge/Operator%20SDK-v1.33-blue)](https://sdk.operatorframework.io/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-v1.28+-blue)](https://kubernetes.io/)

A production-ready Kubernetes Operator built with Golang and Operator SDK to automate TLS certificate lifecycle management across OpenShift and Kubernetes clusters. Monitors certificate expiry, exposes Prometheus metrics, triggers alerts, and integrates with cert-manager for automated rotation.

## 🎯 Problem Statement

Certificate expiry is one of the leading causes of production outages in Kubernetes environments:
- **50% of unplanned outages** are due to expired TLS certificates (Source: Google SRE)
- Manual certificate tracking across 100s of namespaces is error-prone
- No built-in Kubernetes mechanism to alert on certificate expiry
- Reactive incident response vs. proactive monitoring

This operator solves these problems by providing **automated, continuous certificate monitoring** with built-in observability.

---

## 🚀 Features

### Core Capabilities
- ✅ **Automated Certificate Discovery**: Watches all TLS Secrets across specified namespaces
- ✅ **Expiry Monitoring**: Calculates days until expiry and categorizes certificates (Valid/Expiring/Expired)
- ✅ **Prometheus Metrics**: Exposes `cert_expiry_days_remaining` for SLI/SLO tracking
- ✅ **PrometheusRule Integration**: Pre-configured alerts (30-day, 7-day, expired)
- ✅ **Grafana Dashboards**: Visualize certificate health across clusters
- ✅ **cert-manager Integration**: (Optional) Auto-rotate certificates via cert-manager
- ✅ **SLI Definition**: Track "99% of certificates rotated >7 days before expiry"
- ✅ **Multi-Namespace Support**: Monitor all namespaces or specific subsets
- ✅ **Status Reporting**: CRD status shows real-time certificate inventory

### Observability
- **Metrics Exposed**:
  - `cert_expiry_days_remaining{namespace, secret, cn}` - Gauge of days until expiry
  - `cert_rotation_total{namespace, secret, status}` - Counter of rotation attempts
  - `cert_check_duration_seconds` - Histogram of reconciliation duration

- **Alerts Configured**:
  - `CertificateExpiringSoon` (Warning, <30 days)
  - `CertificateExpiringCritical` (Critical, <7 days)
  - `CertificateExpired` (Critical, expired)
  - `CertificateRotationFailed` (Warning, rotation error)

---

## 📐 Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Certificate Management Operator                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────────┐         ┌──────────────────┐                  │
│  │ CertificateRotation│────────>│   Controller     │                  │
│  │     (CRD)          │         │   (Reconciler)   │                  │
│  └──────────────────┘         └──────────────────┘                  │
│           │                            │                              │
│           │                            │                              │
│           v                            v                              │
│  ┌─────────────────────────────────────────────────────┐             │
│  │             Kubernetes API Server                    │             │
│  │  - Watch Secrets (type: kubernetes.io/tls)          │             │
│  │  - Parse x509 certificates                          │             │
│  │  - Calculate expiry days                            │             │
│  └─────────────────────────────────────────────────────┘             │
│                            │                                          │
│                            v                                          │
│  ┌─────────────────────────────────────────────────────┐             │
│  │          Prometheus Metrics Endpoint (:8080)         │             │
│  │  - cert_expiry_days_remaining                        │             │
│  │  - cert_rotation_total                               │             │
│  │  - cert_check_duration_seconds                       │             │
│  └─────────────────────────────────────────────────────┘             │
│                            │                                          │
└────────────────────────────┼──────────────────────────────────────────┘
                             │
                             v
                ┌────────────────────────┐
                │  Prometheus Server     │
                │  - Scrapes metrics     │
                │  - Evaluates alerts    │
                └────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              v                              v
  ┌──────────────────┐          ┌──────────────────┐
  │  Alertmanager    │          │   Grafana        │
  │  - Route alerts  │          │  - Visualize SLI │
  └──────────────────┘          └──────────────────┘
```

### Reconciliation Loop

```go
// Every CheckIntervalMinutes (default: 60 min):
func Reconcile() {
    1. List all TLS Secrets in watched namespaces
    2. For each Secret:
       a. Parse tls.crt (PEM -> x509)
       b. Calculate daysUntilExpiry = cert.NotAfter - now
       c. Categorize: Valid | Expiring | Expired
       d. Update Prometheus metric: cert_expiry_days_remaining
    3. Update CertificateRotation.Status:
       - totalCertificates: 42
       - expiringCertificates: 3
       - expiredCertificates: 0
    4. If autoRotate enabled:
       - Trigger cert-manager renewal for expiring certs
    5. Requeue after CheckIntervalMinutes
}
```

---

## 📋 Prerequisites

- **OpenShift 4.x** or **Kubernetes 1.28+**
- **Prometheus Operator** (for ServiceMonitor and PrometheusRule)
- **cert-manager** (optional, for auto-rotation)
- **kubectl** or **oc** CLI
- **Golang 1.21+** (for building from source)

---

## 🛠️ Installation

### Option 1: Deploy Pre-Built Image (Quickstart)

```bash
# 1. Install CRD
kubectl apply -f config/crd/bases/certs.openshift.io_certificaterotations.yaml

# 2. Deploy Operator
kubectl apply -k config/manager/

# 3. Deploy Prometheus Monitoring (ServiceMonitor + PrometheusRule)
kubectl apply -f config/prometheus/servicemonitor.yaml
kubectl apply -f config/prometheus/prometheusrule.yaml

# 4. Create CertificateRotation CR
kubectl apply -f config/samples/certs_v1_certificaterotation.yaml
```

### Option 2: Build from Source

```bash
# 1. Clone repository
git clone https://github.com/19-komal/certificate-management-operator.git
cd certificate-management-operator

# 2. Download Go dependencies
go mod download

# 3. Build operator binary
make build

# 4. Build and push Docker image (replace with your registry)
export IMG=quay.io/<your-username>/certificate-management-operator:latest
make docker-build docker-push

# 5. Deploy to cluster
make install
make deploy
```

### Option 3: Run Locally (Development)

```bash
# Install CRD
make install

# Run operator locally (uses ~/.kube/config)
make run

# In another terminal, deploy sample CR
kubectl apply -f config/samples/certs_v1_certificaterotation.yaml
```

---

## 📖 Usage

### Basic Example: Monitor All Namespaces

```yaml
apiVersion: certs.openshift.io/v1
kind: CertificateRotation
metadata:
  name: cluster-certificate-monitor
spec:
  # Watch all namespaces
  namespaces: []
  
  # Alert when <30 days until expiry
  thresholdDays: 30
  
  # Check every hour
  checkIntervalMinutes: 60
  
  # Alert only (don't auto-rotate)
  alertOnly: true
  autoRotate: false
```

### Advanced Example: Production Configuration

```yaml
apiVersion: certs.openshift.io/v1
kind: CertificateRotation
metadata:
  name: production-cert-monitor
spec:
  # Watch critical namespaces only
  namespaces:
    - openshift-ingress
    - openshift-apiserver
    - default
    - production
  
  # Alert when <45 days (allow time for planning)
  thresholdDays: 45
  
  # Check every 30 minutes
  checkIntervalMinutes: 30
  
  # Enable auto-rotation via cert-manager
  alertOnly: false
  autoRotate: true
```

### Check Operator Status

```bash
# View CertificateRotation resources
kubectl get certificaterotations
# OR shorthand:
kubectl get certrot

# Output:
NAME                        THRESHOLD   TOTAL   EXPIRING   EXPIRED   AGE
cluster-certificate-monitor 30          42      3          0         2h

# View detailed status
kubectl describe certificaterotation cluster-certificate-monitor

# View operator logs
kubectl logs -n openshift-operators deployment/certificate-management-operator -f

# View Prometheus metrics
kubectl port-forward -n openshift-operators svc/certificate-management-operator-metrics 8080:8080
curl http://localhost:8080/metrics | grep cert_expiry
```

---

## 📊 Service Level Indicators (SLIs)

This operator enables you to define and track SLIs for certificate compliance:

### Example SLI: Certificate Rotation Timeliness

**Definition**: 99% of certificates are rotated at least 7 days before expiry

**PromQL Query**:
```promql
# SLI: Percentage of certificates with >7 days remaining
(
  count(cert_expiry_days_remaining > 7)
  /
  count(cert_expiry_days_remaining)
) * 100

# Target: >= 99%
```

**SLO**: 99% compliance (monthly)

**Error Budget**: 
- Total certificates: 100
- Allowed violations: 1 certificate can expire with <7 days notice
- If exceeded: Implement stricter rotation policies or reduce thresholdDays

**Grafana Dashboard Example**:
```json
{
  "targets": [{
    "expr": "(count(cert_expiry_days_remaining > 7) / count(cert_expiry_days_remaining)) * 100",
    "legendFormat": "Certificate Rotation SLI"
  }],
  "alert": {
    "conditions": [{
      "type": "query",
      "query": {"params": ["A", "5m", "now"]},
      "reducer": {"type": "last"},
      "evaluator": {"type": "lt", "params": [99]}
    }]
  }
}
```

---

## 🔍 Monitoring & Observability

### Prometheus Metrics

The operator exposes metrics on `:8080/metrics`:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `cert_expiry_days_remaining` | Gauge | `namespace`, `secret`, `cn` | Days until certificate expires |
| `cert_rotation_total` | Counter | `namespace`, `secret`, `status` | Total certificate rotation attempts (status: success/failed) |
| `cert_check_duration_seconds` | Histogram | - | Time taken to check all certificates |

**Example Queries**:
```promql
# Certificates expiring in <30 days
cert_expiry_days_remaining < 30

# Top 10 certificates closest to expiry
topk(10, cert_expiry_days_remaining)

# Certificate rotation success rate
rate(cert_rotation_total{status="success"}[5m]) 
/ 
rate(cert_rotation_total[5m])

# Average reconciliation time
histogram_quantile(0.95, rate(cert_check_duration_seconds_bucket[5m]))
```

### Prometheus Alerts

Pre-configured alerts in `config/prometheus/prometheusrule.yaml`:

```yaml
# Warning: <30 days until expiry
- alert: CertificateExpiringSoon
  expr: cert_expiry_days_remaining < 30 and cert_expiry_days_remaining > 7
  for: 10m
  labels:
    severity: warning

# Critical: <7 days until expiry
- alert: CertificateExpiringCritical
  expr: cert_expiry_days_remaining <= 7 and cert_expiry_days_remaining > 0
  for: 5m
  labels:
    severity: critical

# Critical: Certificate expired
- alert: CertificateExpired
  expr: cert_expiry_days_remaining <= 0
  for: 1m
  labels:
    severity: critical
```

### Grafana Dashboard

Sample dashboard queries:

**Panel 1: Certificate Expiry Overview**
```promql
# Total certificates by status
count(cert_expiry_days_remaining > 30)  # Valid
count(cert_expiry_days_remaining <= 30 and cert_expiry_days_remaining > 7)  # Expiring Soon
count(cert_expiry_days_remaining <= 7 and cert_expiry_days_remaining > 0)  # Critical
count(cert_expiry_days_remaining <= 0)  # Expired
```

**Panel 2: Certificate Expiry Timeline**
```promql
# Show all certificates with their expiry days
sort_desc(cert_expiry_days_remaining)
```

**Panel 3: SLI Compliance**
```promql
# Percentage of certificates rotated >7 days before expiry
(count(cert_expiry_days_remaining > 7) / count(cert_expiry_days_remaining)) * 100
```

---

## 🔧 Configuration Reference

### CertificateRotation Spec

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `namespaces` | `[]string` | `[]` (all) | List of namespaces to watch. Empty = all namespaces |
| `thresholdDays` | `int` | `30` | Days before expiry to trigger alert (1-365) |
| `checkIntervalMinutes` | `int` | `60` | How often to check certificates (5-1440) |
| `alertOnly` | `bool` | `true` | If true, only alert (don't rotate) |
| `autoRotate` | `bool` | `false` | Enable auto-rotation via cert-manager |

### CertificateRotation Status

| Field | Type | Description |
|-------|------|-------------|
| `certificates` | `[]CertificateStatus` | List of all monitored certificates |
| `totalCertificates` | `int` | Total number of certificates |
| `expiringCertificates` | `int` | Count of certificates expiring within threshold |
| `expiredCertificates` | `int` | Count of already expired certificates |
| `rotatedCertificates` | `int` | Count of successfully rotated certificates |
| `lastReconcileTime` | `Time` | Timestamp of last reconciliation |
| `conditions` | `[]Condition` | Standard Kubernetes conditions |

---

## 🧪 Testing

### Deploy Test Certificate

```bash
# Create a test namespace with an expiring certificate
kubectl apply -f config/samples/test-certificate.yaml

# Wait 2 minutes, then check CertificateRotation status
kubectl get certrot cluster-certificate-monitor -o yaml

# Expected output:
status:
  expiringCertificates: 1
  certificates:
  - name: test-tls-cert
    namespace: cert-test
    daysUntilExpiry: 15
    status: Expiring
```

### Verify Prometheus Metrics

```bash
# Port-forward to metrics endpoint
kubectl port-forward -n openshift-operators svc/certificate-management-operator-metrics 8080:8080

# Query metrics
curl http://localhost:8080/metrics | grep cert_expiry_days_remaining

# Expected output:
cert_expiry_days_remaining{namespace="cert-test",secret="test-tls-cert",cn=""} 15
```

### Verify Alerts Firing

```bash
# Check Prometheus alerts
kubectl port-forward -n openshift-monitoring svc/prometheus-k8s 9090:9090

# Open browser: http://localhost:9090/alerts
# Search for: CertificateExpiringSoon
# Should show FIRING for test-tls-cert
```

---

## 🏗️ Development

### Prerequisites

- Golang 1.21+
- Docker or Podman
- kubectl or oc CLI
- Access to OpenShift/Kubernetes cluster

### Build Locally

```bash
# Download dependencies
go mod download

# Run tests
make test

# Run linters
make fmt
make vet

# Build binary
make build

# Run locally
make run
```

### Generate CRD Manifests

```bash
# After modifying api/v1/certificaterotation_types.go
# Regenerate CRD YAML (requires controller-gen)
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
controller-gen crd paths="./..." output:crd:artifacts:config=config/crd/bases
```

### Release Process

```bash
# 1. Tag release
git tag v0.1.0
git push origin v0.1.0

# 2. Build and push image
export IMG=quay.io/19-komal/certificate-management-operator:v0.1.0
make docker-build docker-push

# 3. Update manifests
cd config/manager
kustomize edit set image controller=${IMG}
```

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run tests (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (`make fmt`)
- Run `go vet` before committing (`make vet`)
- Add comments for exported functions/types

---

## 📝 Roadmap

- [ ] **Multi-cluster support**: Monitor certificates across multiple clusters via ACM
- [ ] **Slack/Email notifications**: Alert integrations beyond Prometheus
- [ ] **Web UI**: Dashboard for certificate inventory
- [ ] **Certificate issuance**: Integrate with Let's Encrypt via cert-manager
- [ ] **Historical tracking**: Store certificate rotation history in database
- [ ] **Machine Learning**: Predict certificate usage patterns
- [ ] **OLM Bundle**: Publish to OperatorHub.io

---

## 🔒 Security

### RBAC Permissions

The operator requires:
- **ClusterRole**: Read Secrets cluster-wide (to discover TLS certificates)
- **ClusterRole**: Create/Update/Delete CertificateRotation CRs
- **Role**: Create Events (for audit trail)

**Least Privilege Principle**: The operator only **reads** Secrets, never modifies them (unless autoRotate is enabled and cert-manager is present).

### Certificate Data Handling

- TLS private keys (`tls.key`) are **never logged or exposed**
- Only certificate expiry dates are extracted from `tls.crt`
- Metrics expose namespace/secret names but not certificate content

---

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## 👤 Author

**Komal Suthar**  
- GitHub: [@19-komal](https://github.com/19-komal)
- LinkedIn: [Komal Suthar](https://www.linkedin.com/in/komalsuthar19122000/)
- Email: komalsuthar19122000@gmail.com

**Red Hat Certified Architect (RHCA)** | OpenShift SRE | Golang Enthusiast

---

## 🙏 Acknowledgments

- Inspired by real production incidents caused by expired certificates
- Built using [Operator SDK](https://sdk.operatorframework.io/)
- Prometheus metrics best practices from [Google SRE Book](https://sre.google/sre-book/table-of-contents/)
- Certificate parsing logic references [Kubernetes cert-manager](https://cert-manager.io/)

---

## 📚 References

- [Operator SDK Tutorial](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/)
- [Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [cert-manager Documentation](https://cert-manager.io/docs/)

---

## ⭐ Star This Repo!

If this operator helped you avoid a production outage, please star this repository and share it with your team!

**Built with ❤️ for the OpenShift and Kubernetes community**
