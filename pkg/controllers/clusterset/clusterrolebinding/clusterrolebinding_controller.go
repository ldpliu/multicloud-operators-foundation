package clusterrolebinding

import (
	"context"

	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/helpers"
	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ClustersetFinalizerName string = "open-cluster-management.io/clusterset"
	ClusterSetLabel         string = "clusterset.cluster.open-cluster-management.io"
)

//This Controller will generate a Clusterset to Subjects map, and this map will be used to sync
// clusterset related clusterrolebinding.
type Reconciler struct {
	client                       client.Client
	scheme                       *runtime.Scheme
	clusterroleToClustersetAdmin map[string]sets.String
	clustersetAdminToSubject     *helpers.ClustersetSubjectsMapper
	clusterroleToClustersetView  map[string]sets.String
	clustersetViewToSubject      *helpers.ClustersetSubjectsMapper
}

func SetupWithManager(mgr manager.Manager, clustersetAdminToSubject *helpers.ClustersetSubjectsMapper, clustersetViewToSubject *helpers.ClustersetSubjectsMapper) error {
	if err := add(mgr, newReconciler(mgr, clustersetAdminToSubject, clustersetViewToSubject)); err != nil {
		klog.Errorf("Failed to create clustersetAdminToSubject controller, %v", err)
		return err
	}
	return nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, clustersetAdminToSubject *helpers.ClustersetSubjectsMapper, clustersetViewToSubject *helpers.ClustersetSubjectsMapper) reconcile.Reconciler {
	var clusterroleToClustersetAdmin = make(map[string]sets.String)
	var clusterroleToClustersetView = make(map[string]sets.String)
	return &Reconciler{
		client:                       mgr.GetClient(),
		scheme:                       mgr.GetScheme(),
		clustersetAdminToSubject:     clustersetAdminToSubject,
		clustersetViewToSubject:      clustersetViewToSubject,
		clusterroleToClustersetAdmin: clusterroleToClustersetAdmin,
		clusterroleToClustersetView:  clusterroleToClustersetView,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterrolebinding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &rbacv1.ClusterRole{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(
		&source.Kind{Type: &rbacv1.ClusterRoleBinding{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
				clusterrolebinding, ok := a.Object.(*rbacv1.ClusterRoleBinding)
				if !ok {
					// not a clusterrolebinding, returning empty
					klog.Error("Clusterrolebinding handler received non-SyncSet object")
					return []reconcile.Request{}
				}
				var requests []reconcile.Request
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name: clusterrolebinding.RoleRef.Name,
					},
				})
				return requests
			}),
		})
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	clusterrole := &rbacv1.ClusterRole{}
	isAdmin := false
	isView := false
	err := r.client.Get(ctx, req.NamespacedName, clusterrole)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//generate clusterrole's all clustersets Admin
	clustersetAdminInRule := utils.GetClustersetAdminInRules(clusterrole.Rules)
	//generate clusterrole's all clustersets view
	clustersetViewInRule := utils.GetClustersetViewInRules(clusterrole.Rules)

	//Add Finalizer to clusterset related clusterrole
	if (clustersetAdminInRule.Len() != 0 || clustersetViewInRule.Len() != 0) && !utils.ContainsString(clusterrole.GetFinalizers(), ClustersetFinalizerName) {
		klog.Infof("adding ClusterRoleBinding Finalizer to ClusterRole %v", clusterrole.Name)
		clusterrole.ObjectMeta.Finalizers = append(clusterrole.ObjectMeta.Finalizers, ClustersetFinalizerName)
		if err := r.client.Update(context.TODO(), clusterrole); err != nil {
			klog.Warningf("will reconcile since failed to add finalizer to ClusterRole %v, %v", clusterrole.Name, err)
			return reconcile.Result{}, err
		}
	}

	//clusterrole is deleted or clusterrole is not related to any clusterset
	if (clustersetAdminInRule.Len() == 0 && clustersetViewInRule.Len() == 0) || !clusterrole.GetDeletionTimestamp().IsZero() {
		// The object is being deleted
		if utils.ContainsString(clusterrole.GetFinalizers(), ClustersetFinalizerName) {
			klog.Infof("removing ClusterRoleBinding Finalizer in ClusterRole %v", clusterrole.Name)
			clusterrole.ObjectMeta.Finalizers = utils.RemoveString(clusterrole.ObjectMeta.Finalizers, ClustersetFinalizerName)
			if err := r.client.Update(context.TODO(), clusterrole); err != nil {
				klog.Warningf("will reconcile since failed to remove Finalizer from ClusterRole %v, %v", clusterrole.Name, err)
				return reconcile.Result{}, err
			}
		}
		if _, ok := r.clusterroleToClustersetAdmin[clusterrole.Name]; ok {
			isAdmin = true
			delete(r.clusterroleToClustersetAdmin, clusterrole.Name)
		}

		if _, ok := r.clusterroleToClustersetView[clusterrole.Name]; ok {
			isView = true
			delete(r.clusterroleToClustersetView, clusterrole.Name)
		}
		if !isAdmin && !isView {
			return ctrl.Result{}, nil
		}

	}
	if clustersetAdminInRule.Len() != 0 {
		r.clusterroleToClustersetAdmin[clusterrole.Name] = clustersetAdminInRule
		clustersetAdminToSubjects := syncRoleClustersetSubjects(ctx, r.client, r.clusterroleToClustersetAdmin)
		r.clustersetAdminToSubject.SetMap(clustersetAdminToSubjects)
	}

	if clustersetViewInRule.Len() != 0 {
		r.clusterroleToClustersetView[clusterrole.Name] = clustersetViewInRule
		clustersetViewToSubjects := syncRoleClustersetSubjects(ctx, r.client, r.clusterroleToClustersetView)
		r.clustersetViewToSubject.SetMap(clustersetViewToSubjects)
	}

	return ctrl.Result{}, nil
}

// syncRoleClustersetSubjects generate clusterset to subjects map
func syncRoleClustersetSubjects(ctx context.Context, client client.Client, clusterroleToClusterset map[string]sets.String) map[string][]rbacv1.Subject {
	curClustersetToRoles := generateClustersetToClusterroles(clusterroleToClusterset)
	var clustersetToSubjects = make(map[string][]rbacv1.Subject)
	for curClusterset, curClusterRoles := range curClustersetToRoles {
		var clustersetSubjects []rbacv1.Subject
		for _, curClusterRole := range curClusterRoles {
			subjects, err := getClusterroleSubject(ctx, client, curClusterRole)
			if err != nil {
				klog.Errorf("Failed to get clusterrole subject. clusterrole: %v, error:%v", curClusterRole, err)
				continue
			}
			clustersetSubjects = utils.Mergesubjects(clustersetSubjects, subjects)
		}
		clustersetToSubjects[curClusterset] = clustersetSubjects
	}
	return clustersetToSubjects
}

// getClusterroleSubject generate all subject that related to clusterrole
func getClusterroleSubject(ctx context.Context, client client.Client, clusterroleName string) ([]rbacv1.Subject, error) {
	var subjects []rbacv1.Subject
	clusterrolebindinglist := &rbacv1.ClusterRoleBindingList{}
	err := client.List(ctx, clusterrolebindinglist)
	if err != nil {
		return nil, err
	}

	for _, clusterrolebinding := range clusterrolebindinglist.Items {
		if _, ok := clusterrolebinding.Labels[ClusterSetLabel]; ok {
			continue
		}
		if clusterrolebinding.RoleRef.APIGroup == rbacv1.GroupName && clusterrolebinding.RoleRef.Kind == "ClusterRole" && clusterrolebinding.RoleRef.Name == clusterroleName {
			subjects = utils.Mergesubjects(subjects, clusterrolebinding.Subjects)
		}
	}
	return subjects, nil
}

func generateClustersetToClusterroles(clusterroleToClusterset map[string]sets.String) map[string][]string {
	var clustersetToClusterroles = make(map[string][]string)
	for currole, cursets := range clusterroleToClusterset {
		for curset := range cursets {
			clustersetToClusterroles[curset] = append(clustersetToClusterroles[curset], currole)
		}
	}
	return clustersetToClusterroles
}
