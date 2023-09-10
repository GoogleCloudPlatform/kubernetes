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

// Code generated by defaulter-gen. DO NOT EDIT.

package v1beta4

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// RegisterDefaults adds defaulters functions to the given scheme.
// Public to allow building arbitrary schemes.
// All generated defaulters are covering - they call all nested defaulters.
func RegisterDefaults(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&ClusterConfiguration{}, func(obj interface{}) { SetObjectDefaults_ClusterConfiguration(obj.(*ClusterConfiguration)) })
	scheme.AddTypeDefaultingFunc(&InitConfiguration{}, func(obj interface{}) { SetObjectDefaults_InitConfiguration(obj.(*InitConfiguration)) })
	scheme.AddTypeDefaultingFunc(&JoinConfiguration{}, func(obj interface{}) { SetObjectDefaults_JoinConfiguration(obj.(*JoinConfiguration)) })
	scheme.AddTypeDefaultingFunc(&ResetConfiguration{}, func(obj interface{}) { SetObjectDefaults_ResetConfiguration(obj.(*ResetConfiguration)) })
	return nil
}

func SetObjectDefaults_ClusterConfiguration(in *ClusterConfiguration) {
	SetDefaults_ClusterConfiguration(in)
	SetDefaults_APIServer(&in.APIServer)
}

func SetObjectDefaults_InitConfiguration(in *InitConfiguration) {
	SetDefaults_InitConfiguration(in)
	SetDefaults_APIEndpoint(&in.LocalAPIEndpoint)
}

func SetObjectDefaults_JoinConfiguration(in *JoinConfiguration) {
	SetDefaults_JoinConfiguration(in)
	SetDefaults_Discovery(&in.Discovery)
	if in.Discovery.File != nil {
		SetDefaults_FileDiscovery(in.Discovery.File)
	}
	if in.ControlPlane != nil {
		SetDefaults_JoinControlPlane(in.ControlPlane)
		SetDefaults_APIEndpoint(&in.ControlPlane.LocalAPIEndpoint)
	}
}

func SetObjectDefaults_ResetConfiguration(in *ResetConfiguration) {
	SetDefaults_ResetConfiguration(in)
}
