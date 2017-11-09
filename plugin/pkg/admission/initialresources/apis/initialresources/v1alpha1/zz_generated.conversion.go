// +build !ignore_autogenerated

/*
Copyright 2017 The Kubernetes Authors.

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

// This file was autogenerated by conversion-gen. Do not edit it manually!

package v1alpha1

import (
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	initialresources "k8s.io/kubernetes/plugin/pkg/admission/initialresources/apis/initialresources"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedConversionFuncs(
		Convert_v1alpha1_Configuration_To_initialresources_Configuration,
		Convert_initialresources_Configuration_To_v1alpha1_Configuration,
		Convert_v1alpha1_DataSourceInfo_To_initialresources_DataSourceInfo,
		Convert_initialresources_DataSourceInfo_To_v1alpha1_DataSourceInfo,
	)
}

func autoConvert_v1alpha1_Configuration_To_initialresources_Configuration(in *Configuration, out *initialresources.Configuration, s conversion.Scope) error {
	if err := Convert_v1alpha1_DataSourceInfo_To_initialresources_DataSourceInfo(&in.DataSourceInfo, &out.DataSourceInfo, s); err != nil {
		return err
	}
	out.Percentile = in.Percentile
	out.NamespaceOnly = in.NamespaceOnly
	return nil
}

// Convert_v1alpha1_Configuration_To_initialresources_Configuration is an autogenerated conversion function.
func Convert_v1alpha1_Configuration_To_initialresources_Configuration(in *Configuration, out *initialresources.Configuration, s conversion.Scope) error {
	return autoConvert_v1alpha1_Configuration_To_initialresources_Configuration(in, out, s)
}

func autoConvert_initialresources_Configuration_To_v1alpha1_Configuration(in *initialresources.Configuration, out *Configuration, s conversion.Scope) error {
	if err := Convert_initialresources_DataSourceInfo_To_v1alpha1_DataSourceInfo(&in.DataSourceInfo, &out.DataSourceInfo, s); err != nil {
		return err
	}
	out.Percentile = in.Percentile
	out.NamespaceOnly = in.NamespaceOnly
	return nil
}

// Convert_initialresources_Configuration_To_v1alpha1_Configuration is an autogenerated conversion function.
func Convert_initialresources_Configuration_To_v1alpha1_Configuration(in *initialresources.Configuration, out *Configuration, s conversion.Scope) error {
	return autoConvert_initialresources_Configuration_To_v1alpha1_Configuration(in, out, s)
}

func autoConvert_v1alpha1_DataSourceInfo_To_initialresources_DataSourceInfo(in *DataSourceInfo, out *initialresources.DataSourceInfo, s conversion.Scope) error {
	out.DataSource = initialresources.DataSourceType(in.DataSource)
	out.InfluxdbHost = in.InfluxdbHost
	out.InfluxdbUser = in.InfluxdbUser
	out.InfluxdbPassword = in.InfluxdbPassword
	out.InfluxdbName = in.InfluxdbName
	out.HawkularUrl = in.HawkularUrl
	return nil
}

// Convert_v1alpha1_DataSourceInfo_To_initialresources_DataSourceInfo is an autogenerated conversion function.
func Convert_v1alpha1_DataSourceInfo_To_initialresources_DataSourceInfo(in *DataSourceInfo, out *initialresources.DataSourceInfo, s conversion.Scope) error {
	return autoConvert_v1alpha1_DataSourceInfo_To_initialresources_DataSourceInfo(in, out, s)
}

func autoConvert_initialresources_DataSourceInfo_To_v1alpha1_DataSourceInfo(in *initialresources.DataSourceInfo, out *DataSourceInfo, s conversion.Scope) error {
	out.DataSource = DataSourceType(in.DataSource)
	out.InfluxdbHost = in.InfluxdbHost
	out.InfluxdbUser = in.InfluxdbUser
	out.InfluxdbPassword = in.InfluxdbPassword
	out.InfluxdbName = in.InfluxdbName
	out.HawkularUrl = in.HawkularUrl
	return nil
}

// Convert_initialresources_DataSourceInfo_To_v1alpha1_DataSourceInfo is an autogenerated conversion function.
func Convert_initialresources_DataSourceInfo_To_v1alpha1_DataSourceInfo(in *initialresources.DataSourceInfo, out *DataSourceInfo, s conversion.Scope) error {
	return autoConvert_initialresources_DataSourceInfo_To_v1alpha1_DataSourceInfo(in, out, s)
}
