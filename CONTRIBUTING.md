# Contributing to Certificate Management Operator

Thank you for your interest in contributing to the Certificate Management Operator! This document provides guidelines for contributing to the project.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/0/code_of_conduct/). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs. **actual behavior**
- **Environment details** (OpenShift/Kubernetes version, operator version)
- **Logs** from the operator (`kubectl logs -n openshift-operators deployment/certificate-management-operator`)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use case**: What problem does this solve?
- **Proposed solution**: How should it work?
- **Alternatives considered**: What other approaches did you think about?

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes**:
   - Add tests for new functionality
   - Update documentation if needed
   - Follow the existing code style
3. **Run tests**:
   ```bash
   make test
   make fmt
   make vet
   ```
4. **Commit your changes**:
   - Use clear, descriptive commit messages
   - Reference issue numbers if applicable
   ```
   Add certificate rotation metrics (#42)
   
   - Added cert_rotation_total counter
   - Updated Prometheus documentation
   - Added unit tests for rotation logic
   
   Fixes #42
   ```
5. **Push to your fork** and submit a pull request
6. **Wait for review**: Maintainers will review your PR and may request changes

## Development Setup

### Prerequisites

- Golang 1.21+
- Docker or Podman
- kubectl or oc CLI
- Access to OpenShift/Kubernetes cluster (or CodeReady Containers)

### Local Development

```bash
# Clone your fork
git clone https://github.com/<your-username>/certificate-management-operator.git
cd certificate-management-operator

# Download dependencies
go mod download

# Install CRD
make install

# Run operator locally
make run

# In another terminal, deploy sample
kubectl apply -f config/samples/certs_v1_certificaterotation.yaml
```

### Running Tests

```bash
# Run unit tests
make test

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Code Style

### Golang

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Run `go vet` and fix all warnings
- Add comments for exported functions and types
- Keep functions small and focused

### Example

```go
// Good
// checkCertificate parses and validates the TLS certificate in the given secret.
// Returns the certificate status or an error if parsing fails.
func (r *CertificateRotationReconciler) checkCertificate(ctx context.Context, secret *corev1.Secret, thresholdDays int) (*certsv1.CertificateStatus, error) {
    // Implementation
}

// Bad (no documentation, unclear naming)
func (r *CertificateRotationReconciler) check(s *corev1.Secret, t int) (*certsv1.CertificateStatus, error) {
    // Implementation
}
```

### Kubernetes Resources

- Use YAML for all Kubernetes manifests
- Include comments explaining non-obvious configurations
- Follow [Kubernetes resource best practices](https://kubernetes.io/docs/concepts/configuration/overview/)

## Documentation

### README Updates

If your change affects user-facing functionality:
- Update the relevant section in `README.md`
- Add examples if introducing new features
- Update the configuration reference table

### Code Comments

- Document **why**, not **what** (code shows what it does)
- Explain non-obvious decisions or workarounds
- Reference issues or discussions if applicable

### Example

```go
// Good
// We use a 5-minute timeout here because cert-manager rotation
// can take 2-3 minutes in production clusters with slow storage.
// See issue #123 for details.
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

// Bad
// Set timeout to 5 minutes
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
```

## Testing

### Unit Tests

All new code should include unit tests. Aim for >80% coverage.

```go
func TestCheckCertificate(t *testing.T) {
    tests := []struct {
        name          string
        secret        *corev1.Secret
        thresholdDays int
        wantStatus    string
        wantErr       bool
    }{
        {
            name:          "valid certificate",
            secret:        testValidSecret(),
            thresholdDays: 30,
            wantStatus:    "Valid",
            wantErr:       false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Integration Tests

For larger features, consider adding integration tests that deploy to a real cluster.

## Release Process

Releases are managed by maintainers. The process:

1. Update version in relevant files
2. Update `CHANGELOG.md`
3. Create and push tag: `git tag v0.2.0 && git push origin v0.2.0`
4. Build and push Docker image
5. Create GitHub release with changelog

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Open a GitHub Issue
- **Security issues**: Email komalsuthar19122000@gmail.com (do not open public issue)

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- Release notes for significant contributions
- `CONTRIBUTORS.md` file (if added)

Thank you for contributing! 🎉
