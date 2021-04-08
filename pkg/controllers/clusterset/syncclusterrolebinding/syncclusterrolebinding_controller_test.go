package syncclusterrolebinding

import (
	"context"
	"reflect"
	"testing"

	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/helpers"
	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	scheme = runtime.NewScheme()
)

func newTestReconciler(clustersetToClusters *helpers.ClusterSetMapper, clustersetAdminToSubject, clustersetViewToSubject *helpers.ClustersetSubjectsMapper) *Reconciler {
	cb := generateRequiredClusterRoleBinding("c0", nil, "admin")
	objs := []runtime.Object{cb}
	r := &Reconciler{
		client:                   fake.NewFakeClient(objs...),
		scheme:                   scheme,
		clustersetToClusters:     clustersetToClusters,
		clustersetAdminToSubject: clustersetAdminToSubject,
		clustersetViewToSubject:  clustersetViewToSubject,
	}
	return r
}

func generateClustersetToClusters(ms map[string]sets.String) *helpers.ClusterSetMapper {
	clustersetToClusters := helpers.NewClusterSetMapper()
	for s, c := range ms {
		clustersetToClusters.UpdateClusterSetByObjects(s, c)
	}
	return clustersetToClusters
}

func generateClustersetToSubjects(mc map[string][]rbacv1.Subject) *helpers.ClustersetSubjectsMapper {
	clustersetToSubject := helpers.NewClustersetSubjectsMapper()
	clustersetToSubject.SetMap(mc)
	return clustersetToSubject
}

func TestReconcile(t *testing.T) {
	ctc1 := generateClustersetToClusters(nil)
	cts1 := generateClustersetToSubjects(nil)

	ms2 := map[string]sets.String{"cs1": sets.NewString("c1", "c2")}
	ctc2 := generateClustersetToClusters(ms2)

	mc2 := map[string][]rbacv1.Subject{"cs1": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}}
	cts2 := generateClustersetToSubjects(mc2)

	tests := []struct {
		name                     string
		clustersetToClusters     *helpers.ClusterSetMapper
		clustersetAdminToSubject *helpers.ClustersetSubjectsMapper
		clustersetViewToSubject  *helpers.ClustersetSubjectsMapper
		req                      reconcile.Request
		clusterrolebindingName   string
		exist                    bool
	}{
		{
			name:                     "init:",
			clustersetToClusters:     ctc1,
			clustersetAdminToSubject: cts1,
			clustersetViewToSubject:  cts1,
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole1",
				},
			},
			clusterrolebindingName: utils.GenerateClustersetClusterRoleBindingName("c1", "admin"),
			exist:                  false,
		},
		{
			name:                     "delete c0:",
			clustersetToClusters:     ctc1,
			clustersetAdminToSubject: cts1,
			clustersetViewToSubject:  cts1,
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole1",
				},
			},
			clusterrolebindingName: utils.GenerateClustersetClusterRoleBindingName("c0", "admin"),
			exist:                  false,
		},
		{
			name:                     "need create:",
			clustersetToClusters:     ctc2,
			clustersetAdminToSubject: cts2,
			clustersetViewToSubject:  cts2,
			req: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: "clusterRole2",
				},
			},
			clusterrolebindingName: utils.GenerateClustersetClusterRoleBindingName("c1", "admin"),
			exist:                  true,
		},
	}

	for _, test := range tests {
		r := newTestReconciler(test.clustersetToClusters, test.clustersetAdminToSubject, test.clustersetViewToSubject)
		r.Reconcile(test.req)
		validateResult(t, r, test.clusterrolebindingName, test.exist)
	}
}

func validateResult(t *testing.T, r *Reconciler, clusterrolebindingName string, exist bool) {
	ctx := context.Background()
	clusterrolebinding := &rbacv1.ClusterRoleBinding{}
	r.client.Get(ctx, types.NamespacedName{Name: clusterrolebindingName}, clusterrolebinding)
	if exist && clusterrolebinding == nil {
		t.Errorf("Failed to apply clusterrolebinding")
	}
}

func Test_getClusterNameInClusterrolebinding(t *testing.T) {
	type args struct {
		clusterrolebindingName string
		role                   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"case1",
			args{
				clusterrolebindingName: "open-cluster-management:managedclusterset:admin:managedcluster:managedcluster1",
				role:                   "admin",
			},
			"managedcluster1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getClusterNameInClusterrolebinding(tt.args.clusterrolebindingName, tt.args.role); got != tt.want {
				t.Errorf("getClusterNameInClusterrolebinding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateClusterSubjectMap(t *testing.T) {
	ctc1 := generateClustersetToClusters(nil)

	ms2 := map[string]sets.String{"cs1": sets.NewString("c1", "c2")}
	ctc2 := generateClustersetToClusters(ms2)

	mc2 := map[string][]rbacv1.Subject{"cs1": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}}
	cts2 := generateClustersetToSubjects(mc2)

	mc3 := map[string][]rbacv1.Subject{"*": {{Kind: "k1", APIGroup: "a1", Name: "n1"}}}
	cts3 := generateClustersetToSubjects(mc3)

	type args struct {
		clustersetToClusters *helpers.ClusterSetMapper
		clustersetToSubject  *helpers.ClustersetSubjectsMapper
	}
	tests := []struct {
		name string
		args args
		want map[string][]rbacv1.Subject
	}{
		{
			name: "no clusters:",
			args: args{
				clustersetToClusters: ctc1,
				clustersetToSubject:  cts2,
			},
			want: map[string][]rbacv1.Subject{},
		},
		{
			name: "need create:",
			args: args{
				clustersetToClusters: ctc2,
				clustersetToSubject:  cts2,
			},
			want: map[string][]rbacv1.Subject{
				"c1": {
					{
						Kind:     "k1",
						APIGroup: "a1",
						Name:     "n1",
					},
				},
				"c2": {
					{
						Kind:     "k1",
						APIGroup: "a1",
						Name:     "n1",
					},
				},
			},
		},
		{
			name: "test all:",
			args: args{
				clustersetToClusters: ctc2,
				clustersetToSubject:  cts3,
			},
			want: map[string][]rbacv1.Subject{
				"c1": {
					{
						Kind:     "k1",
						APIGroup: "a1",
						Name:     "n1",
					},
				},
				"c2": {
					{
						Kind:     "k1",
						APIGroup: "a1",
						Name:     "n1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateClusterSubjectMap(tt.args.clustersetToClusters, tt.args.clustersetToSubject); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateClusterSubjectMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
