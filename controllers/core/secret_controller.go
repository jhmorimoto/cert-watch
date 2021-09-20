package core

import (
	"context"
	"time"

	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	"github.com/jhmorimoto/cert-watch/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var retryPeriod = time.Second * 10
var log = ctrl.Log.WithName("SecretController")

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get

func (r *SecretReconciler) updateCertWatcher(ctx context.Context, certwatcher *certwatchv1.CertWatcher) (ctrl.Result, error) {
	if err := r.Status().Update(ctx, certwatcher); err != nil {
		log.Error(err, certwatcher.Namespace+"/"+certwatcher.Name+" Unable to update CertWatcher")
		return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
	}
	return ctrl.Result{}, nil
}

func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var secretlogname string = req.Namespace+"/"+req.Name
	var s corev1.Secret
	err := r.Get(ctx, req.NamespacedName, &s)
	if err != nil {
		log.Error(err, secretlogname+" Unable to get Secret")
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}
	if s.Type != "kubernetes.io/tls" {
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}
	dataChecksum, err := util.SecretDataChecksum(&s)
	if err != nil {
		log.Error(err, secretlogname+" Unable to serialize secret data")
		// Wait a minute before trying again
		return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
	}

	// Find CertWatchers that watch this particular Secret and update their statuses
	var cwList certwatchv1.CertWatcherList
	err = r.List(ctx, &cwList, client.MatchingFields{".spec.secret.name": s.Name}, client.InNamespace(s.Namespace))
	if err != nil {
		log.Error(err, secretlogname+" Unable to get CertWatcher list")
	}
	cwListLen := len(cwList.Items)
	if cwListLen > 0 {
		for _, cw := range cwList.Items {
			var cwlogname string = cw.Namespace+"/"+cw.Name
			if cw.Status.Status != "Ready" {
				log.Info(secretlogname+" Secret updated, but CertWatcher "+cwlogname+" is not Ready. Will retry in "+string(retryPeriod/time.Second)+" seconds")
				return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
			}
			if cw.Status.ActionStatus == "Pending" {
				log.Error(err, secretlogname+" Secret updated, but CertWatcher "+cwlogname+" actions still pending. Will retry in "+string(retryPeriod/time.Second)+" seconds")
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
		log.Info(secretlogname+" Secret does not seem to have any CertWatchers")
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
