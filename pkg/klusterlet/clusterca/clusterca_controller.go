package clusterca

import (
	"context"
	"reflect"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	clusterv1 "github.com/open-cluster-management/api/cluster/v1"
	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/utils"
	openshiftclientset "github.com/openshift/client-go/config/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	UpdateInterval           = 60 * time.Second
	infrastructureConfigName = "cluster"
)

type ClusterCaReconciler struct {
	Client          client.Client
	OpenshiftClient openshiftclientset.Interface
	KubeClient      kubernetes.Interface
}

func (r *ClusterCaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.ManagedCluster{}).
		Complete(r)
}

func (r *ClusterCaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cluster := &clusterv1.ManagedCluster{}
	klog.V(5).Infof("reconcile: %+v", req)
	err := r.Client.Get(ctx, req.NamespacedName, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	//get ocp apiserver url
	infraConfig, err := r.OpenshiftClient.ConfigV1().Infrastructures().Get(ctx, infrastructureConfigName, v1.GetOptions{})
	if err != nil {
		return ctrl.Result{}, err
	}
	kubeAPIServer := infraConfig.Status.APIServerURL

	//get ocp ca
	clusterca, err := utils.GetClusterCa(r.OpenshiftClient, r.KubeClient, kubeAPIServer)
	if err != nil {
		klog.Errorf("Failed to get clusterca. error:%v", err)
		return ctrl.Result{}, err
	}
	klog.V(5).Infof("kubeapiserver: %v, clusterca:%v", kubeAPIServer, clusterca)
	//check if managedcluster ca need update
	var needUpdate = false
	for i, clientConfig := range cluster.Spec.ManagedClusterClientConfigs {
		if clientConfig.URL != kubeAPIServer {
			continue
		}
		if reflect.DeepEqual(clusterca, clientConfig.CABundle) {
			return ctrl.Result{RequeueAfter: UpdateInterval}, nil
		}
		needUpdate = true
		cluster.Spec.ManagedClusterClientConfigs[i].CABundle = clusterca
	}

	if !needUpdate {
		newConfig := clusterv1.ClientConfig{
			URL:      kubeAPIServer,
			CABundle: clusterca,
		}
		cluster.Spec.ManagedClusterClientConfigs = append(cluster.Spec.ManagedClusterClientConfigs, newConfig)
	}

	err = r.Client.Update(ctx, cluster, &client.UpdateOptions{})
	if err != nil {
		klog.Errorf("Failed to update clusterca. error:%v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: UpdateInterval}, nil
}
