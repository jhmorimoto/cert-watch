package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CertWatcherSecret struct {
	// Name of the Secret watched by CertWatcher
	Name string `json:"name"`

	// Namespace of the Secret watched by CertWatcher.
	Namespace string `json:"namespace"`
}

type CertWatcherAction struct {
	// Dummy action that simply outputs the secret name into the controller logs.
	Echo CertWatcherActionEcho `json:"echo,omitempty"`
}

// CertWatcherActionEcho has no meaningful prameters. It can only be enabled or disabled.
type CertWatcherActionEcho struct {
	// Indicates whether this action is enabled. Defaults to false.
	Enabled bool `json:"enabled,omitempty"`
}

// CertWatcherSpec defines the desired state of CertWatcher
type CertWatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Secret watched by CertWatcher
	Secret CertWatcherSecret `json:"secret"`

	Actions CertWatcherAction `json:"actions,omitempty"`
}

// CertWatcherStatus defines the observed state of CertWatcher
type CertWatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	LastUpdate   metav1.Time `json:"lastUpdate,omitempty"`
	LastChecksum string      `json:"lastChecksum,omitempty"`
	Checksum     string      `json:"checksum,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CertWatcher is the Schema for the certwatchers API
type CertWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertWatcherSpec   `json:"spec,omitempty"`
	Status CertWatcherStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CertWatcherList contains a list of CertWatcher
type CertWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertWatcher{}, &CertWatcherList{})
}
