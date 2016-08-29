/*
Copyright 2016 The Kubernetes Authors.

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

package libstorage

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/volume"
)

var _ volume.Provisioner = &lsVolume{}

func (m *lsVolume) Provision() (*api.PersistentVolume, error) {
	glog.V(4).Info("libStorage: attempting to provision volume %s", m.volName)

	// create volume and returns a libStorage Volume value
	vol, err := m.plugin.lsMgr.createVolume(m)
	if err != nil {
		return nil, err
	}

	// describe created pv
	pv := &api.PersistentVolume{
		ObjectMeta: api.ObjectMeta{
			Name:   m.options.PVName,
			Labels: map[string]string{},
			Annotations: map[string]string{
				"kubernetes.io/createdby": "libstorage-dynamic-provisioner",
			},
		},
		Spec: api.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: m.options.PersistentVolumeReclaimPolicy,
			AccessModes:                   m.options.AccessModes,
			Capacity: api.ResourceList{
				api.ResourceName(api.ResourceStorage): resource.MustParse(
					fmt.Sprintf("%dGi", vol.Size),
				),
			},
			PersistentVolumeSource: api.PersistentVolumeSource{
				LibStorage: &api.LibStorageVolumeSource{
					Host:       m.plugin.lsMgr.getHost(),
					Service:    m.plugin.lsMgr.getService(),
					VolumeName: vol.VolumeName(),
					FSType:     m.fsType,
					ReadOnly:   m.readOnly,
				},
			},
		},
	}

	glog.V(4).Infof("libStorage: provisioned PV: %#v", pv)
	return pv, nil
}
