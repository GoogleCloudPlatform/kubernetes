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

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1beta1 "k8s.io/api/apps/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/storage"
	appsv1beta1 "k8s.io/client-go/applyconfigurations/apps/v1beta1"
	testing "k8s.io/client-go/testing"
)

// FakeDeployments implements DeploymentInterface
type FakeDeployments struct {
	Fake *FakeAppsV1beta1
	ns   string
}

var deploymentsResource = schema.GroupVersionResource{Group: "apps", Version: "v1beta1", Resource: "deployments"}

var deploymentsKind = schema.GroupVersionKind{Group: "apps", Version: "v1beta1", Kind: "Deployment"}

// Get takes name of the deployment, and returns the corresponding deployment object, and an error if there is any.
func (c *FakeDeployments) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Deployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(deploymentsResource, c.ns, name), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

// List takes label and field selectors, and returns the list of Deployments that match those selectors.
func (c *FakeDeployments) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.DeploymentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(deploymentsResource, deploymentsKind, c.ns, opts), &v1beta1.DeploymentList{})

	if obj == nil {
		return nil, err
	}
	label, field, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}

	_, flagSet, err := storage.DefaultNamespaceScopedAttr(obj)
	if err != nil {
		return nil, err
	}

	list := &v1beta1.DeploymentList{ListMeta: obj.(*v1beta1.DeploymentList).ListMeta}
	for _, item := range obj.(*v1beta1.DeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) || field.Matches(flagSet) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested deployments.
func (c *FakeDeployments) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(deploymentsResource, c.ns, opts))

}

// Create takes the representation of a deployment and creates it.  Returns the server's representation of the deployment, and an error, if there is any.
func (c *FakeDeployments) Create(ctx context.Context, deployment *v1beta1.Deployment, opts v1.CreateOptions) (result *v1beta1.Deployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(deploymentsResource, c.ns, deployment), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

// Update takes the representation of a deployment and updates it. Returns the server's representation of the deployment, and an error, if there is any.
func (c *FakeDeployments) Update(ctx context.Context, deployment *v1beta1.Deployment, opts v1.UpdateOptions) (result *v1beta1.Deployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(deploymentsResource, c.ns, deployment), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDeployments) UpdateStatus(ctx context.Context, deployment *v1beta1.Deployment, opts v1.UpdateOptions) (*v1beta1.Deployment, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(deploymentsResource, "status", c.ns, deployment), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

// Delete takes name of the deployment and deletes it. Returns an error if one occurs.
func (c *FakeDeployments) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(deploymentsResource, c.ns, name, opts), &v1beta1.Deployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDeployments) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(deploymentsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.DeploymentList{})
	return err
}

// Patch applies the patch and returns the patched deployment.
func (c *FakeDeployments) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Deployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(deploymentsResource, c.ns, name, pt, data, subresources...), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied deployment.
func (c *FakeDeployments) Apply(ctx context.Context, deployment *appsv1beta1.DeploymentApplyConfiguration, opts v1.ApplyOptions) (result *v1beta1.Deployment, err error) {
	if deployment == nil {
		return nil, fmt.Errorf("deployment provided to Apply must not be nil")
	}
	data, err := json.Marshal(deployment)
	if err != nil {
		return nil, err
	}
	name := deployment.Name
	if name == nil {
		return nil, fmt.Errorf("deployment.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(deploymentsResource, c.ns, *name, types.ApplyPatchType, data), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeDeployments) ApplyStatus(ctx context.Context, deployment *appsv1beta1.DeploymentApplyConfiguration, opts v1.ApplyOptions) (result *v1beta1.Deployment, err error) {
	if deployment == nil {
		return nil, fmt.Errorf("deployment provided to Apply must not be nil")
	}
	data, err := json.Marshal(deployment)
	if err != nil {
		return nil, err
	}
	name := deployment.Name
	if name == nil {
		return nil, fmt.Errorf("deployment.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(deploymentsResource, c.ns, *name, types.ApplyPatchType, data, "status"), &v1beta1.Deployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Deployment), err
}
