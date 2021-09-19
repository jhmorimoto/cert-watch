package certwatch

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	v1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
)

// CertWatcherReconciler reconciles a CertWatcher object
type CertWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	cw := v1.CertWatcher{}
	err := r.Get(ctx, req.NamespacedName, &cw)
	if err != nil {
		klog.Warningf("Unable to get CertWatcher %s/%s: %s", req.Namespace, req.Name, err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}
	klog.Infof("CertWatcher is here %s/%s echo.enabled=%t", cw.Namespace, cw.Name, cw.Spec.Actions.Echo.Enabled)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Creating field indeces allows the CRD resource to be searched by MatchingFields.

	// Create index for .spec.secret.name
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1.CertWatcher{}, ".spec.secret.name", func(rawObj client.Object) []string {
		cw := rawObj.(*v1.CertWatcher)
		return []string{cw.Spec.Secret.Name}
	})
	if err != nil {
		klog.Errorf("Unable to create index for .spec.secret.name: %s", err.Error())
		return err
	}

	// Create index for .spec.secret.namespace
	err = mgr.GetFieldIndexer().IndexField(context.Background(), &v1.CertWatcher{}, ".spec.secret.namespace", func(rawObj client.Object) []string {
		cw := rawObj.(*v1.CertWatcher)
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
