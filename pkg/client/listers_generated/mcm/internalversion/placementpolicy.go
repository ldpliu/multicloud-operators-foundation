// Licensed Materials - Property of IBM
// (c) Copyright IBM Corporation 2018. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.

// Code generated by lister-gen. DO NOT EDIT.

package internalversion

import (
	mcm "github.ibm.com/IBMPrivateCloud/multicloud-operators-foundation/pkg/apis/mcm"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PlacementPolicyLister helps list PlacementPolicies.
type PlacementPolicyLister interface {
	// List lists all PlacementPolicies in the indexer.
	List(selector labels.Selector) (ret []*mcm.PlacementPolicy, err error)
	// PlacementPolicies returns an object that can list and get PlacementPolicies.
	PlacementPolicies(namespace string) PlacementPolicyNamespaceLister
	PlacementPolicyListerExpansion
}

// placementPolicyLister implements the PlacementPolicyLister interface.
type placementPolicyLister struct {
	indexer cache.Indexer
}

// NewPlacementPolicyLister returns a new PlacementPolicyLister.
func NewPlacementPolicyLister(indexer cache.Indexer) PlacementPolicyLister {
	return &placementPolicyLister{indexer: indexer}
}

// List lists all PlacementPolicies in the indexer.
func (s *placementPolicyLister) List(selector labels.Selector) (ret []*mcm.PlacementPolicy, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*mcm.PlacementPolicy))
	})
	return ret, err
}

// PlacementPolicies returns an object that can list and get PlacementPolicies.
func (s *placementPolicyLister) PlacementPolicies(namespace string) PlacementPolicyNamespaceLister {
	return placementPolicyNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PlacementPolicyNamespaceLister helps list and get PlacementPolicies.
type PlacementPolicyNamespaceLister interface {
	// List lists all PlacementPolicies in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*mcm.PlacementPolicy, err error)
	// Get retrieves the PlacementPolicy from the indexer for a given namespace and name.
	Get(name string) (*mcm.PlacementPolicy, error)
	PlacementPolicyNamespaceListerExpansion
}

// placementPolicyNamespaceLister implements the PlacementPolicyNamespaceLister
// interface.
type placementPolicyNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all PlacementPolicies in the indexer for a given namespace.
func (s placementPolicyNamespaceLister) List(selector labels.Selector) (ret []*mcm.PlacementPolicy, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*mcm.PlacementPolicy))
	})
	return ret, err
}

// Get retrieves the PlacementPolicy from the indexer for a given namespace and name.
func (s placementPolicyNamespaceLister) Get(name string) (*mcm.PlacementPolicy, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(mcm.Resource("placementpolicy"), name)
	}
	return obj.(*mcm.PlacementPolicy), nil
}
