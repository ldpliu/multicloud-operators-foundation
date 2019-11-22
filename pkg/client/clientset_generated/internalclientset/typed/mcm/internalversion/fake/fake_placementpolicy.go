// Licensed Materials - Property of IBM
// (c) Copyright IBM Corporation 2018. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	mcm "github.ibm.com/IBMPrivateCloud/multicloud-operators-foundation/pkg/apis/mcm"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePlacementPolicies implements PlacementPolicyInterface
type FakePlacementPolicies struct {
	Fake *FakeMcm
	ns   string
}

var placementpoliciesResource = schema.GroupVersionResource{Group: "mcm.ibm.com", Version: "", Resource: "placementpolicies"}

var placementpoliciesKind = schema.GroupVersionKind{Group: "mcm.ibm.com", Version: "", Kind: "PlacementPolicy"}

// Get takes name of the placementPolicy, and returns the corresponding placementPolicy object, and an error if there is any.
func (c *FakePlacementPolicies) Get(name string, options v1.GetOptions) (result *mcm.PlacementPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(placementpoliciesResource, c.ns, name), &mcm.PlacementPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*mcm.PlacementPolicy), err
}

// List takes label and field selectors, and returns the list of PlacementPolicies that match those selectors.
func (c *FakePlacementPolicies) List(opts v1.ListOptions) (result *mcm.PlacementPolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(placementpoliciesResource, placementpoliciesKind, c.ns, opts), &mcm.PlacementPolicyList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &mcm.PlacementPolicyList{ListMeta: obj.(*mcm.PlacementPolicyList).ListMeta}
	for _, item := range obj.(*mcm.PlacementPolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested placementPolicies.
func (c *FakePlacementPolicies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(placementpoliciesResource, c.ns, opts))

}

// Create takes the representation of a placementPolicy and creates it.  Returns the server's representation of the placementPolicy, and an error, if there is any.
func (c *FakePlacementPolicies) Create(placementPolicy *mcm.PlacementPolicy) (result *mcm.PlacementPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(placementpoliciesResource, c.ns, placementPolicy), &mcm.PlacementPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*mcm.PlacementPolicy), err
}

// Update takes the representation of a placementPolicy and updates it. Returns the server's representation of the placementPolicy, and an error, if there is any.
func (c *FakePlacementPolicies) Update(placementPolicy *mcm.PlacementPolicy) (result *mcm.PlacementPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(placementpoliciesResource, c.ns, placementPolicy), &mcm.PlacementPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*mcm.PlacementPolicy), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePlacementPolicies) UpdateStatus(placementPolicy *mcm.PlacementPolicy) (*mcm.PlacementPolicy, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(placementpoliciesResource, "status", c.ns, placementPolicy), &mcm.PlacementPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*mcm.PlacementPolicy), err
}

// Delete takes name of the placementPolicy and deletes it. Returns an error if one occurs.
func (c *FakePlacementPolicies) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(placementpoliciesResource, c.ns, name), &mcm.PlacementPolicy{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePlacementPolicies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(placementpoliciesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &mcm.PlacementPolicyList{})
	return err
}

// Patch applies the patch and returns the patched placementPolicy.
func (c *FakePlacementPolicies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *mcm.PlacementPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(placementpoliciesResource, c.ns, name, pt, data, subresources...), &mcm.PlacementPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*mcm.PlacementPolicy), err
}
