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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BasicDeviceApplyConfiguration represents a declarative configuration of the BasicDevice type for use
// with apply.
type BasicDeviceApplyConfiguration struct {
	Attributes               map[resourcev1alpha3.QualifiedName]DeviceAttributeApplyConfiguration `json:"attributes,omitempty"`
	Capacity                 map[resourcev1alpha3.QualifiedName]resource.Quantity                 `json:"capacity,omitempty"`
	UsageRestrictedToNode    *bool                                                                `json:"usageRestrictedToNode,omitempty"`
	BindingConditions        []string                                                             `json:"bindingConditions,omitempty"`
	BindingFailureConditions []string                                                             `json:"bindingFailureConditions,omitempty"`
	BindingTimeout           *v1.Duration                                                         `json:"bindingTimeout,omitempty"`
}

// BasicDeviceApplyConfiguration constructs a declarative configuration of the BasicDevice type for use with
// apply.
func BasicDevice() *BasicDeviceApplyConfiguration {
	return &BasicDeviceApplyConfiguration{}
}

// WithAttributes puts the entries into the Attributes field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Attributes field,
// overwriting an existing map entries in Attributes field with the same key.
func (b *BasicDeviceApplyConfiguration) WithAttributes(entries map[resourcev1alpha3.QualifiedName]DeviceAttributeApplyConfiguration) *BasicDeviceApplyConfiguration {
	if b.Attributes == nil && len(entries) > 0 {
		b.Attributes = make(map[resourcev1alpha3.QualifiedName]DeviceAttributeApplyConfiguration, len(entries))
	}
	for k, v := range entries {
		b.Attributes[k] = v
	}
	return b
}

// WithCapacity puts the entries into the Capacity field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Capacity field,
// overwriting an existing map entries in Capacity field with the same key.
func (b *BasicDeviceApplyConfiguration) WithCapacity(entries map[resourcev1alpha3.QualifiedName]resource.Quantity) *BasicDeviceApplyConfiguration {
	if b.Capacity == nil && len(entries) > 0 {
		b.Capacity = make(map[resourcev1alpha3.QualifiedName]resource.Quantity, len(entries))
	}
	for k, v := range entries {
		b.Capacity[k] = v
	}
	return b
}

// WithUsageRestrictedToNode sets the UsageRestrictedToNode field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the UsageRestrictedToNode field is set to the value of the last call.
func (b *BasicDeviceApplyConfiguration) WithUsageRestrictedToNode(value bool) *BasicDeviceApplyConfiguration {
	b.UsageRestrictedToNode = &value
	return b
}

// WithBindingConditions adds the given value to the BindingConditions field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the BindingConditions field.
func (b *BasicDeviceApplyConfiguration) WithBindingConditions(values ...string) *BasicDeviceApplyConfiguration {
	for i := range values {
		b.BindingConditions = append(b.BindingConditions, values[i])
	}
	return b
}

// WithBindingFailureConditions adds the given value to the BindingFailureConditions field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the BindingFailureConditions field.
func (b *BasicDeviceApplyConfiguration) WithBindingFailureConditions(values ...string) *BasicDeviceApplyConfiguration {
	for i := range values {
		b.BindingFailureConditions = append(b.BindingFailureConditions, values[i])
	}
	return b
}

// WithBindingTimeout sets the BindingTimeout field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BindingTimeout field is set to the value of the last call.
func (b *BasicDeviceApplyConfiguration) WithBindingTimeout(value v1.Duration) *BasicDeviceApplyConfiguration {
	b.BindingTimeout = &value
	return b
}
