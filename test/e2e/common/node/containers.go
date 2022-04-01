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

package node

import (
	"github.com/onsi/gomega"
	e2eutils "k8s.io/kubernetes/test/e2e/framework/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/test/e2e/framework"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	admissionapi "k8s.io/pod-security-admission/api"
)

var _ = SIGDescribe("Containers", func() {
	f := framework.NewDefaultFramework("containers")
	f.NamespacePodSecurityEnforceLevel = admissionapi.LevelBaseline

	/*
		Release: v1.9
		Testname: Containers, without command and arguments
		Description: Default command and arguments from the container image entrypoint MUST be used when Pod does not specify the container command
	*/
	framework.ConformanceIt("should use the image defaults if command and args are blank [NodeConformance]", func() {
		pod := entrypointTestPod(f.Namespace.Name)
		pod.Spec.Containers[0].Args = nil
		pod = f.PodClient().Create(pod)
		err := e2epod.WaitForPodNameRunningInNamespace(f.ClientSet, pod.Name, f.Namespace.Name)
		e2eutils.ExpectNoError(err, "Expected pod %q to be running, got error: %v", pod.Name, err)
		pollLogs := func() (string, error) {
			return e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, pod.Name, pod.Spec.Containers[0].Name)
		}

		// The agnhost's image default entrypoint / args are: "/agnhost pause"
		// which will print out "Paused".
		gomega.Eventually(pollLogs, 3, e2eutils.Poll).Should(gomega.ContainSubstring("Paused"))
	})

	/*
		Release: v1.9
		Testname: Containers, with arguments
		Description: Default command and  from the container image entrypoint MUST be used when Pod does not specify the container command but the arguments from Pod spec MUST override when specified.
	*/
	framework.ConformanceIt("should be able to override the image's default arguments (container cmd) [NodeConformance]", func() {
		pod := entrypointTestPod(f.Namespace.Name, "entrypoint-tester", "override", "arguments")
		f.TestContainerOutput("override arguments", pod, 0, []string{
			"[/agnhost entrypoint-tester override arguments]",
		})
	})

	// Note: when you override the entrypoint, the image's arguments (container cmd)
	// are ignored.
	/*
		Release: v1.9
		Testname: Containers, with command
		Description: Default command from the container image entrypoint MUST NOT be used when Pod specifies the container command.  Command from Pod spec MUST override the command in the image.
	*/
	framework.ConformanceIt("should be able to override the image's default command (container entrypoint) [NodeConformance]", func() {
		pod := entrypointTestPod(f.Namespace.Name, "entrypoint-tester")
		pod.Spec.Containers[0].Command = []string{"/agnhost-2"}

		f.TestContainerOutput("override command", pod, 0, []string{
			"[/agnhost-2 entrypoint-tester]",
		})
	})

	/*
		Release: v1.9
		Testname: Containers, with command and arguments
		Description: Default command and arguments from the container image entrypoint MUST NOT be used when Pod specifies the container command and arguments.  Command and arguments from Pod spec MUST override the command and arguments in the image.
	*/
	framework.ConformanceIt("should be able to override the image's default command and arguments [NodeConformance]", func() {
		pod := entrypointTestPod(f.Namespace.Name, "entrypoint-tester", "override", "arguments")
		pod.Spec.Containers[0].Command = []string{"/agnhost-2"}

		f.TestContainerOutput("override all", pod, 0, []string{
			"[/agnhost-2 entrypoint-tester override arguments]",
		})
	})
})

// Return a prototypical entrypoint test pod
func entrypointTestPod(namespace string, entrypointArgs ...string) *v1.Pod {
	podName := "client-containers-" + string(uuid.NewUUID())
	pod := e2epod.NewAgnhostPod(namespace, podName, nil, nil, nil, entrypointArgs...)

	one := int64(1)
	pod.Spec.TerminationGracePeriodSeconds = &one
	pod.Spec.RestartPolicy = v1.RestartPolicyNever
	return pod
}
