/*
Copyright 2025 The Kubernetes Authors.

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

package resourceslicepatch

import (
	"context"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/resource"
	"k8s.io/kubernetes/pkg/apis/resource/validation"
	"k8s.io/kubernetes/pkg/features"
)

// resourceSlicePatchStrategy implements behavior for ResourceSlicePatch objects
type resourceSlicePatchStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = resourceSlicePatchStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

func (resourceSlicePatchStrategy) NamespaceScoped() bool {
	return false
}

func (resourceSlicePatchStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	patch := obj.(*resource.ResourceSlicePatch)
	patch.Generation = 1

	dropDisabledFields(patch, nil)
}

func (resourceSlicePatchStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	patch := obj.(*resource.ResourceSlicePatch)
	return validation.ValidateResourceSlicePatch(patch)
}

func (resourceSlicePatchStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (resourceSlicePatchStrategy) Canonicalize(obj runtime.Object) {
}

func (resourceSlicePatchStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (resourceSlicePatchStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	patch := obj.(*resource.ResourceSlicePatch)
	oldPatch := old.(*resource.ResourceSlicePatch)

	// Any changes to the spec increment the generation number.
	if !apiequality.Semantic.DeepEqual(oldPatch.Spec, patch.Spec) {
		patch.Generation = oldPatch.Generation + 1
	}

	dropDisabledFields(patch, oldPatch)
}

func (resourceSlicePatchStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateResourceSlicePatchUpdate(obj.(*resource.ResourceSlicePatch), old.(*resource.ResourceSlicePatch))
}

func (resourceSlicePatchStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

func (resourceSlicePatchStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// dropDisabledFields removes fields which are covered by a feature gate.
func dropDisabledFields(newPatch, oldPatch *resource.ResourceSlicePatch) {
	dropDisabledDRAAdminControlledDeviceAttributesFields(newPatch, oldPatch)
}

func dropDisabledDRAAdminControlledDeviceAttributesFields(newPatch, oldPatch *resource.ResourceSlicePatch) {
	if utilfeature.DefaultFeatureGate.Enabled(features.DRAAdminControlledDeviceAttributes) {
		// No need to drop anything.
		return
	}
	if draAdminControlledDeviceAttributesFeatureInUse(oldPatch) {
		// If anything was set in the past, then fields must not get
		// dropped on potentially unrelated updates.
		return
	}

	newPatch.Spec.Devices.Attributes = nil
	newPatch.Spec.Devices.Capacity = nil
}

func draAdminControlledDeviceAttributesFeatureInUse(patch *resource.ResourceSlicePatch) bool {
	if patch == nil {
		return false
	}

	return patch.Spec.Devices.Attributes != nil ||
		patch.Spec.Devices.Capacity != nil
}
