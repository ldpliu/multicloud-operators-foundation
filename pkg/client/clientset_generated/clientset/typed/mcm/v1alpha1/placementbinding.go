// Licensed Materials - Property of IBM
// (c) Copyright IBM Corporation 2018. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.ibm.com/IBMPrivateCloud/multicloud-operators-foundation/pkg/apis/mcm/v1alpha1"
	scheme "github.ibm.com/IBMPrivateCloud/multicloud-operators-foundation/pkg/client/clientset_generated/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PlacementBindingsGetter has a method to return a PlacementBindingInterface.
// A group's client should implement this interface.
type PlacementBindingsGetter interface {
	PlacementBindings(namespace string) PlacementBindingInterface
}

// PlacementBindingInterface has methods to work with PlacementBinding resources.
type PlacementBindingInterface interface {
	Create(*v1alpha1.PlacementBinding) (*v1alpha1.PlacementBinding, error)
	Update(*v1alpha1.PlacementBinding) (*v1alpha1.PlacementBinding, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.PlacementBinding, error)
	List(opts v1.ListOptions) (*v1alpha1.PlacementBindingList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PlacementBinding, err error)
	PlacementBindingExpansion
}

// placementBindings implements PlacementBindingInterface
type placementBindings struct {
	client rest.Interface
	ns     string
}

// newPlacementBindings returns a PlacementBindings
func newPlacementBindings(c *McmV1alpha1Client, namespace string) *placementBindings {
	return &placementBindings{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the placementBinding, and returns the corresponding placementBinding object, and an error if there is any.
func (c *placementBindings) Get(name string, options v1.GetOptions) (result *v1alpha1.PlacementBinding, err error) {
	result = &v1alpha1.PlacementBinding{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("placementbindings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PlacementBindings that match those selectors.
func (c *placementBindings) List(opts v1.ListOptions) (result *v1alpha1.PlacementBindingList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.PlacementBindingList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("placementbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested placementBindings.
func (c *placementBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("placementbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a placementBinding and creates it.  Returns the server's representation of the placementBinding, and an error, if there is any.
func (c *placementBindings) Create(placementBinding *v1alpha1.PlacementBinding) (result *v1alpha1.PlacementBinding, err error) {
	result = &v1alpha1.PlacementBinding{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("placementbindings").
		Body(placementBinding).
		Do().
		Into(result)
	return
}

// Update takes the representation of a placementBinding and updates it. Returns the server's representation of the placementBinding, and an error, if there is any.
func (c *placementBindings) Update(placementBinding *v1alpha1.PlacementBinding) (result *v1alpha1.PlacementBinding, err error) {
	result = &v1alpha1.PlacementBinding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("placementbindings").
		Name(placementBinding.Name).
		Body(placementBinding).
		Do().
		Into(result)
	return
}

// Delete takes name of the placementBinding and deletes it. Returns an error if one occurs.
func (c *placementBindings) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("placementbindings").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *placementBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("placementbindings").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched placementBinding.
func (c *placementBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PlacementBinding, err error) {
	result = &v1alpha1.PlacementBinding{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("placementbindings").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
