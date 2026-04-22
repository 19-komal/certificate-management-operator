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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertificateRotationSpec defines the desired state of CertificateRotation
type CertificateRotationSpec struct {
	// Namespaces to watch for TLS certificates. Empty means all namespaces.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// ThresholdDays is the number of days before expiry to trigger rotation
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=365
	ThresholdDays int `json:"thresholdDays,omitempty"`

	// AutoRotate enables automatic certificate rotation via cert-manager
	// +kubebuilder:default=false
	// +optional
	AutoRotate bool `json:"autoRotate,omitempty"`

	// AlertOnly if true, only alerts but doesn't rotate certificates
	// +kubebuilder:default=true
	// +optional
	AlertOnly bool `json:"alertOnly,omitempty"`

	// CheckIntervalMinutes defines how often to check certificates
	// +kubebuilder:default=60
	// +kubebuilder:validation:Minimum=5
	// +kubebuilder:validation:Maximum=1440
	CheckIntervalMinutes int `json:"checkIntervalMinutes,omitempty"`
}

// CertificateStatus represents the status of a single certificate
type CertificateStatus struct {
	// Name of the secret containing the certificate
	Name string `json:"name"`

	// Namespace of the secret
	Namespace string `json:"namespace"`

	// ExpiryDate of the certificate
	ExpiryDate metav1.Time `json:"expiryDate"`

	// DaysUntilExpiry calculated days until certificate expires
	DaysUntilExpiry int `json:"daysUntilExpiry"`

	// Status of the certificate (Valid, Expiring, Expired, Rotated, Error)
	Status string `json:"status"`

	// LastChecked timestamp
	LastChecked metav1.Time `json:"lastChecked"`
}

// CertificateRotationStatus defines the observed state of CertificateRotation
type CertificateRotationStatus struct {
	// Certificates contains the status of all monitored certificates
	// +optional
	Certificates []CertificateStatus `json:"certificates,omitempty"`

	// TotalCertificates is the total number of certificates being monitored
	TotalCertificates int `json:"totalCertificates"`

	// ExpiringCertificates is the number of certificates expiring within threshold
	ExpiringCertificates int `json:"expiringCertificates"`

	// ExpiredCertificates is the number of already expired certificates
	ExpiredCertificates int `json:"expiredCertificates"`

	// RotatedCertificates is the number of successfully rotated certificates
	RotatedCertificates int `json:"rotatedCertificates"`

	// LastReconcileTime is the timestamp of the last reconciliation
	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`

	// Conditions represent the latest available observations of the CertificateRotation's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=certrot
// +kubebuilder:printcolumn:name="Threshold",type="integer",JSONPath=".spec.thresholdDays"
// +kubebuilder:printcolumn:name="Total",type="integer",JSONPath=".status.totalCertificates"
// +kubebuilder:printcolumn:name="Expiring",type="integer",JSONPath=".status.expiringCertificates"
// +kubebuilder:printcolumn:name="Expired",type="integer",JSONPath=".status.expiredCertificates"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CertificateRotation is the Schema for the certificaterotations API
type CertificateRotation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateRotationSpec   `json:"spec,omitempty"`
	Status CertificateRotationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateRotationList contains a list of CertificateRotation
type CertificateRotationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificateRotation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertificateRotation{}, &CertificateRotationList{})
}
