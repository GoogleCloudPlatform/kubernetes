/*
Copyright 2022 The Kubernetes Authors.

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

package debug

import (
	"fmt"

	"k8s.io/utils/pointer"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type legacyProfile struct {
}

type generalProfile struct {
}

type baselineProfile struct {
}

type restrictedProfile struct {
}

func (p *legacyProfile) Apply(pod *corev1.Pod, containerName string, target runtime.Object) error {
	switch target.(type) {
	case *corev1.Pod:
		// do nothing to the copied pod
		return nil
	case *corev1.Node:
		const volumeName = "host-root"
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/"},
			},
		})

		for i := range pod.Spec.Containers {
			container := &pod.Spec.Containers[i]
			if container.Name != containerName {
				continue
			}
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				MountPath: "/host",
				Name:      volumeName,
			})
		}

		setHostNamespace(pod, true)
		return nil
	default:
		return fmt.Errorf("the %s profile doesn't support objects of type %T", ProfileLegacy, target)
	}
}

func isPodCopy(pod *corev1.Pod, target runtime.Object) bool {
	if asserted, ok := target.(*corev1.Pod); !ok {
		return false
	} else {
		return pod != asserted
	}
}

type debugStyle int

const (
	styleEphemeral debugStyle = iota
	stylePodCopy
	styleNode
)

func getDebugStyle(pod *corev1.Pod, target runtime.Object) debugStyle {
	switch target.(type) {
	case *corev1.Pod:
		if isPodCopy(pod, target) {
			return stylePodCopy
		}
		return styleEphemeral
	case *corev1.Node:
		return styleNode
	}
	return debugStyle(-1) // unknown
}

func (p *generalProfile) Apply(pod *corev1.Pod, containerName string, target runtime.Object) error {
	switch getDebugStyle(pod, target) {
	case stylePodCopy:
		for i := range pod.Spec.Containers {
			container := &pod.Spec.Containers[i]
			if container.Name != containerName {
				continue
			}
			// For copy of pod: sets SYS_PTRACE in debugging container, sets shareProcessNamespace
			pod.Spec.ShareProcessNamespace = pointer.BoolPtr(true)
			if container.SecurityContext == nil {
				container.SecurityContext = &corev1.SecurityContext{}
			}
			container.SecurityContext.Capabilities = &corev1.Capabilities{
				Add: []corev1.Capability{"SYS_PTRACE"},
			}
		}

	case styleEphemeral:
		for i := range pod.Spec.EphemeralContainers {
			container := &pod.Spec.EphemeralContainers[i]
			if container.Name != containerName {
				continue
			}
			// For ephemeral container: sets SYS_PTRACE in ephemeral container
			container.SecurityContext = &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"SYS_PTRACE"},
				},
			}
		}

	case styleNode:
		// empty securityContext; uses host namespaces, mounts root partition
		const volumeName = "host-root"
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: "/"},
			},
		})
		for i := range pod.Spec.Containers {
			container := &pod.Spec.Containers[i]
			if container.Name != containerName {
				continue
			}
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				MountPath: "/host",
				Name:      volumeName,
			})
		}

		pod.Spec.SecurityContext = nil
		setHostNamespace(pod, true)
	default:
		return fmt.Errorf("the %s profile doesn't support objects of type %T", ProfileGeneral, target)
	}

	return nil
}

func (p *baselineProfile) Apply(pod *corev1.Pod, containerName string, target runtime.Object) error {
	switch getDebugStyle(pod, target) {
	case stylePodCopy:
		pod.Spec.SecurityContext = nil
		setHostNamespace(pod, false)

		for i := range pod.Spec.Containers {
			container := &pod.Spec.Containers[i]
			if container.Name != containerName {
				continue
			}
			// For copy of pod: empty securityContext; sets shareProcessNamespace
			container.SecurityContext = nil
			pod.Spec.ShareProcessNamespace = pointer.BoolPtr(true)
		}

	case styleEphemeral:
		for i := range pod.Spec.EphemeralContainers {
			container := &pod.Spec.EphemeralContainers[i]
			if container.Name != containerName {
				continue
			}
			// For ephemeral container: empty securityContext
			container.SecurityContext = nil
		}

	case styleNode:
		// empty securityContext; uses isolated namespaces
		pod.Spec.SecurityContext = nil
		setHostNamespace(pod, false)

	default:
		return fmt.Errorf("the %s profile doesn't support objects of type %T", ProfileBaseline, target)
	}

	return nil
}

func (p *restrictedProfile) Apply(pod *corev1.Pod, containerName string, target runtime.Object) error {
	switch getDebugStyle(pod, target) {
	case stylePodCopy:
		for i := range pod.Spec.Containers {
			container := &pod.Spec.Containers[i]
			if container.Name != containerName {
				continue
			}
			// For copy of pod: empty securityContext; sets shareProcessNamespace
			container.SecurityContext = &corev1.SecurityContext{
				RunAsNonRoot: pointer.BoolPtr(true),
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"ALL"},
				},
			}
			pod.Spec.ShareProcessNamespace = pointer.BoolPtr(true)
		}

	case styleEphemeral:
		for i := range pod.Spec.EphemeralContainers {
			container := &pod.Spec.EphemeralContainers[i]
			if container.Name != containerName {
				continue
			}
			// For ephemeral container: empty securityContext
			container.SecurityContext = &corev1.SecurityContext{
				RunAsNonRoot: pointer.BoolPtr(true),
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"ALL"},
				},
			}
		}

	case styleNode:
		// no additional settings required other than the common
		// settings applied below

	default:
		return fmt.Errorf("the %s profile doesn't support objects of type %T", ProfileRestricted, target)
	}

	// common settings
	pod.Spec.SecurityContext = nil
	setHostNamespace(pod, false)

	return nil
}

func setHostNamespace(pod *corev1.Pod, enabled bool) {
	pod.Spec.HostNetwork = enabled
	pod.Spec.HostPID = enabled
	pod.Spec.HostIPC = enabled
}
