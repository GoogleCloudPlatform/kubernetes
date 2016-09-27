// +build cgo,linux

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

package cadvisor

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"crypto/tls"
	"crypto/x509"
	"github.com/golang/glog"
	"github.com/google/cadvisor/cache/memory"
	cadvisorMetrics "github.com/google/cadvisor/container"
	"github.com/google/cadvisor/events"
	cadvisorfs "github.com/google/cadvisor/fs"
	cadvisorhttp "github.com/google/cadvisor/http"
	cadvisorapi "github.com/google/cadvisor/info/v1"
	cadvisorapiv2 "github.com/google/cadvisor/info/v2"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/metrics"
	"github.com/google/cadvisor/utils/sysfs"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/kubelet/types"
	"k8s.io/kubernetes/pkg/util/runtime"
)

type cadvisorClient struct {
	runtime string
	manager.Manager
}

var _ Interface = new(cadvisorClient)

// TODO(vmarmol): Make configurable.
// The amount of time for which to keep stats in memory.
const statsCacheDuration = 2 * time.Minute
const maxHousekeepingInterval = 15 * time.Second
const defaultHousekeepingInterval = 10 * time.Second
const allowDynamicHousekeeping = true

func init() {
	// Override cAdvisor flag defaults.
	flagOverrides := map[string]string{
		// Override the default cAdvisor housekeeping interval.
		"housekeeping_interval": defaultHousekeepingInterval.String(),
		// Disable event storage by default.
		"event_storage_event_limit": "default=0",
		"event_storage_age_limit":   "default=0",
	}
	for name, defaultValue := range flagOverrides {
		if f := flag.Lookup(name); f != nil {
			f.DefValue = defaultValue
			f.Value.Set(defaultValue)
		} else {
			glog.Errorf("Expected cAdvisor flag %q not found", name)
		}
	}
}

func containerLabels(c *cadvisorapi.ContainerInfo) map[string]string {
	set := map[string]string{metrics.LabelID: c.Name}
	if len(c.Aliases) > 0 {
		set[metrics.LabelName] = c.Aliases[0]
	}
	if image := c.Spec.Image; len(image) > 0 {
		set[metrics.LabelImage] = image
	}
	if v, ok := c.Spec.Labels[types.KubernetesPodNameLabel]; ok {
		set["pod_name"] = v
	}
	if v, ok := c.Spec.Labels[types.KubernetesPodNamespaceLabel]; ok {
		set["namespace"] = v
	}
	if v, ok := c.Spec.Labels[types.KubernetesContainerNameLabel]; ok {
		set["container_name"] = v
	}
	return set
}

type CAdvisorCustomMetricsConfig struct {
	CollectorClientCertFile       string
	CollectorClientPrivateKeyFile string
	InsecureSkipVerify            bool
	RootCAFile                    string
}

// New creates a cAdvisor and exports its API on the specified port if port > 0.
func New(port uint, runtime string, config CAdvisorCustomMetricsConfig) (Interface, error) {
	sysFs, err := sysfs.NewRealSysFs()
	if err != nil {
		return nil, err
	}

	//generate the http.Client to be used by the cAdvisor custom metric providers
	httpClient, err := generateCollectorHttpClient(config)
	if err != nil {
		return nil, err
	}

	// Create and start the cAdvisor container manager.
	m, err := manager.New(memory.New(statsCacheDuration, nil), sysFs, maxHousekeepingInterval, allowDynamicHousekeeping, cadvisorMetrics.MetricSet{cadvisorMetrics.NetworkTcpUsageMetrics: struct{}{}}, httpClient)
	if err != nil {
		return nil, err
	}

	cadvisorClient := &cadvisorClient{
		runtime: runtime,
		Manager: m,
	}

	err = cadvisorClient.exportHTTP(port)
	if err != nil {
		return nil, err
	}
	return cadvisorClient, nil
}

func generateCollectorHttpClient(config CAdvisorCustomMetricsConfig) (*http.Client, error) {
	tlsConfig := &tls.Config{}
	// if configured, set insecureSkipVerify when connecting to insecure tls metric endpoints
	if config.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	// if the customMetricsRootCAFile is empty we just use the system defaults.
	// if it is specified, we use the certificates contained within this file
	if config.RootCAFile != "" {
		rootCAs := x509.NewCertPool()

		pem, err := ioutil.ReadFile(config.RootCAFile)
		if err != nil {
			return nil, fmt.Errorf("Error trying to load the custom metrics root CA file: %v", err)
		}

		rootCAs.AppendCertsFromPEM(pem)

		tlsConfig.RootCAs = rootCAs
	}

	if config.CollectorClientCertFile != "" {
		if config.CollectorClientPrivateKeyFile == "" {
			return nil, fmt.Errorf("invalid configuration: customMetricsCollectorClientCertFile was specified and customMetricsCollectorClientPrivateKeyFile was not specified.")
		}
		cert, err := tls.LoadX509KeyPair(config.CollectorClientCertFile, config.CollectorClientPrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading the cAdvisor client certificates: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	httpClient := http.Client{Transport: transport}

	return &httpClient, nil
}

func (cc *cadvisorClient) Start() error {
	return cc.Manager.Start()
}

func (cc *cadvisorClient) exportHTTP(port uint) error {
	// Register the handlers regardless as this registers the prometheus
	// collector properly.
	mux := http.NewServeMux()
	err := cadvisorhttp.RegisterHandlers(mux, cc, "", "", "", "")
	if err != nil {
		return err
	}

	cadvisorhttp.RegisterPrometheusHandler(mux, cc, "/metrics", containerLabels)

	// Only start the http server if port > 0
	if port > 0 {
		serv := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}

		// TODO(vmarmol): Remove this when the cAdvisor port is once again free.
		// If export failed, retry in the background until we are able to bind.
		// This allows an existing cAdvisor to be killed before this one registers.
		go func() {
			defer runtime.HandleCrash()

			err := serv.ListenAndServe()
			for err != nil {
				glog.Infof("Failed to register cAdvisor on port %d, retrying. Error: %v", port, err)
				time.Sleep(time.Minute)
				err = serv.ListenAndServe()
			}
		}()
	}

	return nil
}

func (cc *cadvisorClient) ContainerInfo(name string, req *cadvisorapi.ContainerInfoRequest) (*cadvisorapi.ContainerInfo, error) {
	return cc.GetContainerInfo(name, req)
}

func (cc *cadvisorClient) ContainerInfoV2(name string, options cadvisorapiv2.RequestOptions) (map[string]cadvisorapiv2.ContainerInfo, error) {
	return cc.GetContainerInfoV2(name, options)
}

func (cc *cadvisorClient) VersionInfo() (*cadvisorapi.VersionInfo, error) {
	return cc.GetVersionInfo()
}

func (cc *cadvisorClient) SubcontainerInfo(name string, req *cadvisorapi.ContainerInfoRequest) (map[string]*cadvisorapi.ContainerInfo, error) {
	infos, err := cc.SubcontainersInfo(name, req)
	if err != nil && len(infos) == 0 {
		return nil, err
	}

	result := make(map[string]*cadvisorapi.ContainerInfo, len(infos))
	for _, info := range infos {
		result[info.Name] = info
	}
	return result, err
}

func (cc *cadvisorClient) MachineInfo() (*cadvisorapi.MachineInfo, error) {
	return cc.GetMachineInfo()
}

func (cc *cadvisorClient) ImagesFsInfo() (cadvisorapiv2.FsInfo, error) {
	var label string

	switch cc.runtime {
	case "docker":
		label = cadvisorfs.LabelDockerImages
	case "rkt":
		label = cadvisorfs.LabelRktImages
	default:
		return cadvisorapiv2.FsInfo{}, fmt.Errorf("ImagesFsInfo: unknown runtime: %v", cc.runtime)
	}

	return cc.getFsInfo(label)
}

func (cc *cadvisorClient) RootFsInfo() (cadvisorapiv2.FsInfo, error) {
	return cc.getFsInfo(cadvisorfs.LabelSystemRoot)
}

func (cc *cadvisorClient) getFsInfo(label string) (cadvisorapiv2.FsInfo, error) {
	res, err := cc.GetFsInfo(label)
	if err != nil {
		return cadvisorapiv2.FsInfo{}, err
	}
	if len(res) == 0 {
		return cadvisorapiv2.FsInfo{}, fmt.Errorf("failed to find information for the filesystem labeled %q", label)
	}
	// TODO(vmarmol): Handle this better when a label has more than one image filesystem.
	if len(res) > 1 {
		glog.Warningf("More than one filesystem labeled %q: %#v. Only using the first one", label, res)
	}

	return res[0], nil
}

func (cc *cadvisorClient) WatchEvents(request *events.Request) (*events.EventChannel, error) {
	return cc.WatchForEvents(request)
}
