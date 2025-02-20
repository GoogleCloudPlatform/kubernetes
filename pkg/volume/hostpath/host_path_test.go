/*
Copyright 2014 The Kubernetes Authors.

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

package hostpath

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2/ktesting"
	"k8s.io/kubernetes/pkg/volume"
	volumetest "k8s.io/kubernetes/pkg/volume/testing"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
	utilpath "k8s.io/utils/path"
	"k8s.io/utils/ptr"
)

func TestCanSupport(t *testing.T) {
	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{}), nil /* prober */, volumetest.NewFakeKubeletVolumeHost(t, "fake", nil, nil))

	plug, err := plugMgr.FindPluginByName(hostPathPluginName)
	if err != nil {
		t.Fatal("Can't find the plugin by name")
	}
	if plug.GetPluginName() != hostPathPluginName {
		t.Errorf("Wrong name: %s", plug.GetPluginName())
	}
	if !plug.CanSupport(&volume.Spec{Volume: &v1.Volume{VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{}}}}) {
		t.Errorf("Expected true")
	}
	if !plug.CanSupport(&volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{PersistentVolumeSource: v1.PersistentVolumeSource{HostPath: &v1.HostPathVolumeSource{}}}}}) {
		t.Errorf("Expected true")
	}
	if plug.CanSupport(&volume.Spec{Volume: &v1.Volume{VolumeSource: v1.VolumeSource{}}}) {
		t.Errorf("Expected false")
	}
}

func TestGetAccessModes(t *testing.T) {
	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{}), nil /* prober */, volumetest.NewFakeKubeletVolumeHost(t, "/tmp/fake", nil, nil))

	plug, err := plugMgr.FindPersistentPluginByName(hostPathPluginName)
	if err != nil {
		t.Fatal("Can't find the plugin by name")
	}
	if len(plug.GetAccessModes()) != 1 || plug.GetAccessModes()[0] != v1.ReadWriteOnce {
		t.Errorf("Expected %s PersistentVolumeAccessMode", v1.ReadWriteOnce)
	}
}

func TestRecycler(t *testing.T) {
	plugMgr := volume.VolumePluginMgr{}
	pluginHost := volumetest.NewFakeKubeletVolumeHost(t, "/tmp/fake", nil, nil)
	plugMgr.InitPlugins([]volume.VolumePlugin{&hostPathPlugin{nil, volume.VolumeConfig{}, false}}, nil, pluginHost)

	spec := &volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{PersistentVolumeSource: v1.PersistentVolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/foo"}}}}}
	_, err := plugMgr.FindRecyclablePluginBySpec(spec)
	if err != nil {
		t.Errorf("Can't find the plugin by name")
	}
}

func TestDeleter(t *testing.T) {
	// Deleter has a hard-coded regex for "/tmp".
	tempPath := fmt.Sprintf("/tmp/hostpath.%s", uuid.NewUUID())
	err := os.MkdirAll(tempPath, 0750)
	if err != nil {
		t.Fatalf("Failed to create tmp directory for deleter: %v", err)
	}
	defer os.RemoveAll(tempPath)
	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{}), nil /* prober */, volumetest.NewFakeKubeletVolumeHost(t, "/tmp/fake", nil, nil))

	spec := &volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{PersistentVolumeSource: v1.PersistentVolumeSource{HostPath: &v1.HostPathVolumeSource{Path: tempPath}}}}}
	plug, err := plugMgr.FindDeletablePluginBySpec(spec)
	if err != nil {
		t.Fatal("Can't find the plugin by name")
	}
	logger, _ := ktesting.NewTestContext(t)
	deleter, err := plug.NewDeleter(logger, spec)
	if err != nil {
		t.Errorf("Failed to make a new Deleter: %v", err)
	}
	if deleter.GetPath() != tempPath {
		t.Errorf("Expected %s but got %s", tempPath, deleter.GetPath())
	}
	if err := deleter.Delete(); err != nil {
		t.Errorf("Mock Recycler expected to return nil but got %s", err)
	}
	if exists, _ := utilpath.Exists(utilpath.CheckFollowSymlink, tempPath); exists {
		t.Errorf("Temp path expected to be deleted, but was found at %s", tempPath)
	}
}

func TestDeleterTempDir(t *testing.T) {
	tests := map[string]struct {
		expectedFailure bool
		path            string
	}{
		"just-tmp": {true, "/tmp"},
		"not-tmp":  {true, "/nottmp"},
		"good-tmp": {false, "/tmp/scratch"},
	}
	logger, _ := ktesting.NewTestContext(t)
	for name, test := range tests {
		plugMgr := volume.VolumePluginMgr{}
		plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{}), nil /* prober */, volumetest.NewFakeKubeletVolumeHost(t, "/tmp/fake", nil, nil))
		spec := &volume.Spec{PersistentVolume: &v1.PersistentVolume{Spec: v1.PersistentVolumeSpec{PersistentVolumeSource: v1.PersistentVolumeSource{HostPath: &v1.HostPathVolumeSource{Path: test.path}}}}}
		plug, _ := plugMgr.FindDeletablePluginBySpec(spec)
		deleter, _ := plug.NewDeleter(logger, spec)
		err := deleter.Delete()
		if err == nil && test.expectedFailure {
			t.Errorf("Expected failure for test '%s' but got nil err", name)
		}
		if err != nil && !test.expectedFailure {
			t.Errorf("Unexpected failure for test '%s': %v", name, err)
		}
	}
}

func TestProvisioner(t *testing.T) {
	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{ProvisioningEnabled: true}),
		nil,
		volumetest.NewFakeKubeletVolumeHost(t, "/tmp/fake", nil, nil))
	plug, err := plugMgr.FindProvisionablePluginByName(hostPathPluginName)
	if err != nil {
		t.Fatalf("Can't find the plugin by name")
	}
	options := volume.VolumeOptions{
		PVC:                           volumetest.CreateTestPVC("1Gi", []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}),
		PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete,
	}
	logger, _ := ktesting.NewTestContext(t)
	creator, err := plug.NewProvisioner(logger, options)
	if err != nil {
		t.Fatalf("Failed to make a new Provisioner: %v", err)
	}

	hostPathCreator, ok := creator.(*hostPathProvisioner)
	if !ok {
		t.Fatal("Not a hostPathProvisioner")
	}
	hostPathCreator.basePath = fmt.Sprintf("%s.%s", "hostPath_pv", uuid.NewUUID())

	pv, err := hostPathCreator.Provision(nil, nil)
	if err != nil {
		t.Errorf("Unexpected error creating volume: %v", err)
	}
	if pv.Spec.HostPath.Path == "" {
		t.Errorf("Expected pv.Spec.HostPath.Path to not be empty: %#v", pv)
	}
	expectedCapacity := resource.NewQuantity(1*1024*1024*1024, resource.BinarySI)
	actualCapacity := pv.Spec.Capacity[v1.ResourceStorage]
	expectedAmt := expectedCapacity.Value()
	actualAmt := actualCapacity.Value()
	if expectedAmt != actualAmt {
		t.Errorf("Expected capacity %+v but got %+v", expectedAmt, actualAmt)
	}

	if pv.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimDelete {
		t.Errorf("Expected reclaim policy %+v but got %+v", v1.PersistentVolumeReclaimDelete, pv.Spec.PersistentVolumeReclaimPolicy)
	}

	os.RemoveAll(hostPathCreator.basePath)

}

func TestInvalidHostPath(t *testing.T) {
	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{}), nil /* prober */, volumetest.NewFakeKubeletVolumeHost(t, "fake", nil, nil))

	plug, err := plugMgr.FindPluginByName(hostPathPluginName)
	if err != nil {
		t.Fatalf("Unable to find plugin %s by name: %v", hostPathPluginName, err)
	}
	spec := &v1.Volume{
		Name:         "vol1",
		VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/no/backsteps/allowed/.."}},
	}
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	mounter, err := plug.NewMounter(volume.NewSpecFromVolume(spec), pod)
	if err != nil {
		t.Fatal(err)
	}

	err = mounter.SetUp(volume.MounterArgs{})
	expectedMsg := "invalid HostPath `/no/backsteps/allowed/..`: must not contain '..'"
	if err.Error() != expectedMsg {
		t.Fatalf("expected error `%s` but got `%s`", expectedMsg, err)
	}
}

func TestPlugin(t *testing.T) {
	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{}), nil /* prober */, volumetest.NewFakeKubeletVolumeHost(t, "fake", nil, nil))

	plug, err := plugMgr.FindPluginByName(hostPathPluginName)
	if err != nil {
		t.Fatal("Can't find the plugin by name")
	}

	volPath := "/tmp/vol1"
	spec := &v1.Volume{
		Name:         "vol1",
		VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: volPath, Type: ptr.To(v1.HostPathDirectoryOrCreate)}},
	}
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	defer os.RemoveAll(volPath)
	mounter, err := plug.NewMounter(volume.NewSpecFromVolume(spec), pod)
	if err != nil {
		t.Errorf("Failed to make a new Mounter: %v", err)
	}
	if mounter == nil {
		t.Fatalf("Got a nil Mounter")
	}

	path := mounter.GetPath()
	if path != volPath {
		t.Errorf("Got unexpected path: %s", path)
	}

	if err := mounter.SetUp(volume.MounterArgs{}); err != nil {
		t.Errorf("Expected success, got: %v", err)
	}

	unmounter, err := plug.NewUnmounter("vol1", types.UID("poduid"))
	if err != nil {
		t.Errorf("Failed to make a new Unmounter: %v", err)
	}
	if unmounter == nil {
		t.Fatalf("Got a nil Unmounter")
	}

	if err := unmounter.TearDown(); err != nil {
		t.Errorf("Expected success, got: %v", err)
	}
}

func TestPersistentClaimReadOnlyFlag(t *testing.T) {
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pvA",
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{Path: "foo", Type: ptr.To(v1.HostPathDirectoryOrCreate)},
			},
			ClaimRef: &v1.ObjectReference{
				Name: "claimA",
			},
		},
	}
	defer os.RemoveAll("foo")

	claim := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "claimA",
			Namespace: "nsA",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName: "pvA",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimBound,
		},
	}

	client := fake.NewSimpleClientset(pv, claim)

	plugMgr := volume.VolumePluginMgr{}
	plugMgr.InitPlugins(ProbeVolumePlugins(volume.VolumeConfig{}), nil /* prober */, volumetest.NewFakeKubeletVolumeHost(t, "/tmp/fake", client, nil))
	plug, _ := plugMgr.FindPluginByName(hostPathPluginName)

	// readOnly bool is supplied by persistent-claim volume source when its mounter creates other volumes
	spec := volume.NewSpecFromPersistentVolume(pv, true)
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: types.UID("poduid")}}
	mounter, _ := plug.NewMounter(spec, pod)
	if mounter == nil {
		t.Fatalf("Got a nil Mounter")
	}

	if !mounter.GetAttributes().ReadOnly {
		t.Errorf("Expected true for mounter.IsReadOnly")
	}
}

func TestOSFileTypeChecker(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hostpathtest")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tempDir)
	testCases := []struct {
		name           string
		path           string
		pathType       *v1.HostPathType
		simulatedType  *hostutil.FileType
		expectError    bool
		checkExistence bool
		checkDirectory bool
	}{
		// The order of these test cases is significant, some of them create files or directories
		{
			name:        "Missing file",
			path:        filepath.Join(tempDir, "nosuchfile"),
			pathType:    ptr.To(v1.HostPathFile),
			expectError: true,
		},
		{
			name:           "File to create",
			path:           filepath.Join(tempDir, "newfile"),
			pathType:       ptr.To(v1.HostPathFileOrCreate),
			checkExistence: true,
		},
		{
			name:           "Created file",
			path:           filepath.Join(tempDir, "newfile"),
			pathType:       ptr.To(v1.HostPathFile),
			checkExistence: true,
		},
		{
			name:           "Created file (mounted with OrCreate)",
			path:           filepath.Join(tempDir, "newfile"),
			pathType:       ptr.To(v1.HostPathFileOrCreate),
			checkExistence: true,
		},
		{
			name:        "Directory conflicting with a file",
			path:        filepath.Join(tempDir, "newfile"),
			pathType:    ptr.To(v1.HostPathDirectory),
			expectError: true,
		},
		{
			name:        "Missing directory",
			path:        filepath.Join(tempDir, "nosuchdirectory"),
			pathType:    ptr.To(v1.HostPathDirectory),
			expectError: true,
		},
		{
			name:           "Directory to create",
			path:           filepath.Join(tempDir, "newdirectory"),
			pathType:       ptr.To(v1.HostPathDirectoryOrCreate),
			checkExistence: true,
			checkDirectory: true,
		},
		{
			name:           "Created directory",
			path:           filepath.Join(tempDir, "newdirectory"),
			pathType:       ptr.To(v1.HostPathDirectory),
			checkExistence: true,
			checkDirectory: true,
		},
		{
			name:           "Created directory (mounted with OrCreate)",
			path:           filepath.Join(tempDir, "newdirectory"),
			pathType:       ptr.To(v1.HostPathDirectoryOrCreate),
			checkExistence: true,
			checkDirectory: true,
		},
		{
			name:          "Simulated Socket File",
			path:          filepath.Join(tempDir, "socket"),
			pathType:      ptr.To(v1.HostPathSocket),
			simulatedType: ptr.To(hostutil.FileTypeSocket),
		},
		{
			name:          "Simulated Character Device",
			path:          filepath.Join(tempDir, "chardev"),
			pathType:      ptr.To(v1.HostPathCharDev),
			simulatedType: ptr.To(hostutil.FileTypeCharDev),
		},
		{
			name:          "Simulated Block Device",
			path:          filepath.Join(tempDir, "blockdev"),
			pathType:      ptr.To(v1.HostPathBlockDev),
			simulatedType: ptr.To(hostutil.FileTypeBlockDev),
		},
	}

	for i, tc := range testCases {
		var hu hostutil.HostUtils
		if tc.simulatedType != nil {
			hu = hostutil.NewFakeHostUtil(
				map[string]hostutil.FileType{
					tc.path: hostutil.FileType(*tc.simulatedType),
				})
		} else {
			hu = hostutil.NewHostUtil()
		}

		err := checkType(tc.path, tc.pathType, hu)
		if tc.expectError && err == nil {
			t.Errorf("[%d: %s] expected an error but didn't get one", i, tc.name)
		}
		if !tc.expectError && err != nil {
			t.Errorf("[%d: %s] didn't expect an error but got one: %v", i, tc.name, err)
		}
		if tc.checkExistence {
			fi, err := os.Stat(tc.path)
			if errors.Is(err, os.ErrNotExist) {
				t.Errorf("[%d: %q] path %q does not exist", i, tc.name, tc.path)
			}
			if tc.checkDirectory && !fi.IsDir() {
				t.Errorf("[%d: %q] path %q is not a directory", i, tc.name, tc.path)
			}
		}
	}
}
