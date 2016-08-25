/*
Copyright 2015 The Kubernetes Authors.

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

package flocker

import (
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/volume"

	"github.com/stretchr/testify/assert"
)

func TestFlockerUtil_CreateVolume(t *testing.T) {
	assert := assert.New(t)

	// test CreateVolume happy path
	options := volume.VolumeOptions{
		Capacity: resource.MustParse("3Gi"),
		AccessModes: []api.PersistentVolumeAccessMode{
			api.ReadWriteOnce,
		},
		PersistentVolumeReclaimPolicy: api.PersistentVolumeReclaimDelete,
	}

	provisioner := newTestableProvisioner(assert, options).(*flockerVolumeProvisioner)
	provisioner.flockerClient = newFakeFlockerClient()

	flockerUtil := &FlockerUtil{}

	datasetID, size, _, err := flockerUtil.CreateVolume(provisioner)
	assert.NoError(err)
	assert.Equal(datasetOneID, datasetID)
	assert.Equal(3, size)

}
