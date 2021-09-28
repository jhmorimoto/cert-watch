package v1

import (
	v1 "k8s.io/api/batch/v1"
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

// CertWatcherAction represents one or more actions that will be performed when a
// Secret change is identified.
type CertWatcherAction struct {
	// Dummy action used for testing and debugging.
	Echo *CertWatcherActionEcho `json:"echo,omitempty"`

	// React to Secret change by sending e-mails.
	Email *CertWatchActionEmail `json:"email,omitempty"`

	// React to Secret change by copying files to a remote host via SCP (ssh).
	Scp *CertWatchActionScp `json:"scp,omitempty"`

	// React to Secret change by running a custom Kubernetes Job. Follow the same spec from batch/v1 API.
	Job *CertWatchActionJob `json:"job,omitempty"`
}

// CertWatchActionJob is used to perform actions upon certificate change by
// running a Kubernetes Job. The job spec follows the same declaration from the
// batch/v1 api. https://kubernetes.io/docs/concepts/workloads/controllers/job/
type CertWatchActionJob struct {
	// Name identifies the job that will be executed.
	Name string `json:"name"`

	// VolumeName controls the name of the volume that will be created to mount
	// certificate files into the Job's containers. Defaults to "certs".
	VolumeName string `json:"volumeName,omitempty"`

	// MountPath controls the mountPath used in the volume created to mount
	// certificate files into the Job's containers. Defaults to "/workspace".
	MountPath string `json:"mountPath,omitempty"`

	// Spec is a standard Kubernetes job spec.
	Spec v1.JobSpec `json:"spec"`
}

// CertWatchActionScp is used to send certificate files via SCP (ssh copy).
// Authentication credentials are recovered from a given Secret name.
// Authentication type (AuthType) can be either `password` (for username and
// password) or `key` for SSH keys.
type CertWatchActionScp struct {
	// Hostname is the remote hostname to connect to.
	Hostname string `json:"hostname"`

	// Port number to connect to. Defaults to 22.
	Port int `json:"port,omitempty"`

	// CredentialSecret is the name of the Secret containing credentials to authenticate. Depending on
	// AuthType, it may contain username, password, key or passphrase values.
	// The reference to the Secret should be in the form namespace/secret-name.
	CredentialSecret string `json:"credentialSecret"`

	// AuthType is the authentication type to use: password|key. Defaults to `password`.
	AuthType string `json:"authType,omitempty"`

	// Files is the list of files to copy. Filenames are relative to a temporary
	// workspace where certificates are stored while they are being processed. After
	// processing, this temporary directory and all its files are removed.
	Files []CertWatchScpFile `json:"files"`
}

// CertWatchScpFile represents a file that must be copied to a remote location
// using the CertWatchActionScp action. Mode defaults to 0600.
type CertWatchScpFile struct {
	// Name is the name of the local certificate file. Filenames are relative to the
	// temporary workspace directory.
	Name string `json:"name"`

	// RemotePath is the full directory path in the remote host where the certificate
	// will be copied to.
	RemotePath string `json:"remotePath"`

	// Mode is the file mode the file on the remote host will have. A string in
	// numeric form, such as 0644.
	Mode string `json:"mode,omitempty"`
}

// CertWatchActionEmail is used to send certificate files via e-mail.
// Before sending, both private and public keys are saved into a temporary
// workspace directory and converted to various popular formats that can be used
// as attachments, such as PEM and PKCS#12. All files are also zipped to give
// users the option to send zipped files, instead of the raw certificates. There
// will be one zip file for each individual certificate format and another with
// all of them together. Zip files can also be password protected. All these
// options are provided to give user multiple options. Quite often, e-mail
// recipients have anti-virus software that scans incoming mail and blocks
// certain file extensions (scripts and certificates included). To overcome these
// restrictions, cert-watch users have the option to send a password-protected
// zip file. This password is assumed to be shared secret between sender and
// receiver and is not managed by cert-watch.
type CertWatchActionEmail struct {
	// ConfigFile is the configuration file with information about the email server
	// to use
	ConfigFile string `json:"configFile,omitempty"`

	// From is the header that identifies the sender of the e-mail. If not specified
	// here, the value must be specified in configuration file.
	From string `json:"from,omitempty"`

	// To is the header that identifies the recipients of the e-mail. A comma
	// separated list of e-mail addresses.
	To string `json:"to"`

	// Cc is the header that identifies carbon copy receivers of the e-mail. A comma
	// separated list of e-mail addresses.
	Cc string `json:"cc,omitempty"`

	// Bcc is the header that identifies blind carbon copy receivers of the e-mail. A
	// comma separated list of e-mail addresses.
	Bcc string `json:"bcc,omitempty"`

	// Subject is the header that informs the subject of the e-mail.
	Subject string `json:"subject,omitempty"`

	// BodyTemplate is the full contents of the e-mail body to send.
	BodyTemplate string `json:"bodyTemplate,omitempty"`

	// BodyContentType is the header that identifies the type of content the e-mail
	// will have: text/plain or text/html
	BodyContentType string `json:"bodyContentType,omitempty"`

	// Attachments is the list of attachments to send with the e-mail. Paths are
	// relative to a temporary workspace directory where different versions of the
	// certificate files are saved before sending the email. Files will be available
	// in popular formats, like PEM and PKCS#12, zipped and unzipped.
	Attachments []string `json:"attachments,omitempty"`
}

// CertWatcherActionEcho Dummy action that simply generates an Event informing
// the Secret change. Does not perform any useful action and is mostly used for
// testing and debugging.
type CertWatcherActionEcho struct {
}

// CertWatcherSpec defines the desired state of CertWatcher
type CertWatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Secret watched by CertWatcher
	Secret CertWatcherSecret `json:"secret"`

	// ZipFilesPassword is the password that should be used to zip certificate files.
	// Zipped versions of each certificates are kept along with the raw files. If
	// this values is empty, zip files will no tbe protected with any password.
	ZipFilesPassword string `json:"zipFilesPassword,omitempty"`

	// Pkcs12Password is the password that should be used in the PKCS#12 envelope. If
	// empty, p12 certificate files will not be protected by any password.
	Pkcs12Password string `json:"pkcs12Password,omitempty"`

	// FilenamesPrefix is the prefix that should be used in the exported certificate
	// filenames. If empty, defaults to "tls", so files will be created in the
	// temporary workspace directory as tls.key, tls.crt, tls.p12, etc...
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
