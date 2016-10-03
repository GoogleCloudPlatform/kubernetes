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

package images

import (
	"fmt"
	"runtime"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/api"
	"k8s.io/kubernetes/pkg/kubelet/dockertools"

	dockertypes "github.com/docker/engine-api/types"
)

const (
	KubeEtcdImage = "etcd"

	KubeAPIServerImage         = "apiserver"
	KubeControllerManagerImage = "controller-manager"
	KubeSchedulerImage         = "scheduler"
	KubeProxyImage             = "proxy"

	KubeDNSImage         = "kube-dns"
	KubeDNSmasqImage     = "dnsmasq"
	KubeExechealthzImage = "exechealthz"

	gcrPrefix   = "gcr.io/google_containers"
	etcdVersion = "2.2.5"

	kubeDNSVersion     = "1.7"
	dnsmasqVersion     = "1.3"
	exechealthzVersion = "1.1"
)

func PrePullImages(s *kubeadmapi.KubeadmConfig, dockerClient dockertools.DockerInterface) error {
	opts := dockertypes.ImagePullOptions{}
	images := []string{
		GetCoreImage(KubeAPIServerImage, s, s.EnvParams["hyperkube_image"]),
		GetCoreImage(KubeControllerManagerImage, s, s.EnvParams["hyperkube_image"]),
		GetCoreImage(KubeSchedulerImage, s, s.EnvParams["hyperkube_image"]),
		s.EnvParams["discovery_image"],
	}
	if len(s.InitFlags.API.Etcd.ExternalEndpoints) == 0 {
		images = append(images, GetCoreImage(KubeEtcdImage, s, s.EnvParams["etcd_image"]))
	}

	for _, img := range images {
		fmt.Printf("<images/images> Pulling image: %s\n", img)
		if err := dockerClient.PullImage(img, dockertypes.AuthConfig{}, opts); err != nil {
			return err
		}
	}

	return nil
}

func GetCoreImage(image string, cfg *kubeadmapi.KubeadmConfig, overrideImage string) string {
	if overrideImage != "" {
		return overrideImage
	}

	return map[string]string{
		KubeEtcdImage:              fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "etcd", runtime.GOARCH, etcdVersion),
		KubeAPIServerImage:         fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "kube-apiserver", runtime.GOARCH, cfg.Versions.Kubernetes),
		KubeControllerManagerImage: fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "kube-controller-manager", runtime.GOARCH, cfg.Versions.Kubernetes),
		KubeSchedulerImage:         fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "kube-scheduler", runtime.GOARCH, cfg.Versions.Kubernetes),
		KubeProxyImage:             fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "kube-proxy", runtime.GOARCH, cfg.Versions.Kubernetes),
	}[image]
}

func GetAddonImage(image string) string {
	return map[string]string{
		KubeDNSImage:         fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "kubedns", runtime.GOARCH, kubeDNSVersion),
		KubeDNSmasqImage:     fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "kube-dnsmasq", runtime.GOARCH, dnsmasqVersion),
		KubeExechealthzImage: fmt.Sprintf("%s/%s-%s:%s", gcrPrefix, "exechealthz", runtime.GOARCH, exechealthzVersion),
	}[image]
}
