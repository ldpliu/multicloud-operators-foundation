package clusterrolebinding

import (
	"testing"

	clusterv1alpha1 "github.com/open-cluster-management/api/cluster/v1alpha1"
	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/helpers"
	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	scheme = runtime.NewScheme()
)

func newTestReconciler(clusterroleToClustersetAdmin map[string]sets.String, clustersetAdminToSubject *helpers.ClustersetSubjectsMapper, clusterroleToClustersetView map[string]sets.String, clustersetViewToSubject *helpers.ClustersetSubjectsMapper, roleobjs, rolebindingobjs []runtime.Object) *Reconciler {
	objs := roleobjs
	objs = append(objs, rolebindingobjs...)

	r := &Reconciler{
		client:                       fake.NewFakeClient(objs...),
		scheme:                       scheme,
		clusterroleToClustersetAdmin: clusterroleToClustersetAdmin,
		clustersetAdminToSubject:     clustersetAdminToSubject,
		clustersetViewToSubject:      clustersetViewToSubject,
		clusterroleToClustersetView:  clusterroleToClustersetView,
	}
	return r
}

func generateClustersetSubjectMap() *helpers.ClustersetSubjectsMapper {
	clusterserSubject := make(map[string][]rbacv1.Subject)
	subjects1 := []rbacv1.Subject{
		{Kind: "k1", APIGroup: "a1", Name: "n1"}}
	clusterserSubject["s1"] = subjects1
	clustersetToSubject := helpers.NewClustersetSubjectsMapper()
	clustersetToSubject.SetMap(clusterserSubject)
	return clustersetToSubject
}

func TestReconcile(t *testing.T) {
	tests := []struct {
		name                         string
		clusterroleToClustersetAdmin map[string]sets.String
		clustersetAdminToSubject     *helpers.ClustersetSubjectsMapper
		clusterroleToClustersetView  map[string]sets.String
		clustersetViewToSubject      *helpers.ClustersetSubjectsMapper
		clusterRoleObjs              []runtime.Object
		clusterRoleBindingObjs       []runtime.Object
		expectedAdminMapperData      map[string][]rbacv1.Subject
		expectedViewMapperData       map[string][]rbacv1.Subject
		req                          reconcile.Request
	}{
		{
			name:                         "one set in clusterrole",
			clusterroleToClustersetAdmin: make(map[string]sets.String),
			clustersetAdminToSubject:     helpers.NewClustersetSubjectsMapper(),
			clusterroleToClustersetView:  make(map[string]sets.String),
			clustersetViewToSubject:      helpers.NewClustersetSubjectsMapper(),
			clusterRoleObjs: []runtime.Object{
				&rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRole1",
					},
					Rules: []rbacv1.PolicyRule{
						{Verbs: []string{"update"}, APIGroups: []string{clusterv1alpha1.GroupName}, Resources: []string{"managedclustersets"}},
					},
				},
			},
			clusterRoleBindingObjs: []runtime.Object{
				&rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRolebinding1",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "clusterRole1",
					},
					Subjects: []rbacv1.Subject{
						{Kind: "k1", APIGroup: "a1", Name: "n1"},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole1",
				},
			},
			expectedAdminMapperData: map[string][]rbacv1.Subject{"*": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}},
			expectedViewMapperData:  map[string][]rbacv1.Subject{},
		},
		{
			name:                         "two clusterrole",
			clusterroleToClustersetAdmin: make(map[string]sets.String),
			clustersetAdminToSubject:     helpers.NewClustersetSubjectsMapper(),
			clusterroleToClustersetView:  make(map[string]sets.String),
			clustersetViewToSubject:      helpers.NewClustersetSubjectsMapper(),
			clusterRoleObjs: []runtime.Object{
				&rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRole1",
					},
					Rules: []rbacv1.PolicyRule{
						{Verbs: []string{"update"}, APIGroups: []string{clusterv1alpha1.GroupName}, Resources: []string{"managedclustersets"}, ResourceNames: []string{"s1", "s2"}},
					},
				},
			},
			clusterRoleBindingObjs: []runtime.Object{
				&rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRolebinding1",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "clusterRole1",
					},
					Subjects: []rbacv1.Subject{
						{Kind: "k1", APIGroup: "a1", Name: "n1"},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole1",
				},
			},
			expectedAdminMapperData: map[string][]rbacv1.Subject{"s1": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}, "s2": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}},
			expectedViewMapperData:  map[string][]rbacv1.Subject{},
		},
		{
			name: "update clusterrole",
			clusterroleToClustersetAdmin: map[string]sets.String{
				"clusterRole1": sets.NewString("s1", "s3"),
			},
			clustersetAdminToSubject:    generateClustersetSubjectMap(),
			clusterroleToClustersetView: make(map[string]sets.String),
			clustersetViewToSubject:     helpers.NewClustersetSubjectsMapper(),
			clusterRoleObjs: []runtime.Object{
				&rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRole1",
					},
					Rules: []rbacv1.PolicyRule{
						{Verbs: []string{"update"}, APIGroups: []string{clusterv1alpha1.GroupName}, Resources: []string{"managedclustersets"}, ResourceNames: []string{"s1", "s2"}},
					},
				},
			},
			clusterRoleBindingObjs: []runtime.Object{
				&rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRolebinding1",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "clusterRole1",
					},
					Subjects: []rbacv1.Subject{
						{Kind: "k1", APIGroup: "a1", Name: "n1"},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole1",
				},
			},
			expectedAdminMapperData: map[string][]rbacv1.Subject{"s1": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}, "s2": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}},
			expectedViewMapperData:  map[string][]rbacv1.Subject{},
		},
		{
			name: "delete clusterrolebinding",
			clusterroleToClustersetAdmin: map[string]sets.String{
				"clusterRole1": sets.NewString("s1", "s3"),
			},
			clustersetAdminToSubject:    generateClustersetSubjectMap(),
			clusterroleToClustersetView: make(map[string]sets.String),
			clustersetViewToSubject:     helpers.NewClustersetSubjectsMapper(),
			clusterRoleObjs: []runtime.Object{
				&rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRole1",
					},
					Rules: []rbacv1.PolicyRule{
						{Verbs: []string{"update"}, APIGroups: []string{clusterv1alpha1.GroupName}, Resources: []string{"managedclustersets"}, ResourceNames: []string{"s1", "s2"}},
					},
				},
			},
			clusterRoleBindingObjs: []runtime.Object{},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole1",
				},
			},
			expectedAdminMapperData: map[string][]rbacv1.Subject{"s1": {}, "s2": {}},
			expectedViewMapperData:  map[string][]rbacv1.Subject{},
		},
		{
			name: "remove subjects",
			clusterroleToClustersetAdmin: map[string]sets.String{
				"clusterRole1": sets.NewString("s1", "s3"),
			},
			clustersetAdminToSubject:    generateClustersetSubjectMap(),
			clusterroleToClustersetView: make(map[string]sets.String),
			clustersetViewToSubject:     helpers.NewClustersetSubjectsMapper(),
			clusterRoleObjs: []runtime.Object{
				&rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRole1",
					},
					Rules: []rbacv1.PolicyRule{
						{Verbs: []string{"update"}, APIGroups: []string{clusterv1alpha1.GroupName}, Resources: []string{"managedclustersets"}, ResourceNames: []string{"s1", "s2"}},
					},
				},
			},
			clusterRoleBindingObjs: []runtime.Object{
				&rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRolebinding1",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "clusterRole1",
					},
					Subjects: []rbacv1.Subject{},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole1",
				},
			},
			expectedAdminMapperData: map[string][]rbacv1.Subject{"s1": {}, "s2": {}},
			expectedViewMapperData:  map[string][]rbacv1.Subject{},
		},
		{
			name:                         "one set in clusterrole for view",
			clusterroleToClustersetAdmin: make(map[string]sets.String),
			clustersetAdminToSubject:     helpers.NewClustersetSubjectsMapper(),
			clusterroleToClustersetView:  make(map[string]sets.String),
			clustersetViewToSubject:      helpers.NewClustersetSubjectsMapper(),
			clusterRoleObjs: []runtime.Object{
				&rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRole3",
					},
					Rules: []rbacv1.PolicyRule{
						{Verbs: []string{"get"}, APIGroups: []string{clusterv1alpha1.GroupName}, Resources: []string{"managedclustersets"}},
					},
				},
			},
			clusterRoleBindingObjs: []runtime.Object{
				&rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: "clusterRolebinding1",
					},
					RoleRef: rbacv1.RoleRef{
						APIGroup: rbacv1.GroupName,
						Kind:     "ClusterRole",
						Name:     "clusterRole3",
					},
					Subjects: []rbacv1.Subject{
						{Kind: "k1", APIGroup: "a1", Name: "n1"},
					},
				},
			},
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole3",
				},
			},
			expectedAdminMapperData: map[string][]rbacv1.Subject{},
			expectedViewMapperData:  map[string][]rbacv1.Subject{"*": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}},
		},
	}

	for _, test := range tests {
		r := newTestReconciler(test.clusterroleToClustersetAdmin, test.clustersetAdminToSubject, test.clusterroleToClustersetView, test.clustersetViewToSubject, test.clusterRoleBindingObjs, test.clusterRoleObjs)
		r.Reconcile(test.req)
		validateResult(t, test.name, r, test.expectedAdminMapperData, test.expectedViewMapperData)
	}
}

func validateResult(t *testing.T, name string, r *Reconciler, expectedAdminMapperData map[string][]rbacv1.Subject, expectedViewMapperData map[string][]rbacv1.Subject) {
	mapperAdminData := r.clustersetAdminToSubject.GetMap()
	mapperViewData := r.clustersetViewToSubject.GetMap()
	if len(mapperAdminData) != len(expectedAdminMapperData) {
		t.Errorf("Expect admin map.len is not same as result, case: %v, return Map:%v, expect Map: %v", name, mapperAdminData, expectedAdminMapperData)
	} else {
		for clusterSet, subjects := range mapperAdminData {
			if _, ok := expectedAdminMapperData[clusterSet]; !ok {
				t.Errorf("Expect admin map is not same as result, case: %v, return Map:%v, expect Map: %v", name, mapperAdminData, expectedAdminMapperData)
				return
			}
			if !utils.EqualSubjects(expectedAdminMapperData[clusterSet], subjects) {
				t.Errorf("Expect admin map is not same as result, case: %v,return Map:%v, expect Map: %v", name, mapperAdminData, expectedAdminMapperData)
				return
			}
		}
	}

	if len(mapperViewData) != len(expectedViewMapperData) {
		t.Errorf("Expect view map.len is not same as result, case: %v, return Map:%v, expect Map: %v", name, mapperViewData, expectedViewMapperData)
	} else {
		for clusterSet, subjects := range mapperViewData {
			if _, ok := expectedViewMapperData[clusterSet]; !ok {
				t.Errorf("Expect view map is not same as result, case: %v, return Map:%v, expect Map: %v", name, mapperViewData, expectedViewMapperData)
				return
			}
			if !utils.EqualSubjects(expectedViewMapperData[clusterSet], subjects) {
				t.Errorf("Expect view map is not same as result, case: %v, return Map:%v, expect Map: %v", name, mapperViewData, expectedViewMapperData)
				return
			}
		}
	}

}
