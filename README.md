# Certificate Management Operator

Kubernetes Operator built with Golang that automates TLS certificate rotation for OpenShift clusters.

## What It Does

- ✅ Monitors TLS certificate expiry across the cluster
- ✅ Automates certificate rotation before expiration
- ✅ Integrates with cert-manager for certificate lifecycle
- ✅ Exposes Prometheus metrics for monitoring
- ✅ Custom Resource Definition (CRD) for certificate management

## Why I Built This

As an OpenShift support engineer, I saw many customer incidents caused by expired certificates. This operator automates certificate rotation to prevent outages.

## Architecture

```
certificate-management-operator/
├── main.go                                    # Operator entry point
├── api/v1/
│   ├── certificaterotation_types.go          # CRD definition
│   └── groupversion_info.go                  # API version info
├── controllers/
│   └── certificaterotation_controller.go     # Reconciliation logic
├── config/
│   ├── crd/                                   # CRD manifests
│   ├── rbac/                                  # RBAC permissions
│   ├── prometheus/                            # Metrics & alerts
│   └── samples/                               # Example CRs
├── go.mod                                     # Go dependencies
└── go.sum                                     # Dependency checksums
```

## How It Works

1. **Custom Resource**: Defines `CertificateRotation` CRD with cert-manager integration
2. **Controller**: Watches certificates, checks expiry, triggers rotation
3. **Reconciliation Loop**: Kubernetes operator pattern (watch → reconcile → update)
4. **Prometheus Metrics**: Exposes certificate expiry time, rotation events
5. **Alerting**: PrometheusRule triggers alerts for certificates expiring soon

## Technology Stack

- **Language**: Golang
- **Framework**: Operator SDK / Kubebuilder
- **CRD**: Custom Resource Definition for certificate specs
- **Monitoring**: Prometheus metrics, ServiceMonitor, PrometheusRule
- **Integration**: cert-manager for certificate lifecycle

## Prerequisites

- Kubernetes cluster (or OpenShift)
- cert-manager installed
- Operator SDK (for development)

## Installation

```bash
# Apply CRD
kubectl apply -f config/crd/bases/certs.openshift.io_certificaterotations.yaml

# Apply RBAC
kubectl apply -f config/rbac/

# Build and run operator
go build -o bin/manager main.go
./bin/manager
```

## Example Usage

Create a CertificateRotation resource:

```yaml
apiVersion: certs.openshift.io/v1
kind: CertificateRotation
metadata:
  name: api-server-cert
spec:
  secretName: api-server-tls
  namespace: openshift-config
  expiryThreshold: 720h  # 30 days
  issuerRef:
    name: cluster-ca
    kind: ClusterIssuer
```

Check status:

```bash
kubectl get certificaterotation api-server-cert -o yaml
```

## Prometheus Metrics

Exposed metrics:
- `cert_expiry_timestamp_seconds` - Certificate expiration time
- `cert_rotation_total` - Total number of certificate rotations
- `cert_rotation_errors_total` - Failed rotation attempts

## Key Concepts Demonstrated

### 1. Kubernetes Operator Pattern
- Custom Resource Definition (CRD)
- Controller reconciliation loop
- Watch-reconcile-update pattern

### 2. Golang Programming
- Struct tags for Kubernetes API objects
- Kubernetes client-go library usage
- Error handling and logging

### 3. Operator SDK
- Kubebuilder markers for RBAC
- Controller runtime
- Webhook integration (optional)

### 4. Observability
- Prometheus metrics export
- ServiceMonitor for scraping
- PrometheusRule for alerting

## Code Overview

### main.go (~106 lines)
Operator entry point - initializes manager, registers controllers, starts metrics server.

### certificaterotation_controller.go (~341 lines)
Core reconciliation logic:
- Watches CertificateRotation CRs
- Checks certificate expiry via cert-manager
- Triggers rotation when threshold reached
- Updates status with last rotation time

### certificaterotation_types.go (~129 lines)
CRD type definitions:
- `CertificateRotationSpec` - desired state
- `CertificateRotationStatus` - observed state
- Kubebuilder markers for validation

## Development Experience

This was my first Kubernetes Operator built from scratch using Operator SDK. Key learnings:

1. **Reconciliation Loop**: Understanding eventual consistency vs immediate updates
2. **RBAC**: Proper permissions for reading Secrets and updating CRs
3. **Testing**: Writing controller tests with fake Kubernetes clients
4. **Metrics**: Integrating Prometheus metrics into operator logic

## Time Savings

- **Manual certificate rotation**: ~2 hours per certificate (checking expiry, generating new cert, applying, validating)
- **With this operator**: Automated (0 manual intervention)
- **Impact**: Prevents outages from expired certificates

## Author

Built by **Komal Suthar** - Red Hat OpenShift Support Engineer

Demonstrates: Golang, Kubernetes Operators, Operator SDK, CRDs, Prometheus

## Contributing

Contributions welcome! This is a learning project showcasing operator development patterns.
