package core

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get

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
	dataChecksum, err := secretDataChecksum(&s)
	if err != nil {
		klog.Errorf("%s/%s Unable to serialize secret data: %s", s.Namespace, s.Name, err.Error())
		// Wait a minute before trying again
		return ctrl.Result{Requeue: true, RequeueAfter: time.Minute}, err
	}

	// Find CertWatchers that watch this particular Secret and update their statuses
	var cwList v1.CertWatcherList
	err = r.List(ctx, &cwList, client.MatchingFields{".spec.secret.name": s.Name}, client.MatchingFields{".spec.secret.namespace": s.Namespace})
	if err != nil {
		klog.Errorf("Unable to get CertWatcher list: %s", err.Error())
	}
	cwListLen := len(cwList.Items)
	if cwListLen > 0 {
		for _, cw := range cwList.Items {
			klog.Infof("%s/%s CertWatcher %s/%s", s.Namespace, s.Name, cw.Namespace, cw.Name)
			cw.Status.LastUpdate = metav1.Now()
			cw.Status.Checksum = dataChecksum
			if err := r.Status().Update(ctx, &cw); err != nil {
				klog.Errorf("%s/%s Unable to update CertWatcher %s/%s: %s", s.Namespace, s.Name, cw.Namespace, cw.Name, err.Error())
			}
			klog.Infof("%s/%s CertWatcher updated %s/%s", s.Namespace, s.Name, cw.Namespace, cw.Name)
		}
	} else {
		klog.Infof("%s/%s does not have any CertWatchers", s.Namespace, s.Name)
	}
	return ctrl.Result{}, nil
}

// Calculate SHA256 from the Secret data
func secretDataChecksum(s *corev1.Secret) (string, error) {
	dataJson, err := json.Marshal(s.Data)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	hash.Write(dataJson)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil)), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
