package core

import (
	"context"
	"time"

	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	"github.com/jhmorimoto/cert-watch/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var retryPeriod = time.Second*10

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get

func (r *SecretReconciler) updateCertWatcher(ctx context.Context, certwatcher *certwatchv1.CertWatcher) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, certwatcher); err != nil {
		klog.Errorf("%s/%s Unable to update CertWatcher: %s", certwatcher.Namespace, certwatcher.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
	}
	return ctrl.Result{}, nil
}

func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var s corev1.Secret
	err := r.Get(ctx, req.NamespacedName, &s)
	if err != nil {
		klog.Warningf("Unable to get Secret %s/%s: %s", req.Namespace, req.Name, err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}
	if s.Type != "kubernetes.io/tls" {
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}
	dataChecksum, err := util.SecretDataChecksum(&s)
	if err != nil {
		klog.Errorf("%s/%s Unable to serialize secret data: %s", s.Namespace, s.Name, err.Error())
		// Wait a minute before trying again
		return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
	}

	// Find CertWatchers that watch this particular Secret and update their statuses
	var cwList certwatchv1.CertWatcherList
	err = r.List(ctx, &cwList, client.MatchingFields{".spec.secret.name": s.Name}, client.InNamespace(s.Namespace))
	if err != nil {
		klog.Errorf("Unable to get CertWatcher list: %s", err.Error())
	}
	cwListLen := len(cwList.Items)
	if cwListLen > 0 {
		for _, cw := range cwList.Items {
			if cw.Status.Status != "Ready" {
				klog.Warningf("%s/%s Secret updated, but CertWatcher %s/%s is not Ready. Will retry in %d seconds", s.Namespace, s.Name, cw.Namespace, cw.Name, retryPeriodSeconds/time.Second)
				return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
			}
			if cw.Status.ActionStatus == "Pending" {
				klog.Errorf("%s/%s Secret updated, but CertWatcher %s/%s actions still pending. Will retry in %d seconds", s.Namespace, s.Name, cw.Namespace, cw.Name, retryPeriodSeconds/time.Second)
				return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
			}
			if cw.Status.LastChecksum != dataChecksum {
				cw.Status.LastUpdate = apimachineryv1.Now()
				cw.Status.LastChecksum = dataChecksum
				cw.Status.Message = "Checksum updated"
				cw.Status.ActionStatus = "Pending"
				return r.updateCertWatcher(ctx, &cw)
			}
		}
	} else {
		klog.Infof("%s/%s does not seem to have any CertWatchers", s.Namespace, s.Name)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
