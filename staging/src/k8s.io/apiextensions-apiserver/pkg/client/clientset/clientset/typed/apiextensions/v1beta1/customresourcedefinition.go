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

package v1beta1

import (
	"context"
	"time"

	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	scheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CustomResourceDefinitionsGetter has a method to return a CustomResourceDefinitionInterface.
// A group's client should implement this interface.
type CustomResourceDefinitionsGetter interface {
	CustomResourceDefinitions() CustomResourceDefinitionInterface
}

// CustomResourceDefinitionInterface has methods to work with CustomResourceDefinition resources.
type CustomResourceDefinitionInterface interface {
	Create(ctx context.Context, customResourceDefinition *v1beta1.CustomResourceDefinition, opts v1.CreateOptions) (*v1beta1.CustomResourceDefinition, error)
	Update(ctx context.Context, customResourceDefinition *v1beta1.CustomResourceDefinition, opts v1.UpdateOptions) (*v1beta1.CustomResourceDefinition, error)
	UpdateStatus(ctx context.Context, customResourceDefinition *v1beta1.CustomResourceDefinition, opts v1.UpdateOptions) (*v1beta1.CustomResourceDefinition, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.CustomResourceDefinition, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.CustomResourceDefinitionList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.CustomResourceDefinition, err error)
	CustomResourceDefinitionExpansion
}

// customResourceDefinitions implements CustomResourceDefinitionInterface
type customResourceDefinitions struct {
	client rest.Interface
}

// newCustomResourceDefinitions returns a CustomResourceDefinitions
func newCustomResourceDefinitions(c *ApiextensionsV1beta1Client) *customResourceDefinitions {
	return &customResourceDefinitions{
		client: c.RESTClient(),
	}
}

// Get takes name of the customResourceDefinition, and returns the corresponding customResourceDefinition object, and an error if there is any.
func (c *customResourceDefinitions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.CustomResourceDefinition, err error) {
	result = &v1beta1.CustomResourceDefinition{}
	err = c.client.Get().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CustomResourceDefinitions that match those selectors.
func (c *customResourceDefinitions) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.CustomResourceDefinitionList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.CustomResourceDefinitionList{}
	err = c.client.Get().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested customResourceDefinitions.
func (c *customResourceDefinitions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a customResourceDefinition and creates it.  Returns the server's representation of the customResourceDefinition, and an error, if there is any.
func (c *customResourceDefinitions) Create(ctx context.Context, customResourceDefinition *v1beta1.CustomResourceDefinition, opts v1.CreateOptions) (result *v1beta1.CustomResourceDefinition, err error) {
	result = &v1beta1.CustomResourceDefinition{}
	err = c.client.Post().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(customResourceDefinition).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a customResourceDefinition and updates it. Returns the server's representation of the customResourceDefinition, and an error, if there is any.
func (c *customResourceDefinitions) Update(ctx context.Context, customResourceDefinition *v1beta1.CustomResourceDefinition, opts v1.UpdateOptions) (result *v1beta1.CustomResourceDefinition, err error) {
	result = &v1beta1.CustomResourceDefinition{}
	err = c.client.Put().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		Name(customResourceDefinition.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(customResourceDefinition).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *customResourceDefinitions) UpdateStatus(ctx context.Context, customResourceDefinition *v1beta1.CustomResourceDefinition, opts v1.UpdateOptions) (result *v1beta1.CustomResourceDefinition, err error) {
	result = &v1beta1.CustomResourceDefinition{}
	err = c.client.Put().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		Name(customResourceDefinition.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(customResourceDefinition).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the customResourceDefinition and deletes it. Returns an error if one occurs.
func (c *customResourceDefinitions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *customResourceDefinitions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched customResourceDefinition.
func (c *customResourceDefinitions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.CustomResourceDefinition, err error) {
	result = &v1beta1.CustomResourceDefinition{}
	err = c.client.Patch(pt).
		GroupVersion(v1beta1.SchemeGroupVersion).
		Resource("customresourcedefinitions").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
