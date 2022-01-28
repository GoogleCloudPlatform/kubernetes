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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/storage"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	testing "k8s.io/client-go/testing"
)

// FakePods implements PodInterface
type FakePods struct {
	Fake *FakeCoreV1
	ns   string
}

var podsResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

var podsKind = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}

// Get takes name of the pod, and returns the corresponding pod object, and an error if there is any.
func (c *FakePods) Get(ctx context.Context, name string, options v1.GetOptions) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podsResource, c.ns, name), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// List takes label and field selectors, and returns the list of Pods that match those selectors.
func (c *FakePods) List(ctx context.Context, opts v1.ListOptions) (result *corev1.PodList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podsResource, podsKind, c.ns, opts), &corev1.PodList{})

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

	list := &corev1.PodList{ListMeta: obj.(*corev1.PodList).ListMeta}
	for _, item := range obj.(*corev1.PodList).Items {
		if label.Matches(labels.Set(item.Labels)) || field.Matches(flagSet) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pods.
func (c *FakePods) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podsResource, c.ns, opts))

}

// Create takes the representation of a pod and creates it.  Returns the server's representation of the pod, and an error, if there is any.
func (c *FakePods) Create(ctx context.Context, pod *corev1.Pod, opts v1.CreateOptions) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podsResource, c.ns, pod), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// Update takes the representation of a pod and updates it. Returns the server's representation of the pod, and an error, if there is any.
func (c *FakePods) Update(ctx context.Context, pod *corev1.Pod, opts v1.UpdateOptions) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podsResource, c.ns, pod), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePods) UpdateStatus(ctx context.Context, pod *corev1.Pod, opts v1.UpdateOptions) (*corev1.Pod, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podsResource, "status", c.ns, pod), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// Delete takes name of the pod and deletes it. Returns an error if one occurs.
func (c *FakePods) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(podsResource, c.ns, name, opts), &corev1.Pod{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePods) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &corev1.PodList{})
	return err
}

// Patch applies the patch and returns the patched pod.
func (c *FakePods) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podsResource, c.ns, name, pt, data, subresources...), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied pod.
func (c *FakePods) Apply(ctx context.Context, pod *applyconfigurationscorev1.PodApplyConfiguration, opts v1.ApplyOptions) (result *corev1.Pod, err error) {
	if pod == nil {
		return nil, fmt.Errorf("pod provided to Apply must not be nil")
	}
	data, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	name := pod.Name
	if name == nil {
		return nil, fmt.Errorf("pod.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podsResource, c.ns, *name, types.ApplyPatchType, data), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakePods) ApplyStatus(ctx context.Context, pod *applyconfigurationscorev1.PodApplyConfiguration, opts v1.ApplyOptions) (result *corev1.Pod, err error) {
	if pod == nil {
		return nil, fmt.Errorf("pod provided to Apply must not be nil")
	}
	data, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	name := pod.Name
	if name == nil {
		return nil, fmt.Errorf("pod.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podsResource, c.ns, *name, types.ApplyPatchType, data, "status"), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// UpdateEphemeralContainers takes the representation of a pod and updates it. Returns the server's representation of the pod, and an error, if there is any.
func (c *FakePods) UpdateEphemeralContainers(ctx context.Context, podName string, pod *corev1.Pod, opts v1.UpdateOptions) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podsResource, "ephemeralcontainers", c.ns, pod), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}
