/*
Copyright 2024 The Kubernetes Authors.

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

package apiclient

import (
	"fmt"
	"reflect"
	"testing"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

// TestGetControlPlaneComponents tests the getControlPlaneComponents function
func TestGetControlPlaneComponents(t *testing.T) {
	// Define test cases with different ClusterConfiguration inputs and expected outputs
	testcases := []struct {
		name     string
		cfg      *kubeadmapi.ClusterConfiguration
		expected []controlPlaneComponent
	}{
		{
			// Test case: Custom port and address settings from config
			name: "port and addresses from config",
			cfg: &kubeadmapi.ClusterConfiguration{
				APIServer: kubeadmapi.APIServer{
					ControlPlaneComponent: kubeadmapi.ControlPlaneComponent{
						ExtraArgs: []kubeadmapi.Arg{
							{Name: "secure-port", Value: "1111"},
							{Name: "bind-address", Value: "0.0.0.0"},
						},
					},
				},
				ControllerManager: kubeadmapi.ControlPlaneComponent{
					ExtraArgs: []kubeadmapi.Arg{
						{Name: "secure-port", Value: "2222"},
						{Name: "bind-address", Value: "0.0.0.0"},
					},
				},
				Scheduler: kubeadmapi.ControlPlaneComponent{
					ExtraArgs: []kubeadmapi.Arg{
						{Name: "secure-port", Value: "3333"},
						{Name: "bind-address", Value: "0.0.0.0"},
					},
				},
			},
			expected: []controlPlaneComponent{
				{name: "kube-apiserver", url: fmt.Sprintf("https://0.0.0.0:1111/%s", endpointLivez)},
				{name: "kube-controller-manager", url: fmt.Sprintf("https://0.0.0.0:2222/%s", endpointHealthz)},
				{name: "kube-scheduler", url: fmt.Sprintf("https://0.0.0.0:3333/%s", endpointLivez)},
			},
		},
		{
			// Test case: Default port and address values
			name: "default ports and addresses",
			cfg:  &kubeadmapi.ClusterConfiguration{},
			expected: []controlPlaneComponent{
				{name: "kube-apiserver", url: fmt.Sprintf("https://127.0.0.1:6443/%s", endpointLivez)},
				{name: "kube-controller-manager", url: fmt.Sprintf("https://127.0.0.1:10257/%s", endpointHealthz)},
				{name: "kube-scheduler", url: fmt.Sprintf("https://127.0.0.1:10259/%s", endpointLivez)},
			},
		},
	}

	// Iterate over each test case and run the test
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function being tested
			actual := getControlPlaneComponents(tc.cfg)
			// Compare the actual result with the expected result
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Fatalf("expected result: %+v, got: %+v", tc.expected, actual)
			}
		})
	}
}
