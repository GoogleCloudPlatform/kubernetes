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

package qos

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/features"
)

var supportedQoSComputeResources = sets.New(string(core.ResourceCPU), string(core.ResourceMemory))

// QOSList is a set of (resource name, QoS class) pairs.
type QOSList map[v1.ResourceName]v1.PodQOSClass

func isSupportedQoSComputeResource(name v1.ResourceName) bool {
	return supportedQoSComputeResources.Has(string(name))
}

// GetPodQOS returns the QoS class of a pod persisted in the PodStatus.QOSClass field.
// If PodStatus.QOSClass is empty, it returns value of ComputePodQOS() which evaluates pod's QoS class.
func GetPodQOS(pod *v1.Pod) v1.PodQOSClass {
	if pod.Status.QOSClass != "" {
		return pod.Status.QOSClass
	}
	return ComputePodQOS(pod)
}

// zeroQuantity represents a resource.Quantity with value "0", used as a baseline
// for resource comparisons.
var zeroQuantity = resource.MustParse("0")

// processResourceList adds non-zero quantities for supported QoS compute resources
// quantities from newList to list.
func processResourceList(list, newList v1.ResourceList) {
	for name, quantity := range newList {
		if !isSupportedQoSComputeResource(name) {
			continue
		}
		if quantity.Cmp(zeroQuantity) == 1 {
			delta := quantity.DeepCopy()
			if _, exists := list[name]; !exists {
				list[name] = delta
			} else {
				delta.Add(list[name])
				list[name] = delta
			}
		}
	}
}

// getQOSResources returns a set of resource names from the provided resource list that:
// 1. Are supported QoS compute resources
// 2. Have quantities greater than zero
func getQOSResources(list v1.ResourceList) sets.Set[string] {
	qosResources := sets.New[string]()
	for name, quantity := range list {
		if !isSupportedQoSComputeResource(name) {
			continue
		}
		if quantity.Cmp(zeroQuantity) == 1 {
			qosResources.Insert(string(name))
		}
	}
	return qosResources
}

// ComputePodQOS evaluates the list of containers to determine a pod's QoS class. This function is more
// expensive than GetPodQOS which should be used for pods having a non-empty .Status.QOSClass.
// A pod is besteffort if none of its containers have specified any requests or limits.
// A pod is guaranteed only when requests and limits are specified for all the containers and they are equal.
// A pod is burstable if limits and requests do not match across all containers.
// TODO(ndixita): Refactor ComputePodQOS into smaller functions to make it more
// readable and maintainable.
func ComputePodQOS(pod *v1.Pod) v1.PodQOSClass {
	requests := v1.ResourceList{}
	limits := v1.ResourceList{}
	isGuaranteed := true
	// When pod-level resources are specified, we use them to determine QoS class.
	if utilfeature.DefaultFeatureGate.Enabled(features.PodLevelResources) &&
		pod.Spec.Resources != nil {
		if len(pod.Spec.Resources.Requests) > 0 {
			// process requests
			processResourceList(requests, pod.Spec.Resources.Requests)
		}

		if len(pod.Spec.Resources.Limits) > 0 {
			// process limits
			processResourceList(limits, pod.Spec.Resources.Limits)
			qosLimitResources := getQOSResources(pod.Spec.Resources.Limits)
			if !qosLimitResources.HasAll(string(v1.ResourceMemory), string(v1.ResourceCPU)) {
				isGuaranteed = false
			}
		}
	} else {
		// note, ephemeral containers are not considered for QoS as they cannot define resources
		allContainers := []v1.Container{}
		allContainers = append(allContainers, pod.Spec.Containers...)
		allContainers = append(allContainers, pod.Spec.InitContainers...)
		for _, container := range allContainers {
			// process requests
			for name, quantity := range container.Resources.Requests {
				if !isSupportedQoSComputeResource(name) {
					continue
				}
				if quantity.Cmp(zeroQuantity) == 1 {
					delta := quantity.DeepCopy()
					if _, exists := requests[name]; !exists {
						requests[name] = delta
					} else {
						delta.Add(requests[name])
						requests[name] = delta
					}
				}
			}
			// process limits
			qosLimitsFound := sets.New[string]()
			for name, quantity := range container.Resources.Limits {
				if !isSupportedQoSComputeResource(name) {
					continue
				}
				if quantity.Cmp(zeroQuantity) == 1 {
					qosLimitsFound.Insert(string(name))
					delta := quantity.DeepCopy()
					if _, exists := limits[name]; !exists {
						limits[name] = delta
					} else {
						delta.Add(limits[name])
						limits[name] = delta
					}
				}
			}

			if !qosLimitsFound.HasAll(string(v1.ResourceMemory), string(v1.ResourceCPU)) {
				isGuaranteed = false
			}
		}
	}

	if len(requests) == 0 && len(limits) == 0 {
		return v1.PodQOSBestEffort
	}
	// Check is requests match limits for all resources.
	if isGuaranteed {
		for name, req := range requests {
			if lim, exists := limits[name]; !exists || lim.Cmp(req) != 0 {
				isGuaranteed = false
				break
			}
		}
	}
	if isGuaranteed &&
		len(requests) == len(limits) {
		return v1.PodQOSGuaranteed
	}
	return v1.PodQOSBurstable
}
