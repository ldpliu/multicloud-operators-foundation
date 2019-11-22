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

// PlacementBindingLister helps list PlacementBindings.
type PlacementBindingLister interface {
	// List lists all PlacementBindings in the indexer.
	List(selector labels.Selector) (ret []*mcm.PlacementBinding, err error)
	// PlacementBindings returns an object that can list and get PlacementBindings.
	PlacementBindings(namespace string) PlacementBindingNamespaceLister
	PlacementBindingListerExpansion
}

// placementBindingLister implements the PlacementBindingLister interface.
type placementBindingLister struct {
	indexer cache.Indexer
}

// NewPlacementBindingLister returns a new PlacementBindingLister.
func NewPlacementBindingLister(indexer cache.Indexer) PlacementBindingLister {
	return &placementBindingLister{indexer: indexer}
}

// List lists all PlacementBindings in the indexer.
func (s *placementBindingLister) List(selector labels.Selector) (ret []*mcm.PlacementBinding, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*mcm.PlacementBinding))
	})
	return ret, err
}

// PlacementBindings returns an object that can list and get PlacementBindings.
func (s *placementBindingLister) PlacementBindings(namespace string) PlacementBindingNamespaceLister {
	return placementBindingNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PlacementBindingNamespaceLister helps list and get PlacementBindings.
type PlacementBindingNamespaceLister interface {
	// List lists all PlacementBindings in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*mcm.PlacementBinding, err error)
	// Get retrieves the PlacementBinding from the indexer for a given namespace and name.
	Get(name string) (*mcm.PlacementBinding, error)
	PlacementBindingNamespaceListerExpansion
}

// placementBindingNamespaceLister implements the PlacementBindingNamespaceLister
// interface.
type placementBindingNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all PlacementBindings in the indexer for a given namespace.
func (s placementBindingNamespaceLister) List(selector labels.Selector) (ret []*mcm.PlacementBinding, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*mcm.PlacementBinding))
	})
	return ret, err
}

// Get retrieves the PlacementBinding from the indexer for a given namespace and name.
func (s placementBindingNamespaceLister) Get(name string) (*mcm.PlacementBinding, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(mcm.Resource("placementbinding"), name)
	}
	return obj.(*mcm.PlacementBinding), nil
}
