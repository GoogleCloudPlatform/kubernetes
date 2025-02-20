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

package v1beta2

import (
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
)

// CapacityPoolApplyConfiguration represents a declarative configuration of the CapacityPool type for use
// with apply.
type CapacityPoolApplyConfiguration struct {
	Name     *string                                                            `json:"name,omitempty"`
	Includes []CapacityPoolMixinRefApplyConfiguration                           `json:"includes,omitempty"`
	Capacity map[resourcev1beta2.QualifiedName]DeviceCapacityApplyConfiguration `json:"capacity,omitempty"`
}

// CapacityPoolApplyConfiguration constructs a declarative configuration of the CapacityPool type for use with
// apply.
func CapacityPool() *CapacityPoolApplyConfiguration {
	return &CapacityPoolApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *CapacityPoolApplyConfiguration) WithName(value string) *CapacityPoolApplyConfiguration {
	b.Name = &value
	return b
}

// WithIncludes adds the given value to the Includes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Includes field.
func (b *CapacityPoolApplyConfiguration) WithIncludes(values ...*CapacityPoolMixinRefApplyConfiguration) *CapacityPoolApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithIncludes")
		}
		b.Includes = append(b.Includes, *values[i])
	}
	return b
}

// WithCapacity puts the entries into the Capacity field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Capacity field,
// overwriting an existing map entries in Capacity field with the same key.
func (b *CapacityPoolApplyConfiguration) WithCapacity(entries map[resourcev1beta2.QualifiedName]DeviceCapacityApplyConfiguration) *CapacityPoolApplyConfiguration {
	if b.Capacity == nil && len(entries) > 0 {
		b.Capacity = make(map[resourcev1beta2.QualifiedName]DeviceCapacityApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.Capacity[k] = v
	}
	return b
}
