package cache

import (
	"time"

	clusterinformerv1alpha1 "github.com/open-cluster-management/api/client/cluster/informers/externalversions/cluster/v1alpha1"
	clusterv1alpha1lister "github.com/open-cluster-management/api/client/cluster/listers/cluster/v1alpha1"
	clusterv1alpha1 "github.com/open-cluster-management/api/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/user"
	rbacv1informers "k8s.io/client-go/informers/rbac/v1"
)

// ClusterSetLister enforces ability to enumerate clusterSet based on role
type ClusterSetLister interface {
	// List returns the list of ManagedClusterSet items that the user can access
	List(user user.Info, selector labels.Selector) (*clusterv1alpha1.ManagedClusterSetList, error)
}

type ClusterSetCache struct {
	cache            *AuthCache
	clusterSetLister clusterv1alpha1lister.ManagedClusterSetLister
}

func NewClusterSetCache(clusterSetInformer clusterinformerv1alpha1.ManagedClusterSetInformer,
	clusterRoleInformer rbacv1informers.ClusterRoleInformer,
	clusterRolebindingInformer rbacv1informers.ClusterRoleBindingInformer,
) *ClusterSetCache {
	clusterSetCache := &ClusterSetCache{
		clusterSetLister: clusterSetInformer.Lister(),
	}
	authCache := NewAuthCache(clusterRoleInformer, clusterRolebindingInformer,
		"cluster.open-cluster-management.io", "managedclustersets",
		clusterSetInformer.Informer(),
		clusterSetCache.ListResources,
	)
	clusterSetCache.cache = authCache

	return clusterSetCache
}

func (c *ClusterSetCache) ListResources() (sets.String, error) {
	allClusterSets := sets.String{}
	clusterSets, err := c.clusterSetLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, clusterSet := range clusterSets {
		allClusterSets.Insert(clusterSet.Name)
	}
	return allClusterSets, nil
}

func (c *ClusterSetCache) List(userInfo user.Info, selector labels.Selector) (*clusterv1alpha1.ManagedClusterSetList, error) {
	names := c.cache.listNames(userInfo)

	clusterSetList := &clusterv1alpha1.ManagedClusterSetList{}
	for key := range names {
		clusterSet, err := c.clusterSetLister.Get(key)
		if errors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		if !selector.Matches(labels.Set(clusterSet.Labels)) {
			continue
		}
		clusterSetList.Items = append(clusterSetList.Items, *clusterSet)
	}
	return clusterSetList, nil
}

func (c *ClusterSetCache) ListObjects(userInfo user.Info) (runtime.Object, error) {
	return c.List(userInfo, labels.Everything())
}

func (c *ClusterSetCache) Get(name string) (runtime.Object, error) {
	return c.clusterSetLister.Get(name)
}

func (c *ClusterSetCache) ConvertResource(name string) runtime.Object {
	clusterSet, err := c.clusterSetLister.Get(name)
	if err != nil {
		clusterSet = &clusterv1alpha1.ManagedClusterSet{ObjectMeta: metav1.ObjectMeta{Name: name}}
	}

	return clusterSet
}

func (c *ClusterSetCache) RemoveWatcher(w CacheWatcher) {
	c.cache.RemoveWatcher(w)
}

func (c *ClusterSetCache) AddWatcher(w CacheWatcher) {
	c.cache.AddWatcher(w)
}

// Run begins watching and synchronizing the cache
func (c *ClusterSetCache) Run(period time.Duration) {
	go utilwait.Forever(func() { c.cache.synchronize() }, period)
}
