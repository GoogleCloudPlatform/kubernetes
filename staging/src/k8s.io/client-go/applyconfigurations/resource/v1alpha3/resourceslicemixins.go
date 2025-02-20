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

// ResourceSliceMixinsApplyConfiguration represents a declarative configuration of the ResourceSliceMixins type for use
// with apply.
type ResourceSliceMixinsApplyConfiguration struct {
	Device                    []DeviceMixinApplyConfiguration                    `json:"device,omitempty"`
	DeviceCapacityConsumption []DeviceCapacityConsumptionMixinApplyConfiguration `json:"deviceCapacityConsumption,omitempty"`
	CapacityPool              []CapacityPoolMixinApplyConfiguration              `json:"capacityPool,omitempty"`
}

// ResourceSliceMixinsApplyConfiguration constructs a declarative configuration of the ResourceSliceMixins type for use with
// apply.
func ResourceSliceMixins() *ResourceSliceMixinsApplyConfiguration {
	return &ResourceSliceMixinsApplyConfiguration{}
}

// WithDevice adds the given value to the Device field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Device field.
func (b *ResourceSliceMixinsApplyConfiguration) WithDevice(values ...*DeviceMixinApplyConfiguration) *ResourceSliceMixinsApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithDevice")
		}
		b.Device = append(b.Device, *values[i])
	}
	return b
}

// WithDeviceCapacityConsumption adds the given value to the DeviceCapacityConsumption field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the DeviceCapacityConsumption field.
func (b *ResourceSliceMixinsApplyConfiguration) WithDeviceCapacityConsumption(values ...*DeviceCapacityConsumptionMixinApplyConfiguration) *ResourceSliceMixinsApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithDeviceCapacityConsumption")
		}
		b.DeviceCapacityConsumption = append(b.DeviceCapacityConsumption, *values[i])
	}
	return b
}

// WithCapacityPool adds the given value to the CapacityPool field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the CapacityPool field.
func (b *ResourceSliceMixinsApplyConfiguration) WithCapacityPool(values ...*CapacityPoolMixinApplyConfiguration) *ResourceSliceMixinsApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithCapacityPool")
		}
		b.CapacityPool = append(b.CapacityPool, *values[i])
	}
	return b
}
