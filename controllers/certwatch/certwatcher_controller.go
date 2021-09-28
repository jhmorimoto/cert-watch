package certwatch

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	apicorev1 "k8s.io/api/core/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"

	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	"github.com/jhmorimoto/cert-watch/controllers/util"
	"github.com/magiconair/properties"
)

var retryFastDelay = time.Second * time.Duration(5)
var retrySlowDelay = time.Second * time.Duration(30)
var retryMaxFastAttempts = 5
var log = ctrl.Log.WithName("CertWatcherController")

// CertWatcherReconciler reconciles a CertWatcher object
type CertWatcherReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	EmailConfiguration *properties.Properties
	EventRecorder      record.EventRecorder
}

func (r *CertWatcherReconciler) updateCertWatcher(ctx context.Context, certwatcher *certwatchv1.CertWatcher, originalError error) (ctrl.Result, error) {
	certwatcher.Status.LastUpdate = apimachineryv1.Now()
	if err := r.Status().Update(ctx, certwatcher); err != nil {
		r.EventRecorder.Eventf(certwatcher, "Warning", "CertWatcherFailure", "Unable update CertWatcher: %s", err.Error())
		// log.Error(err, certwatcher.Namespace+"/"+certwatcher.Namespace+" Unable to update CertWatcher")
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{Requeue: originalError != nil}, originalError
}

//+kubebuilder:rbac:groups=certwatch.morimoto.net.br,resources=certwatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=certwatch.morimoto.net.br,resources=certwatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=certwatch.morimoto.net.br,resources=certwatchers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CertWatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CertWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var certwatcherlogname = req.Namespace + "/" + req.Name
	var certwatcher certwatchv1.CertWatcher
	err := r.Get(ctx, req.NamespacedName, &certwatcher)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Info(certwatcherlogname + " Unable to get CertWatcher: " + err.Error())
		} else {
			log.Error(err, certwatcherlogname+" Unable to get CertWatcher")
		}
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}

	var secretlogname = certwatcher.Spec.Secret.Namespace + "/" + certwatcher.Spec.Secret.Name

	// If Status is not Ready, then initiate this CertWatcher, update the Status
	// and exit. Before initiation, no Secret changes will be processed.
	if certwatcher.Status.Status != "Ready" {
		certwatcher.Status.Status = "NotReady"
		var secret apicorev1.Secret
		var checksum string
		err = r.Get(ctx, types.NamespacedName{Namespace: certwatcher.Spec.Secret.Namespace, Name: certwatcher.Spec.Secret.Name}, &secret)
		if err != nil {
			r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherInit", "Unable to find Secret %s: %s", secretlogname, err.Error())
			certwatcher.Status.Message = "Unable to find Secret " + secretlogname + ": " + err.Error()
			return r.updateCertWatcher(ctx, &certwatcher, err)
		}
		checksum, err = util.SecretDataChecksum(&secret)
		if err != nil {
			r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherInit", "calculate secret checksum %s: %s", secretlogname, err.Error())
			certwatcher.Status.Message = "Unable to calculate Secret checksum " + secretlogname + ": " + err.Error()
			return r.updateCertWatcher(ctx, &certwatcher, err)
		}
		certwatcher.Status.LastChecksum = checksum
		certwatcher.Status.Status = "Ready"
		certwatcher.Status.Message = "CertWatcher successfully initialized"
		certwatcher.Status.ActionStatus = ""
		r.EventRecorder.Eventf(&certwatcher, "Normal", "CertWatcherInit", "CertWatcher successfully initialized")
		return r.updateCertWatcher(ctx, &certwatcher, nil)
	}

	// If ActionStatus is Pending, then process all actions and change the
	// Status back to Ready.
	if certwatcher.Status.ActionStatus == "Pending" {
		r.EventRecorder.Eventf(&certwatcher, "Normal", "CertWatcherProcessing", "Processing pending actions")
		var secret apicorev1.Secret
		var certFilesDir string
		err = r.Get(ctx, types.NamespacedName{Namespace: certwatcher.Spec.Secret.Namespace, Name: certwatcher.Spec.Secret.Name}, &secret)
		if err != nil {
			r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherProcessing", "Unable to find Secret for processing %s", secretlogname)
			certwatcher.Status.Message = "Unable to find Secret for processing" + secretlogname + ": " + err.Error()
			return r.updateCertWatcher(ctx, &certwatcher, err)
		}

		certFilesDir, err = util.CreateCertificateFiles(&secret, certwatcher.Spec.FilenamesPrefix, certwatcher.Spec.ZipFilesPassword, certwatcher.Spec.Pkcs12Password)
		defer os.RemoveAll(certFilesDir)
		if err != nil {
			r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherProcessing", "%s", err.Error())
			certwatcher.Status.Message = err.Error()
			return r.updateCertWatcher(ctx, &certwatcher, err)
		}

		if certwatcher.Spec.Actions.Echo.Enabled {
			r.EventRecorder.Eventf(&certwatcher, "Normal", "CertWatcherProcessing", "ECHO: Good morning to %s", secretlogname)
		}

		if certwatcher.Spec.Actions.Email.Enabled {
			var emailConfig *properties.Properties = r.EmailConfiguration
			if certwatcher.Spec.Actions.Email.ConfigFile != "" {
				emailConfig = properties.MustLoadFile(certwatcher.Spec.Actions.Email.ConfigFile, properties.UTF8)
			}
			r.EventRecorder.Eventf(&certwatcher, "Normal", "CertWatcherProcessing", "EMAIL: Sending mail to %s via %s:%d", certwatcher.Spec.Actions.Email.To, emailConfig.GetString("host", ""), emailConfig.GetInt("port", 0))
			err = util.ProcessEmail(&certwatcher, certFilesDir, emailConfig)
			if err != nil {
				r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherProcessing", "EMAIL: %s", err.Error())
				certwatcher.Status.Message = err.Error()
				return r.updateCertWatcher(ctx, &certwatcher, err)
			}
		}
		if certwatcher.Spec.Actions.Scp.Enabled {
			if certwatcher.Spec.Actions.Scp.Port == 0 {
				certwatcher.Spec.Actions.Scp.Port = 22
			}
			var credentialSecret apicorev1.Secret
			var credentialSecretName []string = strings.Split(certwatcher.Spec.Actions.Scp.CredentialSecret, "/")
			if len(credentialSecretName) < 2 {
				r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherProcessing",
					"SCP: Invalid credentialSecret naming format %s", certwatcher.Spec.Actions.Scp.CredentialSecret)
				certwatcher.Status.Message = fmt.Sprintf("SCP: Invalid credentialSecret naming format %s",
					certwatcher.Spec.Actions.Scp.CredentialSecret)
				return r.updateCertWatcher(ctx, &certwatcher, err)
			}
			err = r.Get(ctx, types.NamespacedName{Namespace: credentialSecretName[0], Name: credentialSecretName[1]}, &credentialSecret)
			if err != nil {
				r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherProcessing", "SCP: %s", err.Error())
				certwatcher.Status.Message = fmt.Sprintf("SCP: %s", err.Error())
				return r.updateCertWatcher(ctx, &certwatcher, err)
			}
			r.EventRecorder.Eventf(&certwatcher, "Normal", "CertWatcherProcessing", "SCP: Sending files to %s:%d", certwatcher.Spec.Actions.Scp.Hostname, certwatcher.Spec.Actions.Scp.Port)
			err = util.ProcessScp(&certwatcher, credentialSecret, certFilesDir)
			if err != nil {
				r.EventRecorder.Eventf(&certwatcher, "Warning", "CertWatcherProcessing", "SCP: %s", err.Error())
				certwatcher.Status.Message = err.Error()
				return r.updateCertWatcher(ctx, &certwatcher, err)
			}
		}

		certwatcher.Status.ActionStatus = "Ready"
		certwatcher.Status.Message = "Waiting for next Secret change"
		r.EventRecorder.Eventf(&certwatcher, "Normal", "CertWatcherProcessing", "Action processig finished successfully")
		return r.updateCertWatcher(ctx, &certwatcher, nil)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Creating field indeces allows the CRD resource to be searched by MatchingFields.

	// Create index for .spec.secret.name
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &certwatchv1.CertWatcher{}, ".spec.secret.name", func(rawObj client.Object) []string {
		cw := rawObj.(*certwatchv1.CertWatcher)
		return []string{cw.Spec.Secret.Name}
	})
	if err != nil {
		log.Error(err, "Unable to create index for .spec.secret.name")
		return err
	}

	// Create index for .spec.secret.namespace
	err = mgr.GetFieldIndexer().IndexField(context.Background(), &certwatchv1.CertWatcher{}, ".spec.secret.namespace", func(rawObj client.Object) []string {
		cw := rawObj.(*certwatchv1.CertWatcher)
		return []string{cw.Spec.Secret.Namespace}
	})
	if err != nil {
		log.Error(err, "Unable to create index for .spec.secret.namespace")
		return err
	}

	var rateLimiter ratelimiter.RateLimiter = workqueue.NewItemFastSlowRateLimiter(retryFastDelay, retrySlowDelay, retryMaxFastAttempts)

	return ctrl.NewControllerManagedBy(mgr).
		For(&certwatchv1.CertWatcher{}).
		WithOptions(controller.Options{RateLimiter: rateLimiter}).
		Complete(r)
}
