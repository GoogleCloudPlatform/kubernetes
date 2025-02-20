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

// NetworkDeviceDataApplyConfiguration represents a declarative configuration of the NetworkDeviceData type for use
// with apply.
type NetworkDeviceDataApplyConfiguration struct {
	InterfaceName   *string  `json:"interfaceName,omitempty"`
	IPs             []string `json:"ips,omitempty"`
	HardwareAddress *string  `json:"hardwareAddress,omitempty"`
}

// NetworkDeviceDataApplyConfiguration constructs a declarative configuration of the NetworkDeviceData type for use with
// apply.
func NetworkDeviceData() *NetworkDeviceDataApplyConfiguration {
	return &NetworkDeviceDataApplyConfiguration{}
}

// WithInterfaceName sets the InterfaceName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the InterfaceName field is set to the value of the last call.
func (b *NetworkDeviceDataApplyConfiguration) WithInterfaceName(value string) *NetworkDeviceDataApplyConfiguration {
	b.InterfaceName = &value
	return b
}

// WithIPs adds the given value to the IPs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the IPs field.
func (b *NetworkDeviceDataApplyConfiguration) WithIPs(values ...string) *NetworkDeviceDataApplyConfiguration {
	for i := range values {
		b.IPs = append(b.IPs, values[i])
	}
	return b
}

// WithHardwareAddress sets the HardwareAddress field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HardwareAddress field is set to the value of the last call.
func (b *NetworkDeviceDataApplyConfiguration) WithHardwareAddress(value string) *NetworkDeviceDataApplyConfiguration {
	b.HardwareAddress = &value
	return b
}
