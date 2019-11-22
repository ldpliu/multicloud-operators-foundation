// Licensed Materials - Property of IBM
// (c) Copyright IBM Corporation 2018. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "github.ibm.com/IBMPrivateCloud/multicloud-operators-foundation/pkg/client/informers_generated/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// ClusterJoinRequests returns a ClusterJoinRequestInformer.
	ClusterJoinRequests() ClusterJoinRequestInformer
	// ClusterStatuses returns a ClusterStatusInformer.
	ClusterStatuses() ClusterStatusInformer
	// PlacementBindings returns a PlacementBindingInformer.
	PlacementBindings() PlacementBindingInformer
	// PlacementPolicies returns a PlacementPolicyInformer.
	PlacementPolicies() PlacementPolicyInformer
	// ResourceViews returns a ResourceViewInformer.
	ResourceViews() ResourceViewInformer
	// Works returns a WorkInformer.
	Works() WorkInformer
	// WorkSets returns a WorkSetInformer.
	WorkSets() WorkSetInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// ClusterJoinRequests returns a ClusterJoinRequestInformer.
func (v *version) ClusterJoinRequests() ClusterJoinRequestInformer {
	return &clusterJoinRequestInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// ClusterStatuses returns a ClusterStatusInformer.
func (v *version) ClusterStatuses() ClusterStatusInformer {
	return &clusterStatusInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PlacementBindings returns a PlacementBindingInformer.
func (v *version) PlacementBindings() PlacementBindingInformer {
	return &placementBindingInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PlacementPolicies returns a PlacementPolicyInformer.
func (v *version) PlacementPolicies() PlacementPolicyInformer {
	return &placementPolicyInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ResourceViews returns a ResourceViewInformer.
func (v *version) ResourceViews() ResourceViewInformer {
	return &resourceViewInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Works returns a WorkInformer.
func (v *version) Works() WorkInformer {
	return &workInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// WorkSets returns a WorkSetInformer.
func (v *version) WorkSets() WorkSetInformer {
	return &workSetInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
