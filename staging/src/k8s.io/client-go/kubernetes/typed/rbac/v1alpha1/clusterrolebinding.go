/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	v1alpha1 "k8s.io/api/rbac/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rbacv1alpha1 "k8s.io/client-go/applyconfigurations/rbac/v1alpha1"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// ClusterRoleBindingsGetter has a method to return a ClusterRoleBindingInterface.
// A group's client should implement this interface.
type ClusterRoleBindingsGetter interface {
	ClusterRoleBindings() ClusterRoleBindingInterface
}

// ClusterRoleBindingInterface has methods to work with ClusterRoleBinding resources.
type ClusterRoleBindingInterface interface {
	Create(ctx context.Context, clusterRoleBinding *v1alpha1.ClusterRoleBinding, opts v1.CreateOptions) (*v1alpha1.ClusterRoleBinding, error)
	Update(ctx context.Context, clusterRoleBinding *v1alpha1.ClusterRoleBinding, opts v1.UpdateOptions) (*v1alpha1.ClusterRoleBinding, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ClusterRoleBinding, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ClusterRoleBindingList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ClusterRoleBinding, err error)
	Apply(ctx context.Context, clusterRoleBinding *rbacv1alpha1.ClusterRoleBindingApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.ClusterRoleBinding, err error)
	ClusterRoleBindingExpansion
}

// clusterRoleBindings implements ClusterRoleBindingInterface
type clusterRoleBindings struct {
	client rest.Interface
}

// newClusterRoleBindings returns a ClusterRoleBindings
func newClusterRoleBindings(c *RbacV1alpha1Client) *clusterRoleBindings {
	return &clusterRoleBindings{
		client: c.RESTClient(),
	}
}

// Get takes name of the clusterRoleBinding, and returns the corresponding clusterRoleBinding object, and an error if there is any.
func (c *clusterRoleBindings) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ClusterRoleBinding, err error) {
	result = &v1alpha1.ClusterRoleBinding{}
	err = c.client.Get().
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ClusterRoleBindings that match those selectors.
func (c *clusterRoleBindings) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ClusterRoleBindingList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ClusterRoleBindingList{}
	err = c.client.Get().
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested clusterRoleBindings.
func (c *clusterRoleBindings) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a clusterRoleBinding and creates it.  Returns the server's representation of the clusterRoleBinding, and an error, if there is any.
func (c *clusterRoleBindings) Create(ctx context.Context, clusterRoleBinding *v1alpha1.ClusterRoleBinding, opts v1.CreateOptions) (result *v1alpha1.ClusterRoleBinding, err error) {
	result = &v1alpha1.ClusterRoleBinding{}
	err = c.client.Post().
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(clusterRoleBinding).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a clusterRoleBinding and updates it. Returns the server's representation of the clusterRoleBinding, and an error, if there is any.
func (c *clusterRoleBindings) Update(ctx context.Context, clusterRoleBinding *v1alpha1.ClusterRoleBinding, opts v1.UpdateOptions) (result *v1alpha1.ClusterRoleBinding, err error) {
	result = &v1alpha1.ClusterRoleBinding{}
	err = c.client.Put().
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		Name(clusterRoleBinding.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(clusterRoleBinding).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the clusterRoleBinding and deletes it. Returns an error if one occurs.
func (c *clusterRoleBindings) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *clusterRoleBindings) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched clusterRoleBinding.
func (c *clusterRoleBindings) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ClusterRoleBinding, err error) {
	result = &v1alpha1.ClusterRoleBinding{}
	err = c.client.Patch(pt).
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

// Apply takes the given apply declarative configuration, applies it and returns the applied clusterRoleBinding.
func (c *clusterRoleBindings) Apply(ctx context.Context, clusterRoleBinding *rbacv1alpha1.ClusterRoleBindingApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.ClusterRoleBinding, err error) {
	if clusterRoleBinding == nil {
		return nil, fmt.Errorf("clusterRoleBinding provided to Apply must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	data, err := json.Marshal(clusterRoleBinding)
	if err != nil {
		return nil, err
	}
	name := clusterRoleBinding.Name
	if name == nil {
		return nil, fmt.Errorf("clusterRoleBinding.Name must be provided to Apply")
	}
	result = &v1alpha1.ClusterRoleBinding{}
	err = c.client.Patch(types.ApplyPatchType).
		GroupVersion(v1alpha1.SchemeGroupVersion).
		Resource("clusterrolebindings").
		Name(*name).
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
