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

package deviceplugin

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
	"k8s.io/kubernetes/pkg/kubelet/lifecycle"
	utilstore "k8s.io/kubernetes/pkg/kubelet/util/store"
	utilfs "k8s.io/kubernetes/pkg/util/filesystem"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

const (
	socketName       = "/tmp/device_plugin/server.sock"
	pluginSocketName = "/tmp/device_plugin/device-plugin.sock"
	testResourceName = "fake-domain/resource"
)

func TestNewManagerImpl(t *testing.T) {
	_, err := newManagerImpl(socketName)
	require.NoError(t, err)
}

func TestNewManagerImplStart(t *testing.T) {
	m, p := setup(t, []*pluginapi.Device{}, func(n string, a, u, r []pluginapi.Device) {})
	cleanup(t, m, p)
}

// Tests that the device plugin manager correctly handles registration and re-registration by
// making sure that after registration, devices are correctly updated and if a re-registration
// happens, we will NOT delete devices; and no orphaned devices left.
func TestDevicePluginReRegistration(t *testing.T) {
	devs := []*pluginapi.Device{
		{ID: "Dev1", Health: pluginapi.Healthy},
		{ID: "Dev2", Health: pluginapi.Healthy},
	}
	devsForRegistration := []*pluginapi.Device{
		{ID: "Dev3", Health: pluginapi.Healthy},
	}

	callbackCount := 0
	callbackChan := make(chan int)
	var stopping int32
	stopping = 0
	callback := func(n string, a, u, r []pluginapi.Device) {
		// Should be called three times, one for each plugin registration, till we are stopping.
		if callbackCount > 2 && atomic.LoadInt32(&stopping) <= 0 {
			t.FailNow()
		}
		callbackCount++
		callbackChan <- callbackCount
	}
	m, p1 := setup(t, devs, callback)
	p1.Register(socketName, testResourceName)
	// Wait for the first callback to be issued.

	<-callbackChan
	// Wait till the endpoint is added to the manager.
	for i := 0; i < 20; i++ {
		if len(m.Devices()) > 0 {
			break
		}
		time.Sleep(1)
	}
	devices := m.Devices()
	require.Equal(t, 2, len(devices[testResourceName]), "Devices are not updated.")

	p2 := NewDevicePluginStub(devs, pluginSocketName+".new")
	err := p2.Start()
	require.NoError(t, err)
	p2.Register(socketName, testResourceName)
	// Wait for the second callback to be issued.
	<-callbackChan

	devices2 := m.Devices()
	require.Equal(t, 2, len(devices2[testResourceName]), "Devices shouldn't change.")

	// Test the scenario that a plugin re-registers with different devices.
	p3 := NewDevicePluginStub(devsForRegistration, pluginSocketName+".third")
	err = p3.Start()
	require.NoError(t, err)
	p3.Register(socketName, testResourceName)
	// Wait for the second callback to be issued.
	<-callbackChan

	devices3 := m.Devices()
	require.Equal(t, 1, len(devices3[testResourceName]), "Devices of plugin previously registered should be removed.")
	// Wait long enough to catch unexpected callbacks.
	time.Sleep(5 * time.Second)

	atomic.StoreInt32(&stopping, 1)
	p2.Stop()
	p3.Stop()
	cleanup(t, m, p1)

}

func setup(t *testing.T, devs []*pluginapi.Device, callback monitorCallback) (Manager, *Stub) {
	m, err := newManagerImpl(socketName)
	require.NoError(t, err)

	m.callback = callback

	activePods := func() []*v1.Pod {
		return []*v1.Pod{}
	}
	err = m.Start(activePods, &sourcesReadyStub{})
	require.NoError(t, err)

	p := NewDevicePluginStub(devs, pluginSocketName)
	err = p.Start()
	require.NoError(t, err)

	return m, p
}

func cleanup(t *testing.T, m Manager, p *Stub) {
	p.Stop()
	m.Stop()
}

func TestUpdateCapacity(t *testing.T) {
	testManager, err := newManagerImpl(socketName)
	as := assert.New(t)
	as.NotNil(testManager)
	as.Nil(err)

	devs := []pluginapi.Device{
		{ID: "Device1", Health: pluginapi.Healthy},
		{ID: "Device2", Health: pluginapi.Healthy},
		{ID: "Device3", Health: pluginapi.Unhealthy},
	}
	callback := testManager.genericDeviceUpdateCallback

	// Adds three devices for resource1, two healthy and one unhealthy.
	// Expects capacity for resource1 to be 2.
	resourceName1 := "domain1.com/resource1"
	testManager.endpoints[resourceName1] = &endpointImpl{devices: make(map[string]pluginapi.Device)}
	callback(resourceName1, devs, []pluginapi.Device{}, []pluginapi.Device{})
	capacity, removedResources := testManager.GetCapacity()
	resource1Capacity, ok := capacity[v1.ResourceName(resourceName1)]
	as.True(ok)
	as.Equal(int64(2), resource1Capacity.Value())
	as.Equal(0, len(removedResources))

	// Deletes an unhealthy device should NOT change capacity.
	callback(resourceName1, []pluginapi.Device{}, []pluginapi.Device{}, []pluginapi.Device{devs[2]})
	capacity, removedResources = testManager.GetCapacity()
	resource1Capacity, ok = capacity[v1.ResourceName(resourceName1)]
	as.True(ok)
	as.Equal(int64(2), resource1Capacity.Value())
	as.Equal(0, len(removedResources))

	// Updates a healthy device to unhealthy should reduce capacity by 1.
	dev2 := devs[1]
	dev2.Health = pluginapi.Unhealthy
	callback(resourceName1, []pluginapi.Device{}, []pluginapi.Device{dev2}, []pluginapi.Device{})
	capacity, removedResources = testManager.GetCapacity()
	resource1Capacity, ok = capacity[v1.ResourceName(resourceName1)]
	as.True(ok)
	as.Equal(int64(1), resource1Capacity.Value())
	as.Equal(0, len(removedResources))

	// Deletes a healthy device should reduce capacity by 1.
	callback(resourceName1, []pluginapi.Device{}, []pluginapi.Device{}, []pluginapi.Device{devs[0]})
	capacity, removedResources = testManager.GetCapacity()
	resource1Capacity, ok = capacity[v1.ResourceName(resourceName1)]
	as.True(ok)
	as.Equal(int64(0), resource1Capacity.Value())
	as.Equal(0, len(removedResources))

	// Tests adding another resource.
	resourceName2 := "resource2"
	testManager.endpoints[resourceName2] = &endpointImpl{devices: make(map[string]pluginapi.Device)}
	callback(resourceName2, devs, []pluginapi.Device{}, []pluginapi.Device{})
	capacity, removedResources = testManager.GetCapacity()
	as.Equal(2, len(capacity))
	resource2Capacity, ok := capacity[v1.ResourceName(resourceName2)]
	as.True(ok)
	as.Equal(int64(2), resource2Capacity.Value())
	as.Equal(0, len(removedResources))

	// Removes resourceName1 endpoint. Verifies testManager.GetCapacity() reports that resourceName1
	// is removed from capacity and it no longer exists in allDevices after the call.
	delete(testManager.endpoints, resourceName1)
	capacity, removed := testManager.GetCapacity()
	as.Equal([]string{resourceName1}, removed)
	_, ok = capacity[v1.ResourceName(resourceName1)]
	as.False(ok)
	val, ok := capacity[v1.ResourceName(resourceName2)]
	as.True(ok)
	as.Equal(int64(2), val.Value())
	_, ok = testManager.allDevices[resourceName1]
	as.False(ok)
}

type stringPairType struct {
	value1 string
	value2 string
}

func constructDevices(devices []string) sets.String {
	ret := sets.NewString()
	for _, dev := range devices {
		ret.Insert(dev)
	}
	return ret
}

func constructAllocResp(devices, mounts, envs map[string]string) *pluginapi.AllocateResponse {
	resp := &pluginapi.AllocateResponse{}
	for k, v := range devices {
		resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
			HostPath:      k,
			ContainerPath: v,
			Permissions:   "mrw",
		})
	}
	for k, v := range mounts {
		resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
			ContainerPath: k,
			HostPath:      v,
			ReadOnly:      true,
		})
	}
	resp.Envs = make(map[string]string)
	for k, v := range envs {
		resp.Envs[k] = v
	}
	return resp
}

func TestCheckpoint(t *testing.T) {
	resourceName1 := "domain1.com/resource1"
	resourceName2 := "domain2.com/resource2"

	as := assert.New(t)
	tmpDir, err := ioutil.TempDir("", "checkpoint")
	as.Nil(err)
	defer os.RemoveAll(tmpDir)
	testManager := &ManagerImpl{
		socketdir:  tmpDir,
		allDevices: make(map[string]sets.String),
		allocation: newAllocation(),
	}
	testManager.store, _ = utilstore.NewFileStore("/tmp", utilfs.DefaultFs{})

	testManager.allocation.insert("pod1", "con1", resourceName1,
		constructDevices([]string{"dev1", "dev2"}),
		constructAllocResp(map[string]string{"/dev/r1dev1": "/dev/r1dev1", "/dev/r1dev2": "/dev/r1dev2"},
			map[string]string{"/home/r1lib1": "/usr/r1lib1"}, map[string]string{}))
	testManager.allocation.insert("pod1", "con1", resourceName2,
		constructDevices([]string{"dev1", "dev2"}),
		constructAllocResp(map[string]string{"/dev/r2dev1": "/dev/r2dev1", "/dev/r2dev2": "/dev/r2dev2"},
			map[string]string{"/home/r2lib1": "/usr/r2lib1"},
			map[string]string{"r2devices": "dev1 dev2"}))
	testManager.allocation.insert("pod1", "con2", resourceName1,
		constructDevices([]string{"dev3"}),
		constructAllocResp(map[string]string{"/dev/r1dev3": "/dev/r1dev3"},
			map[string]string{"/home/r1lib1": "/usr/r1lib1"}, map[string]string{}))
	testManager.allocation.insert("pod2", "con1", resourceName1,
		constructDevices([]string{"dev4"}),
		constructAllocResp(map[string]string{"/dev/r1dev4": "/dev/r1dev4"},
			map[string]string{"/home/r1lib1": "/usr/r1lib1"}, map[string]string{}))

	testManager.allDevices[resourceName1] = sets.NewString()
	testManager.allDevices[resourceName1].Insert("dev1")
	testManager.allDevices[resourceName1].Insert("dev2")
	testManager.allDevices[resourceName1].Insert("dev3")
	testManager.allDevices[resourceName1].Insert("dev4")
	testManager.allDevices[resourceName1].Insert("dev5")
	testManager.allDevices[resourceName2] = sets.NewString()
	testManager.allDevices[resourceName2].Insert("dev1")
	testManager.allDevices[resourceName2].Insert("dev2")

	allocation := testManager.allocation
	expectedAllDevices := testManager.allDevices

	err = testManager.writeCheckpoint()

	as.Nil(err)
	testManager.allocation = newAllocation()
	err = testManager.readCheckpoint()
	as.Nil(err)

	as.Equal(len(allocation.pods()), len(testManager.allocation.pods()))
	for p, cmap := range allocation.allocations() {
		for c, dmap := range cmap {
			for r := range dmap {
				as.True(reflect.DeepEqual(
					allocation.devicesByContainer(p, c, r),
					testManager.allocation.devicesByContainer(p, c, r)))
				opts1 := allocation.collectRunOptions(p, c)
				opts2 := testManager.allocation.collectRunOptions(p, c)
				as.Equal(len(opts1.Envs), len(opts2.Envs))
				as.Equal(len(opts1.Mounts), len(opts2.Mounts))
				as.Equal(len(opts1.Devices), len(opts2.Devices))
			}
		}
	}
	as.True(reflect.DeepEqual(expectedAllDevices, testManager.allDevices))
}

type activePodsStub struct {
	activePods []*v1.Pod
}

func (a *activePodsStub) getActivePods() []*v1.Pod {
	return a.activePods
}

func (a *activePodsStub) updateActivePods(newPods []*v1.Pod) {
	a.activePods = newPods
}

type MockEndpoint struct {
	allocateFunc func(devs []string) (*pluginapi.AllocateResponse, error)
}

func (m *MockEndpoint) stop() {}
func (m *MockEndpoint) run()  {}

func (m *MockEndpoint) getDevices() []pluginapi.Device {
	return []pluginapi.Device{}
}

func (m *MockEndpoint) callback(resourceName string, added, updated, deleted []pluginapi.Device) {}

func (m *MockEndpoint) allocate(devs []string) (*pluginapi.AllocateResponse, error) {
	if m.allocateFunc != nil {
		return m.allocateFunc(devs)
	}
	return nil, nil
}

func makePod(requests v1.ResourceList) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: uuid.NewUUID(),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: requests,
					},
				},
			},
		},
	}
}

func getTestManager(tmpDir string, activePods ActivePodsFunc, testRes []TestResource) *ManagerImpl {
	monitorCallback := func(resourceName string, added, updated, deleted []pluginapi.Device) {}
	testManager := &ManagerImpl{
		socketdir:    tmpDir,
		callback:     monitorCallback,
		allDevices:   make(map[string]sets.String),
		endpoints:    make(map[string]endpoint),
		allocation:   newAllocation(),
		activePods:   activePods,
		sourcesReady: &sourcesReadyStub{},
	}
	testManager.store, _ = utilstore.NewFileStore("/tmp/", utilfs.DefaultFs{})
	for _, res := range testRes {
		testManager.allDevices[res.resourceName] = sets.NewString()
		for _, dev := range res.devs {
			testManager.allDevices[res.resourceName].Insert(dev)
		}
		if res.resourceName == "domain1.com/resource1" {
			testManager.endpoints[res.resourceName] = &MockEndpoint{
				allocateFunc: func(devs []string) (*pluginapi.AllocateResponse, error) {
					resp := new(pluginapi.AllocateResponse)
					resp.Envs = make(map[string]string)
					for _, dev := range devs {
						switch dev {
						case "dev1":
							resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
								ContainerPath: "/dev/aaa",
								HostPath:      "/dev/aaa",
								Permissions:   "mrw",
							})

							resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
								ContainerPath: "/dev/bbb",
								HostPath:      "/dev/bbb",
								Permissions:   "mrw",
							})

							resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
								ContainerPath: "/container_dir1/file1",
								HostPath:      "host_dir1/file1",
								ReadOnly:      true,
							})

						case "dev2":
							resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
								ContainerPath: "/dev/ccc",
								HostPath:      "/dev/ccc",
								Permissions:   "mrw",
							})

							resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
								ContainerPath: "/container_dir1/file2",
								HostPath:      "host_dir1/file2",
								ReadOnly:      true,
							})

							resp.Envs["key1"] = "val1"
						}
					}
					return resp, nil
				},
			}
		}
		if res.resourceName == "domain2.com/resource2" {
			testManager.endpoints[res.resourceName] = &MockEndpoint{
				allocateFunc: func(devs []string) (*pluginapi.AllocateResponse, error) {
					resp := new(pluginapi.AllocateResponse)
					resp.Envs = make(map[string]string)
					for _, dev := range devs {
						switch dev {
						case "dev3":
							resp.Envs["key2"] = "val2"

						case "dev4":
							resp.Envs["key2"] = "val3"
						}
					}
					return resp, nil
				},
			}
		}
	}
	return testManager
}

func getTestNodeInfo(allocatable v1.ResourceList) *schedulercache.NodeInfo {
	cachedNode := &v1.Node{
		Status: v1.NodeStatus{
			Allocatable: allocatable,
		},
	}
	nodeInfo := &schedulercache.NodeInfo{}
	nodeInfo.SetNode(cachedNode)
	return nodeInfo
}

type TestResource struct {
	resourceName     string
	resourceQuantity resource.Quantity
	devs             []string
}

func TestPodContainerDeviceAllocation(t *testing.T) {
	flag.Set("alsologtostderr", fmt.Sprintf("%t", true))
	var logLevel string
	flag.StringVar(&logLevel, "logLevel", "4", "test")
	flag.Lookup("v").Value.Set(logLevel)
	res1 := TestResource{
		resourceName:     "domain1.com/resource1",
		resourceQuantity: *resource.NewQuantity(int64(2), resource.DecimalSI),
		devs:             []string{"dev1", "dev2"},
	}
	res2 := TestResource{
		resourceName:     "domain2.com/resource2",
		resourceQuantity: *resource.NewQuantity(int64(1), resource.DecimalSI),
		devs:             []string{"dev3", "dev4"},
	}
	testResources := make([]TestResource, 2)
	testResources = append(testResources, res1)
	testResources = append(testResources, res2)
	as := require.New(t)
	podsStub := activePodsStub{
		activePods: []*v1.Pod{},
	}
	tmpDir, err := ioutil.TempDir("", "checkpoint")
	as.Nil(err)
	defer os.RemoveAll(tmpDir)
	nodeInfo := getTestNodeInfo(v1.ResourceList{})
	testManager := getTestManager(tmpDir, podsStub.getActivePods, testResources)

	testPods := []*v1.Pod{
		makePod(v1.ResourceList{
			v1.ResourceName(res1.resourceName): res1.resourceQuantity,
			v1.ResourceName("cpu"):             res1.resourceQuantity,
			v1.ResourceName(res2.resourceName): res2.resourceQuantity}),
		makePod(v1.ResourceList{
			v1.ResourceName(res1.resourceName): res2.resourceQuantity}),
		makePod(v1.ResourceList{
			v1.ResourceName(res2.resourceName): res2.resourceQuantity}),
	}
	testCases := []struct {
		description               string
		testPod                   *v1.Pod
		expectedContainerOptsLen  []int
		expectedAllocatedResName1 int
		expectedAllocatedResName2 int
		expErr                    error
	}{
		{
			description:               "Successfull allocation of two Res1 resources and one Res2 resource",
			testPod:                   testPods[0],
			expectedContainerOptsLen:  []int{3, 2, 2},
			expectedAllocatedResName1: 2,
			expectedAllocatedResName2: 1,
			expErr: nil,
		},
		{
			description:               "Requesting to create a pod without enough resources should fail",
			testPod:                   testPods[1],
			expectedContainerOptsLen:  nil,
			expectedAllocatedResName1: 2,
			expectedAllocatedResName2: 1,
			expErr: fmt.Errorf("insufficient resource domain1.com/resource1: requested=1, available=0"),
		},
		{
			description:               "Successfull allocation of all available Res1 resources and Res2 resources",
			testPod:                   testPods[2],
			expectedContainerOptsLen:  []int{0, 0, 1},
			expectedAllocatedResName1: 2,
			expectedAllocatedResName2: 2,
			expErr: nil,
		},
	}
	activePods := []*v1.Pod{}
	for _, testCase := range testCases {
		pod := testCase.testPod
		activePods = append(activePods, pod)
		podsStub.updateActivePods(activePods)
		err := testManager.Allocate(nodeInfo, &lifecycle.PodAdmitAttributes{Pod: pod})
		if !reflect.DeepEqual(err, testCase.expErr) {
			t.Errorf("DevicePluginManager error (%v). expected error: %v but got: %v",
				testCase.description, testCase.expErr, err)
		}
		runContainerOpts := testManager.GetDeviceRunContainerOptions(pod, &pod.Spec.Containers[0])
		if testCase.expectedContainerOptsLen == nil {
			as.Nil(runContainerOpts)
		} else {
			as.Equal(len(runContainerOpts.Devices), testCase.expectedContainerOptsLen[0])
			as.Equal(len(runContainerOpts.Mounts), testCase.expectedContainerOptsLen[1])
			as.Equal(len(runContainerOpts.Envs), testCase.expectedContainerOptsLen[2])
		}
		as.Equal(testCase.expectedAllocatedResName1, testManager.allocation.devices()[res1.resourceName].Len())
		as.Equal(testCase.expectedAllocatedResName2, testManager.allocation.devices()[res2.resourceName].Len())
	}

}

func TestInitContainerDeviceAllocation(t *testing.T) {
	// Requesting to create a pod that requests resourceName1 in init containers and normal containers
	// should succeed with devices allocated to init containers reallocated to normal containers.
	res1 := TestResource{
		resourceName:     "domain1.com/resource1",
		resourceQuantity: *resource.NewQuantity(int64(2), resource.DecimalSI),
		devs:             []string{"dev1", "dev2"},
	}
	res2 := TestResource{
		resourceName:     "domain2.com/resource2",
		resourceQuantity: *resource.NewQuantity(int64(1), resource.DecimalSI),
		devs:             []string{"dev3", "dev4"},
	}
	testResources := make([]TestResource, 2)
	testResources = append(testResources, res1)
	testResources = append(testResources, res2)
	as := require.New(t)
	podsStub := activePodsStub{
		activePods: []*v1.Pod{},
	}
	nodeInfo := getTestNodeInfo(v1.ResourceList{})
	tmpDir, err := ioutil.TempDir("", "checkpoint")
	as.Nil(err)
	defer os.RemoveAll(tmpDir)
	testManager := getTestManager(tmpDir, podsStub.getActivePods, testResources)

	podWithPluginResourcesInInitContainers := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: uuid.NewUUID(),
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
				{
					Name: string(uuid.NewUUID()),
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(res1.resourceName): res2.resourceQuantity,
						},
					},
				},
				{
					Name: string(uuid.NewUUID()),
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(res1.resourceName): res1.resourceQuantity,
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name: string(uuid.NewUUID()),
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(res1.resourceName): res2.resourceQuantity,
							v1.ResourceName(res2.resourceName): res2.resourceQuantity,
						},
					},
				},
				{
					Name: string(uuid.NewUUID()),
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceName(res1.resourceName): res2.resourceQuantity,
							v1.ResourceName(res2.resourceName): res2.resourceQuantity,
						},
					},
				},
			},
		},
	}
	podsStub.updateActivePods([]*v1.Pod{podWithPluginResourcesInInitContainers})
	err = testManager.Allocate(nodeInfo, &lifecycle.PodAdmitAttributes{Pod: podWithPluginResourcesInInitContainers})
	as.Nil(err)
	podUID := string(podWithPluginResourcesInInitContainers.UID)
	initCont1 := podWithPluginResourcesInInitContainers.Spec.InitContainers[0].Name
	initCont2 := podWithPluginResourcesInInitContainers.Spec.InitContainers[1].Name
	normalCont1 := podWithPluginResourcesInInitContainers.Spec.Containers[0].Name
	normalCont2 := podWithPluginResourcesInInitContainers.Spec.Containers[1].Name
	initCont1Devices := testManager.allocation.devicesByContainer(podUID, initCont1, res1.resourceName)
	initCont2Devices := testManager.allocation.devicesByContainer(podUID, initCont2, res1.resourceName)
	normalCont1Devices := testManager.allocation.devicesByContainer(podUID, normalCont1, res1.resourceName)
	normalCont2Devices := testManager.allocation.devicesByContainer(podUID, normalCont2, res1.resourceName)
	as.True(initCont2Devices.IsSuperset(initCont1Devices))
	as.True(initCont2Devices.IsSuperset(normalCont1Devices))
	as.True(initCont2Devices.IsSuperset(normalCont2Devices))
	as.Equal(0, normalCont1Devices.Intersection(normalCont2Devices).Len())
}

func TestSanitizeNodeAllocatable(t *testing.T) {
	resourceName1 := "domain1.com/resource1"
	resourceName2 := "domain2.com/resource2"

	as := assert.New(t)
	monitorCallback := func(resourceName string, added, updated, deleted []pluginapi.Device) {}
	testManager := &ManagerImpl{
		callback:   monitorCallback,
		allDevices: make(map[string]sets.String),
		allocation: newAllocation(),
	}
	testManager.store, _ = utilstore.NewFileStore("/tmp/", utilfs.DefaultFs{})
	testManager.allocation.insert("p1", "c1", resourceName1, constructDevices([]string{"dev1", "dev2"}), nil)
	testManager.allocation.insert("p1", "c2", resourceName2, constructDevices([]string{"dev3"}), nil)

	cachedNode := &v1.Node{
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				// has no resource1 and two of resource2
				v1.ResourceName(resourceName2): *resource.NewQuantity(int64(2), resource.DecimalSI),
			},
		},
	}
	nodeInfo := &schedulercache.NodeInfo{}
	nodeInfo.SetNode(cachedNode)

	testManager.sanitizeNodeAllocatable(nodeInfo, "p1")

	allocatableScalarResources := nodeInfo.AllocatableResource().ScalarResources
	// allocatable in nodeInfo is less than needed, should update
	as.Equal(2, int(allocatableScalarResources[v1.ResourceName(resourceName1)]))
	// allocatable in nodeInfo is more than needed, should skip updating
	as.Equal(2, int(allocatableScalarResources[v1.ResourceName(resourceName2)]))
}
