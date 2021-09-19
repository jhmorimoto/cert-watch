package certwatch

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	"github.com/jhmorimoto/cert-watch/util"
)

var retryPeriod = time.Second * 10

// CertWatcherReconciler reconciles a CertWatcher object
type CertWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CertWatcherReconciler) updateCertWatcher(ctx context.Context, certwatcher *certwatchv1.CertWatcher) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, certwatcher); err != nil {
		klog.Errorf("%s/%s Unable to update CertWatcher: %s", certwatcher.Namespace, certwatcher.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
	}
	return ctrl.Result{}, nil
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
	var certwatcher certwatchv1.CertWatcher
	err := r.Get(ctx, req.NamespacedName, &certwatcher)
	if err != nil {
		klog.Warningf("Unable to get CertWatcher %s/%s: %s", req.Namespace, req.Name, err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}

	if certwatcher.Status.Status != "Ready" {
		certwatcher.Status.Status = "NotReady"
		var secret corev1.Secret
		var checksum string
		err = r.Get(ctx, types.NamespacedName{Namespace: certwatcher.Spec.Secret.Namespace, Name: certwatcher.Spec.Secret.Name}, &secret)
		if err != nil {
			klog.Errorf("%s/%s Unable to find Secret %s/%s: %s", certwatcher.Namespace, certwatcher.Name, certwatcher.Spec.Secret.Namespace, certwatcher.Spec.Secret.Name, err.Error())
			certwatcher.Status.Message = fmt.Sprintf("Unable to find Secret %s/%s: %s", certwatcher.Spec.Secret.Namespace, certwatcher.Spec.Secret.Name, err.Error())
			return r.updateCertWatcher(ctx, &certwatcher)
		}
		checksum, err = util.SecretDataChecksum(&secret)
		if err != nil {
			certwatcher.Status.Message = fmt.Sprintf("Unable to calculate Secret checksum %s/%s: %s", certwatcher.Spec.Secret.Namespace, certwatcher.Spec.Secret.Name, err.Error())
			return r.updateCertWatcher(ctx, &certwatcher)
		}
		certwatcher.Status.LastChecksum = checksum
		certwatcher.Status.Status = "Ready"
		certwatcher.Status.Message = "CertWatcher successfully initialized"
		certwatcher.Status.ActionStatus = ""
		certwatcher.Status.LastUpdate = apimachineryv1.Now()
		return r.updateCertWatcher(ctx, &certwatcher)
	}

	if certwatcher.Status.ActionStatus == "Pending" {
		klog.Infof("%s/%s running actions", certwatcher.Namespace, certwatcher.Name)
		time.Sleep(10 * time.Second)
		if certwatcher.Spec.Actions.Echo.Enabled {
			klog.Infof("%s/%s is letting your know this Secret has just changed %s/%s", certwatcher.Namespace, certwatcher.Name, certwatcher.Spec.Secret.Namespace, certwatcher.Spec.Secret.Name)
		}
		certwatcher.Status.ActionStatus = "Ready"
		certwatcher.Status.Message = "Action processig finished successfully"
		return r.updateCertWatcher(ctx, &certwatcher)
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
		klog.Errorf("Unable to create index for .spec.secret.name: %s", err.Error())
		return err
	}

	// Create index for .spec.secret.namespace
	err = mgr.GetFieldIndexer().IndexField(context.Background(), &certwatchv1.CertWatcher{}, ".spec.secret.namespace", func(rawObj client.Object) []string {
		cw := rawObj.(*certwatchv1.CertWatcher)
		return []string{cw.Spec.Secret.Namespace}
	})
	if err != nil {
		klog.Errorf("Unable to create index for .spec.secret.namespace: %s", err.Error())
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&certwatchv1.CertWatcher{}).
		Complete(r)
}
