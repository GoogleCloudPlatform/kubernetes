/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package apimachinery

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

// GroupMeta stores the metadata of a group.
type GroupMeta struct {
	// GroupVersion represents the preferred version of the group.
	GroupVersion unversioned.GroupVersion

	// GroupVersions is Group + all versions in that group.
	GroupVersions []unversioned.GroupVersion

	// Codec is the default codec for serializing output that should use
	// the preferred version.  Use this Codec when writing to
	// disk, a data store that is not dynamically versioned, or in tests.
	// This codec can decode any object that the schema is aware of.
	Codec runtime.Codec

	// SelfLinker can set or get the SelfLink field of all API types.
	// TODO: when versioning changes, make this part of each API definition.
	// TODO(lavalamp): Combine SelfLinker & ResourceVersioner interfaces, force all uses
	// to go through the InterfacesFor method below.
	SelfLinker runtime.SelfLinker

	// RESTMapper provides the default mapping between REST paths and the objects declared in api.Scheme and all known
	// versions.
	RESTMapper meta.RESTMapper

	// InterfacesByVersion stores the per-version interfaces.
	InterfacesByVersion map[unversioned.GroupVersion]*meta.VersionInterfaces
}

// InterfacesFor returns the default Codec and ResourceVersioner for a given version
// string, or an error if the version is not known.
func (gm *GroupMeta) InterfacesFor(version unversioned.GroupVersion) (*meta.VersionInterfaces, error) {
	if v, ok := gm.InterfacesByVersion[version]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("unsupported storage version: %s (valid: %v)", version, gm.GroupVersions)
}

// Adds the given version to the group. Only call during init, after that
// GroupMeta objects should be immutable. Not thread safe.
func (gm *GroupMeta) AddVersion(version unversioned.GroupVersion, interfaces *meta.VersionInterfaces) error {
	if e, a := gm.GroupVersion.Group, version.Group; a != e {
		return fmt.Errorf("got a version in group %v, but am in group %v", a, e)
	}
	if gm.InterfacesByVersion == nil {
		gm.InterfacesByVersion = make(map[unversioned.GroupVersion]*meta.VersionInterfaces)
	}
	gm.InterfacesByVersion[version] = interfaces
	gm.GroupVersions = append(gm.GroupVersions, version)
	return nil
}
