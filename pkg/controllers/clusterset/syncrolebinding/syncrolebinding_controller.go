package syncrolebinding

import (
	"context"
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

//This controller apply clusterset related clusterrolebinding based on clustersetToObjects and clustersetAdminToSubject map
type Reconciler struct {
	client                   client.Client
	scheme                   *runtime.Scheme
	clustersetAdminToSubject *helpers.ClustersetSubjectsMapper
	clustersetViewToSubject  *helpers.ClustersetSubjectsMapper
	clustersetToObjects      *helpers.ClusterSetMapper
	clustersetToClusters     *helpers.ClusterSetMapper
}

func SetupWithManager(mgr manager.Manager, clustersetAdminToSubject *helpers.ClustersetSubjectsMapper, clustersetViewToSubject *helpers.ClustersetSubjectsMapper, clustersetToObjects *helpers.ClusterSetMapper, clustersetToClusters *helpers.ClusterSetMapper) error {
	if err := add(mgr, newReconciler(mgr, clustersetAdminToSubject, clustersetViewToSubject, clustersetToObjects, clustersetToClusters)); err != nil {
		klog.Errorf("Failed to create clusterset rolebinding controller, %v", err)
		return err
	}
	return nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, clustersetAdminToSubject *helpers.ClustersetSubjectsMapper, clustersetViewToSubject *helpers.ClustersetSubjectsMapper, clustersetToObjects *helpers.ClusterSetMapper, clustersetToClusters *helpers.ClusterSetMapper) reconcile.Reconciler {
	return &Reconciler{
		client:                   mgr.GetClient(),
		scheme:                   mgr.GetScheme(),
		clustersetAdminToSubject: clustersetAdminToSubject,
		clustersetViewToSubject:  clustersetViewToSubject,
		clustersetToObjects:      clustersetToObjects,
		clustersetToClusters:     clustersetToClusters,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterset-rolebinding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

//This function sycn the rolebinding in namespace which in r.clustersetToObjects and r.clustersetToClusters
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	//reconcile every 10s
	reconcile := reconcile.Result{RequeueAfter: time.Duration(10) * time.Second}

	ctx := context.Background()
	unionClustersetToObjects := mergeClusterSetMap(r.clustersetToObjects, r.clustersetToClusters)

	adminErrs := r.syncRoleBinding(ctx, unionClustersetToObjects, r.clustersetAdminToSubject, "admin")
	viewErrs := r.syncRoleBinding(ctx, unionClustersetToObjects, r.clustersetViewToSubject, "view")

	errs := utils.AppendErrors(adminErrs, viewErrs)
	return reconcile, utils.NewMultiLineAggregate(errs)
}

// mergeClusterSetMap merge the objects in clustersetToObjects1 and clustersetToObjects2 when clusterset is same.
func mergeClusterSetMap(clustersetToObjects1, clustersetToObjects2 *helpers.ClusterSetMapper) *helpers.ClusterSetMapper {
	setToObjMap1 := clustersetToObjects1.GetAllClusterSetToObjects()
	if setToObjMap1 == nil {
		return clustersetToObjects2
	}
	setToObjMap2 := clustersetToObjects2.GetAllClusterSetToObjects()
	if setToObjMap2 == nil {
		return clustersetToObjects1
	}

	unionSetToObjMap := helpers.NewClusterSetMapper()
	for set, objs := range setToObjMap1 {
		if _, ok := setToObjMap2[set]; ok {
			unionObjs := objs.Union(setToObjMap2[set])
			unionSetToObjMap.UpdateClusterSetByObjects(set, unionObjs)
			continue
		}
		unionSetToObjMap.UpdateClusterSetByObjects(set, objs)
	}
	return unionSetToObjMap
}

func (r *Reconciler) syncRoleBinding(ctx context.Context, clustersetToObjects *helpers.ClusterSetMapper, clustersetToSubject *helpers.ClustersetSubjectsMapper, role string) []error {
	namespaceToSubject := generateNamespaceSubjectMap(clustersetToObjects, clustersetToSubject)
	//apply all disired clusterrolebinding
	errs := []error{}
	for namespace, subjects := range namespaceToSubject {
		requiredRoleBinding := generateRequiredRoleBinding(namespace, subjects, role)
		err := utils.ApplyRoleBinding(ctx, r.client, requiredRoleBinding)
		if err != nil {
			klog.Errorf("Failed to apply rolebinding: %v, error:%v", requiredRoleBinding.Name, err)
			errs = append(errs, err)
		}
	}

	//Delete rolebinding
	roleBindingList := &rbacv1.RoleBindingList{}

	//List Clusterset related clusterrolebinding
	matchExpressions := metav1.LabelSelectorRequirement{Key: clusterrolebinding.ClusterSetLabel, Operator: metav1.LabelSelectorOpExists}
	labelSelector := metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{matchExpressions}}
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return []error{err}
	}

	err = r.client.List(ctx, roleBindingList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return []error{err}
	}
	for _, roleBinding := range roleBindingList.Items {
		curRoleBinding := roleBinding

		//only handle current resource rolebinding
		matchRoleBindingName := utils.GenerateClustersetResourceRoleBindingName(role)

		if matchRoleBindingName != curRoleBinding.Name {
			continue
		}

		if _, ok := namespaceToSubject[roleBinding.Namespace]; !ok {
			err = r.client.Delete(ctx, &curRoleBinding)
			if err != nil {
				errs = append(errs, err)
			}
			continue
		}
	}
	return errs
}

func generateNamespaceSubjectMap(clustersetToObjects *helpers.ClusterSetMapper, clustersetToSubject *helpers.ClustersetSubjectsMapper) map[string][]rbacv1.Subject {
	var objectToSubject = make(map[string][]rbacv1.Subject)

	for clusterset, subjects := range clustersetToSubject.GetMap() {
		if clusterset == "*" {
			continue
		}
		objects := clustersetToObjects.GetObjectsOfClusterSet(clusterset)
		for _, object := range objects.List() {
			objectToSubject[object] = utils.Mergesubjects(objectToSubject[object], subjects)
		}
	}
	if len(clustersetToSubject.Get("*")) == 0 {
		return objectToSubject
	}
	//if clusterset is "*", should map this subjects to all namespace
	allClustersetToObjects := clustersetToObjects.GetAllClusterSetToObjects()
	for _, objs := range allClustersetToObjects {
		subjects := clustersetToSubject.Get("*")
		for _, obj := range objs.List() {
			objectToSubject[obj] = utils.Mergesubjects(objectToSubject[obj], subjects)
		}
	}
	return objectToSubject
}

func generateRequiredRoleBinding(resourceNameSpace string, subjects []rbacv1.Subject, role string) *rbacv1.RoleBinding {
	roleBindingName := utils.GenerateClustersetResourceRoleBindingName(role)

	var labels = make(map[string]string)
	labels[clusterrolebinding.ClusterSetLabel] = "true"
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: resourceNameSpace,
			Labels:    labels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     role,
		},
		Subjects: subjects,
	}
}
