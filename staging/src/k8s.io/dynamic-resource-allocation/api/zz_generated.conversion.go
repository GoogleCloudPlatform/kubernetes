//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by conversion-gen. DO NOT EDIT.

package api

import (
	unsafe "unsafe"

	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/resource/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*BasicDevice)(nil), (*v1beta1.BasicDevice)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_BasicDevice_To_v1beta1_BasicDevice(a.(*BasicDevice), b.(*v1beta1.BasicDevice), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1beta1.BasicDevice)(nil), (*BasicDevice)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_BasicDevice_To_api_BasicDevice(a.(*v1beta1.BasicDevice), b.(*BasicDevice), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1beta1.Device)(nil), (*Device)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_Device_To_api_Device(a.(*v1beta1.Device), b.(*Device), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*DeviceAttribute)(nil), (*v1beta1.DeviceAttribute)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_DeviceAttribute_To_v1beta1_DeviceAttribute(a.(*DeviceAttribute), b.(*v1beta1.DeviceAttribute), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1beta1.DeviceAttribute)(nil), (*DeviceAttribute)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_DeviceAttribute_To_api_DeviceAttribute(a.(*v1beta1.DeviceAttribute), b.(*DeviceAttribute), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*DeviceCapacity)(nil), (*v1beta1.DeviceCapacity)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_DeviceCapacity_To_v1beta1_DeviceCapacity(a.(*DeviceCapacity), b.(*v1beta1.DeviceCapacity), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1beta1.DeviceCapacity)(nil), (*DeviceCapacity)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_DeviceCapacity_To_api_DeviceCapacity(a.(*v1beta1.DeviceCapacity), b.(*DeviceCapacity), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ResourcePool)(nil), (*v1beta1.ResourcePool)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_ResourcePool_To_v1beta1_ResourcePool(a.(*ResourcePool), b.(*v1beta1.ResourcePool), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1beta1.ResourcePool)(nil), (*ResourcePool)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_ResourcePool_To_api_ResourcePool(a.(*v1beta1.ResourcePool), b.(*ResourcePool), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ResourceSlice)(nil), (*v1beta1.ResourceSlice)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_ResourceSlice_To_v1beta1_ResourceSlice(a.(*ResourceSlice), b.(*v1beta1.ResourceSlice), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1beta1.ResourceSlice)(nil), (*ResourceSlice)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_ResourceSlice_To_api_ResourceSlice(a.(*v1beta1.ResourceSlice), b.(*ResourceSlice), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ResourceSliceSpec)(nil), (*v1beta1.ResourceSliceSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_ResourceSliceSpec_To_v1beta1_ResourceSliceSpec(a.(*ResourceSliceSpec), b.(*v1beta1.ResourceSliceSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1beta1.ResourceSliceSpec)(nil), (*ResourceSliceSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_ResourceSliceSpec_To_api_ResourceSliceSpec(a.(*v1beta1.ResourceSliceSpec), b.(*ResourceSliceSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*Device)(nil), (*v1beta1.Device)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_Device_To_v1beta1_Device(a.(*Device), b.(*v1beta1.Device), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*UniqueString)(nil), (*string)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_api_UniqueString_To_string(a.(*UniqueString), b.(*string), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*string)(nil), (*UniqueString)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_string_To_api_UniqueString(a.(*string), b.(*UniqueString), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_api_BasicDevice_To_v1beta1_BasicDevice(in *BasicDevice, out *v1beta1.BasicDevice, s conversion.Scope) error {
	out.Attributes = *(*map[v1beta1.QualifiedName]v1beta1.DeviceAttribute)(unsafe.Pointer(&in.Attributes))
	out.Capacity = *(*map[v1beta1.QualifiedName]v1beta1.DeviceCapacity)(unsafe.Pointer(&in.Capacity))
	out.UsageRestrictedToNode = in.UsageRestrictedToNode
	out.BindingConditions = *(*[]string)(unsafe.Pointer(&in.BindingConditions))
	out.BindingFailureConditions = *(*[]string)(unsafe.Pointer(&in.BindingFailureConditions))
	out.BindingTimeout = (*v1.Duration)(unsafe.Pointer(in.BindingTimeout))
	return nil
}

// Convert_api_BasicDevice_To_v1beta1_BasicDevice is an autogenerated conversion function.
func Convert_api_BasicDevice_To_v1beta1_BasicDevice(in *BasicDevice, out *v1beta1.BasicDevice, s conversion.Scope) error {
	return autoConvert_api_BasicDevice_To_v1beta1_BasicDevice(in, out, s)
}

func autoConvert_v1beta1_BasicDevice_To_api_BasicDevice(in *v1beta1.BasicDevice, out *BasicDevice, s conversion.Scope) error {
	out.Attributes = *(*map[QualifiedName]DeviceAttribute)(unsafe.Pointer(&in.Attributes))
	out.Capacity = *(*map[QualifiedName]DeviceCapacity)(unsafe.Pointer(&in.Capacity))
	out.UsageRestrictedToNode = in.UsageRestrictedToNode
	out.BindingConditions = *(*[]string)(unsafe.Pointer(&in.BindingConditions))
	out.BindingFailureConditions = *(*[]string)(unsafe.Pointer(&in.BindingFailureConditions))
	out.BindingTimeout = (*v1.Duration)(unsafe.Pointer(in.BindingTimeout))
	return nil
}

// Convert_v1beta1_BasicDevice_To_api_BasicDevice is an autogenerated conversion function.
func Convert_v1beta1_BasicDevice_To_api_BasicDevice(in *v1beta1.BasicDevice, out *BasicDevice, s conversion.Scope) error {
	return autoConvert_v1beta1_BasicDevice_To_api_BasicDevice(in, out, s)
}

func autoConvert_api_Device_To_v1beta1_Device(in *Device, out *v1beta1.Device, s conversion.Scope) error {
	if err := Convert_api_UniqueString_To_string(&in.Name, &out.Name, s); err != nil {
		return err
	}
	out.Basic = (*v1beta1.BasicDevice)(unsafe.Pointer(in.Basic))
	return nil
}

func autoConvert_v1beta1_Device_To_api_Device(in *v1beta1.Device, out *Device, s conversion.Scope) error {
	if err := Convert_string_To_api_UniqueString(&in.Name, &out.Name, s); err != nil {
		return err
	}
	out.Basic = (*BasicDevice)(unsafe.Pointer(in.Basic))
	return nil
}

// Convert_v1beta1_Device_To_api_Device is an autogenerated conversion function.
func Convert_v1beta1_Device_To_api_Device(in *v1beta1.Device, out *Device, s conversion.Scope) error {
	return autoConvert_v1beta1_Device_To_api_Device(in, out, s)
}

func autoConvert_api_DeviceAttribute_To_v1beta1_DeviceAttribute(in *DeviceAttribute, out *v1beta1.DeviceAttribute, s conversion.Scope) error {
	out.IntValue = (*int64)(unsafe.Pointer(in.IntValue))
	out.BoolValue = (*bool)(unsafe.Pointer(in.BoolValue))
	out.StringValue = (*string)(unsafe.Pointer(in.StringValue))
	out.VersionValue = (*string)(unsafe.Pointer(in.VersionValue))
	return nil
}

// Convert_api_DeviceAttribute_To_v1beta1_DeviceAttribute is an autogenerated conversion function.
func Convert_api_DeviceAttribute_To_v1beta1_DeviceAttribute(in *DeviceAttribute, out *v1beta1.DeviceAttribute, s conversion.Scope) error {
	return autoConvert_api_DeviceAttribute_To_v1beta1_DeviceAttribute(in, out, s)
}

func autoConvert_v1beta1_DeviceAttribute_To_api_DeviceAttribute(in *v1beta1.DeviceAttribute, out *DeviceAttribute, s conversion.Scope) error {
	out.IntValue = (*int64)(unsafe.Pointer(in.IntValue))
	out.BoolValue = (*bool)(unsafe.Pointer(in.BoolValue))
	out.StringValue = (*string)(unsafe.Pointer(in.StringValue))
	out.VersionValue = (*string)(unsafe.Pointer(in.VersionValue))
	return nil
}

// Convert_v1beta1_DeviceAttribute_To_api_DeviceAttribute is an autogenerated conversion function.
func Convert_v1beta1_DeviceAttribute_To_api_DeviceAttribute(in *v1beta1.DeviceAttribute, out *DeviceAttribute, s conversion.Scope) error {
	return autoConvert_v1beta1_DeviceAttribute_To_api_DeviceAttribute(in, out, s)
}

func autoConvert_api_DeviceCapacity_To_v1beta1_DeviceCapacity(in *DeviceCapacity, out *v1beta1.DeviceCapacity, s conversion.Scope) error {
	out.Value = in.Value
	return nil
}

// Convert_api_DeviceCapacity_To_v1beta1_DeviceCapacity is an autogenerated conversion function.
func Convert_api_DeviceCapacity_To_v1beta1_DeviceCapacity(in *DeviceCapacity, out *v1beta1.DeviceCapacity, s conversion.Scope) error {
	return autoConvert_api_DeviceCapacity_To_v1beta1_DeviceCapacity(in, out, s)
}

func autoConvert_v1beta1_DeviceCapacity_To_api_DeviceCapacity(in *v1beta1.DeviceCapacity, out *DeviceCapacity, s conversion.Scope) error {
	out.Value = in.Value
	return nil
}

// Convert_v1beta1_DeviceCapacity_To_api_DeviceCapacity is an autogenerated conversion function.
func Convert_v1beta1_DeviceCapacity_To_api_DeviceCapacity(in *v1beta1.DeviceCapacity, out *DeviceCapacity, s conversion.Scope) error {
	return autoConvert_v1beta1_DeviceCapacity_To_api_DeviceCapacity(in, out, s)
}

func autoConvert_api_ResourcePool_To_v1beta1_ResourcePool(in *ResourcePool, out *v1beta1.ResourcePool, s conversion.Scope) error {
	if err := Convert_api_UniqueString_To_string(&in.Name, &out.Name, s); err != nil {
		return err
	}
	out.Generation = in.Generation
	out.ResourceSliceCount = in.ResourceSliceCount
	return nil
}

// Convert_api_ResourcePool_To_v1beta1_ResourcePool is an autogenerated conversion function.
func Convert_api_ResourcePool_To_v1beta1_ResourcePool(in *ResourcePool, out *v1beta1.ResourcePool, s conversion.Scope) error {
	return autoConvert_api_ResourcePool_To_v1beta1_ResourcePool(in, out, s)
}

func autoConvert_v1beta1_ResourcePool_To_api_ResourcePool(in *v1beta1.ResourcePool, out *ResourcePool, s conversion.Scope) error {
	if err := Convert_string_To_api_UniqueString(&in.Name, &out.Name, s); err != nil {
		return err
	}
	out.Generation = in.Generation
	out.ResourceSliceCount = in.ResourceSliceCount
	return nil
}

// Convert_v1beta1_ResourcePool_To_api_ResourcePool is an autogenerated conversion function.
func Convert_v1beta1_ResourcePool_To_api_ResourcePool(in *v1beta1.ResourcePool, out *ResourcePool, s conversion.Scope) error {
	return autoConvert_v1beta1_ResourcePool_To_api_ResourcePool(in, out, s)
}

func autoConvert_api_ResourceSlice_To_v1beta1_ResourceSlice(in *ResourceSlice, out *v1beta1.ResourceSlice, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_api_ResourceSliceSpec_To_v1beta1_ResourceSliceSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	return nil
}

// Convert_api_ResourceSlice_To_v1beta1_ResourceSlice is an autogenerated conversion function.
func Convert_api_ResourceSlice_To_v1beta1_ResourceSlice(in *ResourceSlice, out *v1beta1.ResourceSlice, s conversion.Scope) error {
	return autoConvert_api_ResourceSlice_To_v1beta1_ResourceSlice(in, out, s)
}

func autoConvert_v1beta1_ResourceSlice_To_api_ResourceSlice(in *v1beta1.ResourceSlice, out *ResourceSlice, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1beta1_ResourceSliceSpec_To_api_ResourceSliceSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1beta1_ResourceSlice_To_api_ResourceSlice is an autogenerated conversion function.
func Convert_v1beta1_ResourceSlice_To_api_ResourceSlice(in *v1beta1.ResourceSlice, out *ResourceSlice, s conversion.Scope) error {
	return autoConvert_v1beta1_ResourceSlice_To_api_ResourceSlice(in, out, s)
}

func autoConvert_api_ResourceSliceSpec_To_v1beta1_ResourceSliceSpec(in *ResourceSliceSpec, out *v1beta1.ResourceSliceSpec, s conversion.Scope) error {
	if err := Convert_api_UniqueString_To_string(&in.Driver, &out.Driver, s); err != nil {
		return err
	}
	if err := Convert_api_ResourcePool_To_v1beta1_ResourcePool(&in.Pool, &out.Pool, s); err != nil {
		return err
	}
	if err := Convert_api_UniqueString_To_string(&in.NodeName, &out.NodeName, s); err != nil {
		return err
	}
	out.NodeSelector = (*corev1.NodeSelector)(unsafe.Pointer(in.NodeSelector))
	out.AllNodes = in.AllNodes
	if in.Devices != nil {
		in, out := &in.Devices, &out.Devices
		*out = make([]v1beta1.Device, len(*in))
		for i := range *in {
			if err := Convert_api_Device_To_v1beta1_Device(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Devices = nil
	}
	return nil
}

// Convert_api_ResourceSliceSpec_To_v1beta1_ResourceSliceSpec is an autogenerated conversion function.
func Convert_api_ResourceSliceSpec_To_v1beta1_ResourceSliceSpec(in *ResourceSliceSpec, out *v1beta1.ResourceSliceSpec, s conversion.Scope) error {
	return autoConvert_api_ResourceSliceSpec_To_v1beta1_ResourceSliceSpec(in, out, s)
}

func autoConvert_v1beta1_ResourceSliceSpec_To_api_ResourceSliceSpec(in *v1beta1.ResourceSliceSpec, out *ResourceSliceSpec, s conversion.Scope) error {
	if err := Convert_string_To_api_UniqueString(&in.Driver, &out.Driver, s); err != nil {
		return err
	}
	if err := Convert_v1beta1_ResourcePool_To_api_ResourcePool(&in.Pool, &out.Pool, s); err != nil {
		return err
	}
	if err := Convert_string_To_api_UniqueString(&in.NodeName, &out.NodeName, s); err != nil {
		return err
	}
	out.NodeSelector = (*corev1.NodeSelector)(unsafe.Pointer(in.NodeSelector))
	out.AllNodes = in.AllNodes
	if in.Devices != nil {
		in, out := &in.Devices, &out.Devices
		*out = make([]Device, len(*in))
		for i := range *in {
			if err := Convert_v1beta1_Device_To_api_Device(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Devices = nil
	}
	return nil
}

// Convert_v1beta1_ResourceSliceSpec_To_api_ResourceSliceSpec is an autogenerated conversion function.
func Convert_v1beta1_ResourceSliceSpec_To_api_ResourceSliceSpec(in *v1beta1.ResourceSliceSpec, out *ResourceSliceSpec, s conversion.Scope) error {
	return autoConvert_v1beta1_ResourceSliceSpec_To_api_ResourceSliceSpec(in, out, s)
}
