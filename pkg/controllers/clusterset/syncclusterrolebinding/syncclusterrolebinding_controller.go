package syncclusterrolebinding

import (
	"context"
	"strings"
	"time"

	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/helpers"
	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/controllers/clusterset/clusterrolebinding"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//This controller apply clusterset related clusterrolebinding based on clustersetToClusters and clustersetAdminToSubject map
type Reconciler struct {
	client                   client.Client
	scheme                   *runtime.Scheme
	clustersetAdminToSubject *helpers.ClustersetSubjectsMapper
	clustersetViewToSubject  *helpers.ClustersetSubjectsMapper
	clustersetToClusters     *helpers.ClusterSetMapper
}

func SetupWithManager(mgr manager.Manager, clustersetAdminToSubject *helpers.ClustersetSubjectsMapper, clustersetViewToSubject *helpers.ClustersetSubjectsMapper, clustersetToClusters *helpers.ClusterSetMapper) error {
	if err := add(mgr, newReconciler(mgr, clustersetAdminToSubject, clustersetViewToSubject, clustersetToClusters)); err != nil {
		klog.Errorf("Failed to create ClusterRoleBinding controller, %v", err)
		return err
	}
	return nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, clustersetAdminToSubject *helpers.ClustersetSubjectsMapper, clustersetViewToSubject *helpers.ClustersetSubjectsMapper, clustersetToClusters *helpers.ClusterSetMapper) reconcile.Reconciler {
	return &Reconciler{
		client:                   mgr.GetClient(),
		scheme:                   mgr.GetScheme(),
		clustersetAdminToSubject: clustersetAdminToSubject,
		clustersetViewToSubject:  clustersetViewToSubject,
		clustersetToClusters:     clustersetToClusters,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterinfo-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRoleBinding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	//reconcile every 10s
	reconcile := reconcile.Result{RequeueAfter: time.Duration(10) * time.Second}

	ctx := context.Background()

	adminErrs := r.syncManagedClusterClusterroleBinding(ctx, r.clustersetAdminToSubject, "admin")
	viewErrs := r.syncManagedClusterClusterroleBinding(ctx, r.clustersetViewToSubject, "view")

	errs := utils.AppendErrors(adminErrs, viewErrs)
	return reconcile, utils.NewMultiLineAggregate(errs)
}

func (r *Reconciler) syncManagedClusterClusterroleBinding(ctx context.Context, clustersetToSubject *helpers.ClustersetSubjectsMapper, role string) []error {
	clusterToSubject := generateClusterSubjectMap(r.clustersetToClusters, clustersetToSubject)

	errs := []error{}
	//apply all disired clusterrolebinding
	for cluster, subjects := range clusterToSubject {
		requiredClusterRoleBinding := generateRequiredClusterRoleBinding(cluster, subjects, role)
		err := utils.ApplyClusterRoleBinding(ctx, r.client, requiredClusterRoleBinding)
		if err != nil {
			klog.Errorf("Failed to apply clusterrolebinding: %v, error:%v", requiredClusterRoleBinding.Name, err)
			errs = append(errs, err)
		}
	}
	//Delete clusterrolebinding
	clusterRoleBindingList := &rbacv1.ClusterRoleBindingList{}

	//List Clusterset related clusterrolebinding
	matchExpressions := metav1.LabelSelectorRequirement{Key: clusterrolebinding.ClusterSetLabel, Operator: metav1.LabelSelectorOpExists}
	labelSelector := metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{matchExpressions}}
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return []error{err}
	}

	err = r.client.List(ctx, clusterRoleBindingList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return []error{err}
	}
	for _, clusterRoleBinding := range clusterRoleBindingList.Items {
		curClusterRoleBinding := clusterRoleBinding
		curClusterName := getClusterNameInClusterrolebinding(curClusterRoleBinding.Name, role)
		if curClusterName == "" {
			continue
		}
		if _, ok := clusterToSubject[curClusterName]; !ok {
			err = r.client.Delete(ctx, &curClusterRoleBinding)
			if err != nil {
				errs = append(errs, err)
			}
			continue
		}
	}
	return errs
}

// The clusterset related managedcluster clusterrolebinding format should be: open-cluster-management:managedclusterset:"admin":managedcluster:cluster1
func getClusterNameInClusterrolebinding(clusterrolebinding, role string) string {
	splitName := strings.Split(clusterrolebinding, ":")
	requiredName := utils.GenerateClustersetClusterRoleBindingName("PLACEHOLDER", role)
	requiredSplitName := strings.Split(requiredName, ":")
	l := len(requiredSplitName)
	for k, v := range requiredSplitName {
		if k == l-1 {
			return splitName[k]
		}
		if v != splitName[k] {
			return ""
		}
	}
	return ""
}

func generateClusterSubjectMap(clustersetToClusters *helpers.ClusterSetMapper, clustersetToSubject *helpers.ClustersetSubjectsMapper) map[string][]rbacv1.Subject {
	var clusterToSubject = make(map[string][]rbacv1.Subject)

	for clusterset, subjects := range clustersetToSubject.GetMap() {
		if clusterset == "*" {
			continue
		}
		clusters := clustersetToClusters.GetObjectsOfClusterSet(clusterset)
		for _, cluster := range clusters.List() {
			clusterToSubject[cluster] = utils.Mergesubjects(clusterToSubject[cluster], subjects)
		}
	}

	if len(clustersetToSubject.Get("*")) == 0 {
		return clusterToSubject
	}
	//if clusterset is "*", should map this subjects to all clusters
	allClustersetToClusters := clustersetToClusters.GetAllClusterSetToObjects()
	for _, clusters := range allClustersetToClusters {
		subjects := clustersetToSubject.Get("*")
		for _, cluster := range clusters.List() {
			clusterToSubject[cluster] = utils.Mergesubjects(clusterToSubject[cluster], subjects)
		}
	}
	return clusterToSubject
}

func generateRequiredClusterRoleBinding(clusterName string, subjects []rbacv1.Subject, role string) *rbacv1.ClusterRoleBinding {
	clusterRoleBindingName := utils.GenerateClustersetClusterRoleBindingName(clusterName, role)
	clusterRoleName := utils.GenerateClusterRoleName(clusterName, role)

	var labels = make(map[string]string)
	labels[clusterrolebinding.ClusterSetLabel] = "true"
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   clusterRoleBindingName,
			Labels: labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
		Subjects: subjects,
	}
}
