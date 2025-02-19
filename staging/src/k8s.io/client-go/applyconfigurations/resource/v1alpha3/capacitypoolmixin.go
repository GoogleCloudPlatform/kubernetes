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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha3

import (
	resourcev1alpha3 "k8s.io/api/resource/v1alpha3"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

// CapacityPoolMixinApplyConfiguration represents a declarative configuration of the CapacityPoolMixin type for use
// with apply.
type CapacityPoolMixinApplyConfiguration struct {
	Name     *string                                              `json:"name,omitempty"`
	Capacity map[resourcev1alpha3.QualifiedName]resource.Quantity `json:"capacity,omitempty"`
}

// CapacityPoolMixinApplyConfiguration constructs a declarative configuration of the CapacityPoolMixin type for use with
// apply.
func CapacityPoolMixin() *CapacityPoolMixinApplyConfiguration {
	return &CapacityPoolMixinApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *CapacityPoolMixinApplyConfiguration) WithName(value string) *CapacityPoolMixinApplyConfiguration {
	b.Name = &value
	return b
}

// WithCapacity puts the entries into the Capacity field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Capacity field,
// overwriting an existing map entries in Capacity field with the same key.
func (b *CapacityPoolMixinApplyConfiguration) WithCapacity(entries map[resourcev1alpha3.QualifiedName]resource.Quantity) *CapacityPoolMixinApplyConfiguration {
	if b.Capacity == nil && len(entries) > 0 {
		b.Capacity = make(map[resourcev1alpha3.QualifiedName]resource.Quantity, len(entries))
	}
	for k, v := range entries {
		b.Capacity[k] = v
	}
	return b
}
