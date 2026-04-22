/*
Copyright 2026 Komal Suthar.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"

	certsv1 "github.com/19-komal/certificate-management-operator/api/v1"
)

var (
	// Prometheus metrics
	certificateExpiryDays = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cert_expiry_days_remaining",
			Help: "Number of days until certificate expires",
		},
		[]string{"namespace", "secret", "cn"},
	)

	certificateRotationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cert_rotation_total",
			Help: "Total number of certificate rotation attempts",
		},
		[]string{"namespace", "secret", "status"},
	)

	certificateCheckDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cert_check_duration_seconds",
			Help:    "Time taken to check certificates",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	// Register custom metrics with the global Prometheus registry
	metrics.Registry.MustRegister(
		certificateExpiryDays,
		certificateRotationTotal,
		certificateCheckDuration,
	)
}

// CertificateRotationReconciler reconciles a CertificateRotation object
type CertificateRotationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=certs.openshift.io,resources=certificaterotations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certs.openshift.io,resources=certificaterotations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=certs.openshift.io,resources=certificaterotations/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *CertificateRotationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	startTime := time.Now()

	// Fetch the CertificateRotation instance
	certRotation := &certsv1.CertificateRotation{}
	err := r.Get(ctx, req.NamespacedName, certRotation)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("CertificateRotation resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CertificateRotation")
		return ctrl.Result{}, err
	}

	// Get all secrets with TLS certificates
	secrets, err := r.getTLSSecrets(ctx, certRotation.Spec.Namespaces)
	if err != nil {
		log.Error(err, "Failed to list secrets")
		return ctrl.Result{}, err
	}

	// Process each certificate
	var certStatuses []certsv1.CertificateStatus
	expiringCount := 0
	expiredCount := 0
	rotatedCount := 0

	for _, secret := range secrets.Items {
		status, err := r.checkCertificate(ctx, &secret, certRotation.Spec.ThresholdDays)
		if err != nil {
			log.Error(err, "Failed to check certificate", "secret", secret.Name, "namespace", secret.Namespace)
			continue
		}

		if status != nil {
			certStatuses = append(certStatuses, *status)

			// Update metrics
			certificateExpiryDays.WithLabelValues(
				status.Namespace,
				status.Name,
				"", // CN could be extracted from cert if needed
			).Set(float64(status.DaysUntilExpiry))

			// Count certificates by status
			switch status.Status {
			case "Expiring":
				expiringCount++
			case "Expired":
				expiredCount++
			case "Rotated":
				rotatedCount++
			}

			// Auto-rotate if enabled and certificate is expiring
			if certRotation.Spec.AutoRotate && !certRotation.Spec.AlertOnly &&
				(status.Status == "Expiring" || status.Status == "Expired") {
				err := r.rotateCertificate(ctx, &secret)
				if err != nil {
					log.Error(err, "Failed to rotate certificate", "secret", secret.Name)
					certificateRotationTotal.WithLabelValues(secret.Namespace, secret.Name, "failed").Inc()
				} else {
					certificateRotationTotal.WithLabelValues(secret.Namespace, secret.Name, "success").Inc()
					status.Status = "Rotated"
					rotatedCount++
				}
			}
		}
	}

	// Update status
	now := metav1.Now()
	certRotation.Status.Certificates = certStatuses
	certRotation.Status.TotalCertificates = len(certStatuses)
	certRotation.Status.ExpiringCertificates = expiringCount
	certRotation.Status.ExpiredCertificates = expiredCount
	certRotation.Status.RotatedCertificates = rotatedCount
	certRotation.Status.LastReconcileTime = &now

	// Update conditions
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "ReconcileSuccess",
		Message:            fmt.Sprintf("Successfully checked %d certificates", len(certStatuses)),
		LastTransitionTime: now,
	}

	if expiredCount > 0 {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "CertificatesExpired"
		condition.Message = fmt.Sprintf("%d certificates have expired", expiredCount)
	} else if expiringCount > 0 {
		condition.Status = metav1.ConditionTrue
		condition.Reason = "CertificatesExpiring"
		condition.Message = fmt.Sprintf("%d certificates expiring within %d days", expiringCount, certRotation.Spec.ThresholdDays)
	}

	certRotation.Status.Conditions = []metav1.Condition{condition}

	// Update status
	if err := r.Status().Update(ctx, certRotation); err != nil {
		log.Error(err, "Failed to update CertificateRotation status")
		return ctrl.Result{}, err
	}

	// Record reconciliation duration
	duration := time.Since(startTime).Seconds()
	certificateCheckDuration.Observe(duration)

	log.Info("Reconciliation complete",
		"total", len(certStatuses),
		"expiring", expiringCount,
		"expired", expiredCount,
		"rotated", rotatedCount,
		"duration", duration)

	// Requeue after CheckIntervalMinutes
	requeueAfter := time.Duration(certRotation.Spec.CheckIntervalMinutes) * time.Minute
	if requeueAfter == 0 {
		requeueAfter = 60 * time.Minute // Default to 1 hour
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// getTLSSecrets retrieves all secrets containing TLS certificates
func (r *CertificateRotationReconciler) getTLSSecrets(ctx context.Context, namespaces []string) (*corev1.SecretList, error) {
	secrets := &corev1.SecretList{}

	if len(namespaces) == 0 {
		// Watch all namespaces
		err := r.List(ctx, secrets, client.MatchingFields{"type": string(corev1.SecretTypeTLS)})
		if err != nil {
			// If field selector doesn't work, list all and filter
			err = r.List(ctx, secrets)
			if err != nil {
				return nil, err
			}
			// Filter for TLS secrets
			filtered := &corev1.SecretList{}
			for _, secret := range secrets.Items {
				if secret.Type == corev1.SecretTypeTLS || hasTLSData(secret) {
					filtered.Items = append(filtered.Items, secret)
				}
			}
			return filtered, nil
		}
	} else {
		// Watch specific namespaces
		for _, ns := range namespaces {
			nsList := &corev1.SecretList{}
			err := r.List(ctx, nsList, client.InNamespace(ns))
			if err != nil {
				return nil, err
			}
			for _, secret := range nsList.Items {
				if secret.Type == corev1.SecretTypeTLS || hasTLSData(secret) {
					secrets.Items = append(secrets.Items, secret)
				}
			}
		}
	}

	return secrets, nil
}

// hasTLSData checks if secret contains TLS certificate data
func hasTLSData(secret corev1.Secret) bool {
	_, hasCert := secret.Data["tls.crt"]
	_, hasKey := secret.Data["tls.key"]
	return hasCert && hasKey
}

// checkCertificate parses and checks the certificate expiry
func (r *CertificateRotationReconciler) checkCertificate(ctx context.Context, secret *corev1.Secret, thresholdDays int) (*certsv1.CertificateStatus, error) {
	certData, ok := secret.Data["tls.crt"]
	if !ok {
		return nil, fmt.Errorf("no tls.crt in secret %s/%s", secret.Namespace, secret.Name)
	}

	// Parse PEM certificate
	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate in %s/%s", secret.Namespace, secret.Name)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate in %s/%s: %w", secret.Namespace, secret.Name, err)
	}

	// Calculate days until expiry
	now := time.Now()
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

	// Determine status
	status := "Valid"
	if daysUntilExpiry < 0 {
		status = "Expired"
	} else if daysUntilExpiry <= thresholdDays {
		status = "Expiring"
	}

	return &certsv1.CertificateStatus{
		Name:            secret.Name,
		Namespace:       secret.Namespace,
		ExpiryDate:      metav1.NewTime(cert.NotAfter),
		DaysUntilExpiry: daysUntilExpiry,
		Status:          status,
		LastChecked:     metav1.Now(),
	}, nil
}

// rotateCertificate triggers certificate rotation
func (r *CertificateRotationReconciler) rotateCertificate(ctx context.Context, secret *corev1.Secret) error {
	log := log.FromContext(ctx)

	// Check if cert-manager Certificate resource exists
	// This is a placeholder - in production, you would:
	// 1. Check if the secret is managed by cert-manager (has cert-manager.io/certificate-name annotation)
	// 2. Trigger renewal by deleting the secret (cert-manager will recreate)
	// 3. OR update the Certificate resource to force renewal

	if certName, ok := secret.Annotations["cert-manager.io/certificate-name"]; ok {
		log.Info("Certificate is managed by cert-manager, triggering rotation",
			"secret", secret.Name,
			"namespace", secret.Namespace,
			"certificate", certName)

		// In production, you would patch the Certificate resource or delete the secret
		// For now, we just log it
		log.Info("Auto-rotation not implemented - alertOnly mode recommended")
		return fmt.Errorf("auto-rotation requires cert-manager integration")
	}

	log.Info("Certificate is not managed by cert-manager, cannot auto-rotate",
		"secret", secret.Name,
		"namespace", secret.Namespace)

	return fmt.Errorf("certificate not managed by cert-manager")
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateRotationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certsv1.CertificateRotation{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
