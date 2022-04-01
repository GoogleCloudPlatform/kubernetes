/*
Copyright 2020 The Kubernetes Authors.

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

package e2enode

import (
	e2econfig "k8s.io/kubernetes/test/e2e/framework/config"
	e2eskipper "k8s.io/kubernetes/test/e2e/framework/skipper"
	e2eutils "k8s.io/kubernetes/test/e2e/framework/utils"
)

const (
	// SRIOVDevicePluginCMYAML is the path of the config map to configure the sriov device plugin.
	SRIOVDevicePluginCMYAML = "test/e2e_node/testing-manifests/sriovdp-cm.yaml"
	// SRIOVDevicePluginDSYAML is the path of the daemonset template of the sriov device plugin. // TODO: Parametrize it by making it a feature in TestFramework.
	SRIOVDevicePluginDSYAML = "test/e2e_node/testing-manifests/sriovdp-ds.yaml"
	// SRIOVDevicePluginSAYAML is the path of the service account needed by the sriov device plugin to run.
	SRIOVDevicePluginSAYAML = "test/e2e_node/testing-manifests/sriovdp-sa.yaml"
	// SRIOVDevicePluginName is the name of the device plugin pod
	SRIOVDevicePluginName = "sriov-device-plugin"
)

func requireSRIOVDevices() {
	sriovdevCount, err := countSRIOVDevices()
	e2eutils.ExpectNoError(err)

	if sriovdevCount > 0 {
		return // all good
	}

	msg := "this test is meant to run on a system with at least one configured VF from SRIOV device"
	if e2econfig.TestContext.RequireDevices {
		e2eutils.Failf(msg)
	} else {
		e2eskipper.Skipf(msg)
	}
}
