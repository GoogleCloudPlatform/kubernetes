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

// DeviceRequestApplyConfiguration represents a declarative configuration of the DeviceRequest type for use
// with apply.
type DeviceRequestApplyConfiguration struct {
	Name           *string                                  `json:"name,omitempty"`
	Exactly        *SpecificDeviceRequestApplyConfiguration `json:"exactly,omitempty"`
	FirstAvailable []DeviceSubRequestApplyConfiguration     `json:"firstAvailable,omitempty"`
}

// DeviceRequestApplyConfiguration constructs a declarative configuration of the DeviceRequest type for use with
// apply.
func DeviceRequest() *DeviceRequestApplyConfiguration {
	return &DeviceRequestApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *DeviceRequestApplyConfiguration) WithName(value string) *DeviceRequestApplyConfiguration {
	b.Name = &value
	return b
}

// WithExactly sets the Exactly field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Exactly field is set to the value of the last call.
func (b *DeviceRequestApplyConfiguration) WithExactly(value *SpecificDeviceRequestApplyConfiguration) *DeviceRequestApplyConfiguration {
	b.Exactly = value
	return b
}

// WithFirstAvailable adds the given value to the FirstAvailable field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the FirstAvailable field.
func (b *DeviceRequestApplyConfiguration) WithFirstAvailable(values ...*DeviceSubRequestApplyConfiguration) *DeviceRequestApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithFirstAvailable")
		}
		b.FirstAvailable = append(b.FirstAvailable, *values[i])
	}
	return b
}
