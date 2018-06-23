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

// Code generated by deepcopy-gen. DO NOT EDIT.

package storage

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/kubernetes/pkg/apis/core"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSElasticBlockStoreVolumeSnapshotSource) DeepCopyInto(out *AWSElasticBlockStoreVolumeSnapshotSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSElasticBlockStoreVolumeSnapshotSource.
func (in *AWSElasticBlockStoreVolumeSnapshotSource) DeepCopy() *AWSElasticBlockStoreVolumeSnapshotSource {
	if in == nil {
		return nil
	}
	out := new(AWSElasticBlockStoreVolumeSnapshotSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CSIVolumeSnapshotSource) DeepCopyInto(out *CSIVolumeSnapshotSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CSIVolumeSnapshotSource.
func (in *CSIVolumeSnapshotSource) DeepCopy() *CSIVolumeSnapshotSource {
	if in == nil {
		return nil
	}
	out := new(CSIVolumeSnapshotSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CinderVolumeSnapshotSource) DeepCopyInto(out *CinderVolumeSnapshotSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CinderVolumeSnapshotSource.
func (in *CinderVolumeSnapshotSource) DeepCopy() *CinderVolumeSnapshotSource {
	if in == nil {
		return nil
	}
	out := new(CinderVolumeSnapshotSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCEPersistentDiskSnapshotSource) DeepCopyInto(out *GCEPersistentDiskSnapshotSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCEPersistentDiskSnapshotSource.
func (in *GCEPersistentDiskSnapshotSource) DeepCopy() *GCEPersistentDiskSnapshotSource {
	if in == nil {
		return nil
	}
	out := new(GCEPersistentDiskSnapshotSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GlusterVolumeSnapshotSource) DeepCopyInto(out *GlusterVolumeSnapshotSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GlusterVolumeSnapshotSource.
func (in *GlusterVolumeSnapshotSource) DeepCopy() *GlusterVolumeSnapshotSource {
	if in == nil {
		return nil
	}
	out := new(GlusterVolumeSnapshotSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostPathVolumeSnapshotSource) DeepCopyInto(out *HostPathVolumeSnapshotSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostPathVolumeSnapshotSource.
func (in *HostPathVolumeSnapshotSource) DeepCopy() *HostPathVolumeSnapshotSource {
	if in == nil {
		return nil
	}
	out := new(HostPathVolumeSnapshotSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageClass) DeepCopyInto(out *StorageClass) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ReclaimPolicy != nil {
		in, out := &in.ReclaimPolicy, &out.ReclaimPolicy
		*out = new(core.PersistentVolumeReclaimPolicy)
		**out = **in
	}
	if in.MountOptions != nil {
		in, out := &in.MountOptions, &out.MountOptions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowVolumeExpansion != nil {
		in, out := &in.AllowVolumeExpansion, &out.AllowVolumeExpansion
		*out = new(bool)
		**out = **in
	}
	if in.VolumeBindingMode != nil {
		in, out := &in.VolumeBindingMode, &out.VolumeBindingMode
		*out = new(VolumeBindingMode)
		**out = **in
	}
	if in.AllowedTopologies != nil {
		in, out := &in.AllowedTopologies, &out.AllowedTopologies
		*out = make([]core.TopologySelectorTerm, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SnapshotParameters != nil {
		in, out := &in.SnapshotParameters, &out.SnapshotParameters
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageClass.
func (in *StorageClass) DeepCopy() *StorageClass {
	if in == nil {
		return nil
	}
	out := new(StorageClass)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageClass) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageClassList) DeepCopyInto(out *StorageClassList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StorageClass, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageClassList.
func (in *StorageClassList) DeepCopy() *StorageClassList {
	if in == nil {
		return nil
	}
	out := new(StorageClassList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StorageClassList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachment) DeepCopyInto(out *VolumeAttachment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachment.
func (in *VolumeAttachment) DeepCopy() *VolumeAttachment {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeAttachment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentList) DeepCopyInto(out *VolumeAttachmentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VolumeAttachment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentList.
func (in *VolumeAttachmentList) DeepCopy() *VolumeAttachmentList {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeAttachmentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentSource) DeepCopyInto(out *VolumeAttachmentSource) {
	*out = *in
	if in.PersistentVolumeName != nil {
		in, out := &in.PersistentVolumeName, &out.PersistentVolumeName
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentSource.
func (in *VolumeAttachmentSource) DeepCopy() *VolumeAttachmentSource {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentSpec) DeepCopyInto(out *VolumeAttachmentSpec) {
	*out = *in
	in.Source.DeepCopyInto(&out.Source)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentSpec.
func (in *VolumeAttachmentSpec) DeepCopy() *VolumeAttachmentSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeAttachmentStatus) DeepCopyInto(out *VolumeAttachmentStatus) {
	*out = *in
	if in.AttachmentMetadata != nil {
		in, out := &in.AttachmentMetadata, &out.AttachmentMetadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AttachError != nil {
		in, out := &in.AttachError, &out.AttachError
		*out = new(VolumeError)
		(*in).DeepCopyInto(*out)
	}
	if in.DetachError != nil {
		in, out := &in.DetachError, &out.DetachError
		*out = new(VolumeError)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeAttachmentStatus.
func (in *VolumeAttachmentStatus) DeepCopy() *VolumeAttachmentStatus {
	if in == nil {
		return nil
	}
	out := new(VolumeAttachmentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeError) DeepCopyInto(out *VolumeError) {
	*out = *in
	in.Time.DeepCopyInto(&out.Time)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeError.
func (in *VolumeError) DeepCopy() *VolumeError {
	if in == nil {
		return nil
	}
	out := new(VolumeError)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshot) DeepCopyInto(out *VolumeSnapshot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshot.
func (in *VolumeSnapshot) DeepCopy() *VolumeSnapshot {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshot)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeSnapshot) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotCondition) DeepCopyInto(out *VolumeSnapshotCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotCondition.
func (in *VolumeSnapshotCondition) DeepCopy() *VolumeSnapshotCondition {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotData) DeepCopyInto(out *VolumeSnapshotData) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotData.
func (in *VolumeSnapshotData) DeepCopy() *VolumeSnapshotData {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeSnapshotData) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotDataCondition) DeepCopyInto(out *VolumeSnapshotDataCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotDataCondition.
func (in *VolumeSnapshotDataCondition) DeepCopy() *VolumeSnapshotDataCondition {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotDataCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotDataList) DeepCopyInto(out *VolumeSnapshotDataList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VolumeSnapshotData, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotDataList.
func (in *VolumeSnapshotDataList) DeepCopy() *VolumeSnapshotDataList {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotDataList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeSnapshotDataList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotDataSource) DeepCopyInto(out *VolumeSnapshotDataSource) {
	*out = *in
	if in.HostPath != nil {
		in, out := &in.HostPath, &out.HostPath
		if *in == nil {
			*out = nil
		} else {
			*out = new(HostPathVolumeSnapshotSource)
			**out = **in
		}
	}
	if in.GlusterSnapshotVolume != nil {
		in, out := &in.GlusterSnapshotVolume, &out.GlusterSnapshotVolume
		if *in == nil {
			*out = nil
		} else {
			*out = new(GlusterVolumeSnapshotSource)
			**out = **in
		}
	}
	if in.AWSElasticBlockStore != nil {
		in, out := &in.AWSElasticBlockStore, &out.AWSElasticBlockStore
		if *in == nil {
			*out = nil
		} else {
			*out = new(AWSElasticBlockStoreVolumeSnapshotSource)
			**out = **in
		}
	}
	if in.GCEPersistentDiskSnapshot != nil {
		in, out := &in.GCEPersistentDiskSnapshot, &out.GCEPersistentDiskSnapshot
		if *in == nil {
			*out = nil
		} else {
			*out = new(GCEPersistentDiskSnapshotSource)
			**out = **in
		}
	}
	if in.CinderSnapshot != nil {
		in, out := &in.CinderSnapshot, &out.CinderSnapshot
		if *in == nil {
			*out = nil
		} else {
			*out = new(CinderVolumeSnapshotSource)
			**out = **in
		}
	}
	if in.CSISnapshot != nil {
		in, out := &in.CSISnapshot, &out.CSISnapshot
		if *in == nil {
			*out = nil
		} else {
			*out = new(CSIVolumeSnapshotSource)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotDataSource.
func (in *VolumeSnapshotDataSource) DeepCopy() *VolumeSnapshotDataSource {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotDataSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotDataSpec) DeepCopyInto(out *VolumeSnapshotDataSpec) {
	*out = *in
	in.VolumeSnapshotDataSource.DeepCopyInto(&out.VolumeSnapshotDataSource)
	if in.VolumeSnapshotRef != nil {
		in, out := &in.VolumeSnapshotRef, &out.VolumeSnapshotRef
		if *in == nil {
			*out = nil
		} else {
			*out = new(core.ObjectReference)
			**out = **in
		}
	}
	if in.PersistentVolumeRef != nil {
		in, out := &in.PersistentVolumeRef, &out.PersistentVolumeRef
		if *in == nil {
			*out = nil
		} else {
			*out = new(core.ObjectReference)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotDataSpec.
func (in *VolumeSnapshotDataSpec) DeepCopy() *VolumeSnapshotDataSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotDataSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotDataStatus) DeepCopyInto(out *VolumeSnapshotDataStatus) {
	*out = *in
	in.CreationTimestamp.DeepCopyInto(&out.CreationTimestamp)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]VolumeSnapshotDataCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotDataStatus.
func (in *VolumeSnapshotDataStatus) DeepCopy() *VolumeSnapshotDataStatus {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotDataStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotList) DeepCopyInto(out *VolumeSnapshotList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VolumeSnapshot, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotList.
func (in *VolumeSnapshotList) DeepCopy() *VolumeSnapshotList {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeSnapshotList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotSpec) DeepCopyInto(out *VolumeSnapshotSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotSpec.
func (in *VolumeSnapshotSpec) DeepCopy() *VolumeSnapshotSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSnapshotStatus) DeepCopyInto(out *VolumeSnapshotStatus) {
	*out = *in
	in.CreationTimestamp.DeepCopyInto(&out.CreationTimestamp)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]VolumeSnapshotCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSnapshotStatus.
func (in *VolumeSnapshotStatus) DeepCopy() *VolumeSnapshotStatus {
	if in == nil {
		return nil
	}
	out := new(VolumeSnapshotStatus)
	in.DeepCopyInto(out)
	return out
}
