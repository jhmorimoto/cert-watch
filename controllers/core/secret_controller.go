package core

import (
	"context"
	"strings"
	"time"

	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	"github.com/jhmorimoto/cert-watch/controllers/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
)

var retryFastDelay = time.Second * time.Duration(5)
var retrySlowDelay = time.Second * time.Duration(30)
var retryMaxFastAttempts = 5
var log = ctrl.Log.WithName("SecretController")

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get

func (r *SecretReconciler) updateCertWatcher(ctx context.Context, certwatcher *certwatchv1.CertWatcher) (ctrl.Result, error) {
	certwatcher.Status.LastUpdate = apimachineryv1.Now()
	if err := r.Status().Update(ctx, certwatcher); err != nil {
		log.Error(err, certwatcher.Namespace+"/"+certwatcher.Name+" Unable to update CertWatcher")
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var secretlogname string = req.Namespace + "/" + req.Name
	var s corev1.Secret
	err := r.Get(ctx, req.NamespacedName, &s)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Info(secretlogname + " Unable to get Secret: " + err.Error())
		} else {
			log.Error(err, secretlogname+" Unable to get Secret")
		}
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}
	if s.Type != "kubernetes.io/tls" {
		return ctrl.Result{}, client.IgnoreNotFound(nil)
	}
	dataChecksum, err := util.SecretDataChecksum(&s)
	if err != nil {
		log.Error(err, secretlogname+" Unable to serialize secret data")
		// Wait a minute before trying again
		return ctrl.Result{Requeue: true}, err
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
			if cw.Status.Status != "Ready" {
				r.EventRecorder.Eventf(&cw, "Warning", "SecretChanged", "Secret changed, but CertWatcher not Ready.")
				// return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
			}
			if cw.Status.ActionStatus == "Pending" {
				r.EventRecorder.Eventf(&cw, "Warning", "SecretChanged", "Secret changed, but CertWatcher has Pending actions.")
				// return ctrl.Result{Requeue: true, RequeueAfter: retryPeriod}, err
			}
			if cw.Status.LastChecksum != dataChecksum {
				cw.Status.LastChecksum = dataChecksum
				cw.Status.Message = "Checksum updated"
				cw.Status.ActionStatus = "Pending"
				r.EventRecorder.Eventf(&cw, "Normal", "SecretChanged", "Updating CertWatcher status.")
				// return r.updateCertWatcher(ctx, &cw)
				r.updateCertWatcher(ctx, &cw)
			}
		}
	} else {
		log.Info(secretlogname + " Secret does not seem to have any CertWatchers")
	}
	return ctrl.Result{}, nil
}

var rateLimiter ratelimiter.RateLimiter = workqueue.NewItemFastSlowRateLimiter(retryFastDelay, retrySlowDelay, retryMaxFastAttempts)

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithOptions(controller.Options{RateLimiter: rateLimiter}).
		Complete(r)
}
