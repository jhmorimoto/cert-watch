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
	Echo CertWatcherActionEcho `json:"echo,omitempty"`

	Email CertWatchActionEmail `json:"email,omitempty"`
}

// This actions is used to send certificate files via e-mail. Before sending,
// both private and public keys are saved into a temporary workspace directory
// and converted to various popular formats that can be used as attachments,
// such as PEM and PKCS#12. All files are also zipped to give users the option
// to send zipped files, instead of the raw certificates. There will be one zip
// file for each indidivual certificate format and another with all of them
// together. Zip files can also be password protected. All these options are
// provided to give user multiple options. Quite often, e-mail recipients have
// anti-virus software that scans incoming mail and blocks certain file
// extensions (scripts and certificates included). To overcome these
// restrictions, cert-watch users have the option to send a password protected
// zip file. This password is assumed to be shared secret between sender and
// receiver and is not managed by cert-watch.
type CertWatchActionEmail struct {
	// Indicates whether this action is enabled. Defaults to false.
	Enabled bool `json:"enabled,omitempty"`

	// Configuration file with information about the email server to use
	ConfigFile string `json:"configFile,omitempty"`

	// FROM header to use in the email. If not specified, the value must be
	// specified in configuration file.
	From string `json:"from,omitempty"`

	// TO header to use in the email. A comma separated list of email addresses.
	To string `json:"to"`

	// CC  header to use in the email.
	Cc string `json:"cc,omitempty"`

	// BCC header to use in the email.
	Bcc string `json:"bcc,omitempty"`

	// SUBJECT header to use in the email.
	Subject string `json:"subject,omitempty"`

	// Template file used to generate the email body contents.
	BodyTemplate string `json:"bodyTemplate,omitempty"`

	// Email body content type: text/plain or text/html
	BodyContentType string `json:"bodyContentType,omitempty"`

	// List of attachments to send with the email. Paths are relative to a
	// temporary workspace directory where different versions of the certificate
	// files are saved before sending the email. Files will be available in
	// popular formats, like PEM and PKCS#12, zipped and unzipped.
	Attachments []string `json:"attachments,omitempty"`
}

// Dummy action that simply generates an Event informing the Secret change.
// Does not perform any useful action and is mostly used for testing and debugging.
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

	// Password that should be used to zip certificate files. Zipped versions of
	// each certificates are kept along with the raw files. If this values is
	// empty, zip files will no tbe protected with any password.
	ZipFilesPassword string `json:"zipFilesPassword,omitempty"`

	// Password that should be used in the PKCS#12 envelope. If empty, p12
	// certificate files will not be protected by any password.
	Pkcs12Password string `json:"pkcs12Password,omitempty"`

	// Filename prefix that should be used in the exported certificate files. If
	// empty, defaults to "tls", so files will be created in the temporary
	// workspace directory as tls.key, tls.crt, tls.p12, etc...
	FilenamesPrefix string `json:"filenamesPrefix,omitempty"`

	// Actions that should be performed when the watched Secret changes.
	Actions CertWatcherAction `json:"actions,omitempty"`
}

// CertWatcherStatus defines the observed state of CertWatcher
type CertWatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Status       string      `json:"status,omitempty"`
	LastUpdate   metav1.Time `json:"lastUpdate,omitempty"`
	LastChecksum string      `json:"lastChecksum,omitempty"`
	ActionStatus string      `json:"actionStatus,omitempty"`
	Message      string      `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CertWatcher is the Schema for the certwatchers API
// +kubebuilder:printcolumn:name="SECRET_NS",type=string,JSONPath=`.spec.secret.namespace`
// +kubebuilder:printcolumn:name="SECRET_NAME",type=string,JSONPath=`.spec.secret.name`
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="ACTION_STATUS",type=string,JSONPath=`.status.actionStatus`
// +kubebuilder:printcolumn:name="LAST_UPDATE",type=string,JSONPath=`.status.lastUpdate`
// +kubebuilder:printcolumn:name="LAST_CHECKSUM",type=string,JSONPath=`.status.lastChecksum`
// +kubebuilder:printcolumn:name="MESSAGE",type=string,JSONPath=`.status.message`
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
