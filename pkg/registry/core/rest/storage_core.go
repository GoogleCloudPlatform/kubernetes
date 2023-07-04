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

package rest

import (
	"crypto/tls"
	goerrors "errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	networkingv1alpha1client "k8s.io/client-go/kubernetes/typed/networking/v1alpha1"
	policyclient "k8s.io/client-go/kubernetes/typed/policy/v1"
	restclient "k8s.io/client-go/rest"

	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/cluster/ports"
	"k8s.io/kubernetes/pkg/features"
	kubeletclient "k8s.io/kubernetes/pkg/kubelet/client"
	"k8s.io/kubernetes/pkg/registry/core/componentstatus"
	configmapstore "k8s.io/kubernetes/pkg/registry/core/configmap/storage"
	endpointsstore "k8s.io/kubernetes/pkg/registry/core/endpoint/storage"
	eventstore "k8s.io/kubernetes/pkg/registry/core/event/storage"
	limitrangestore "k8s.io/kubernetes/pkg/registry/core/limitrange/storage"
	namespacestore "k8s.io/kubernetes/pkg/registry/core/namespace/storage"
	nodestore "k8s.io/kubernetes/pkg/registry/core/node/storage"
	pvstore "k8s.io/kubernetes/pkg/registry/core/persistentvolume/storage"
	pvcstore "k8s.io/kubernetes/pkg/registry/core/persistentvolumeclaim/storage"
	podstore "k8s.io/kubernetes/pkg/registry/core/pod/storage"
	podtemplatestore "k8s.io/kubernetes/pkg/registry/core/podtemplate/storage"
	"k8s.io/kubernetes/pkg/registry/core/rangeallocation"
	controllerstore "k8s.io/kubernetes/pkg/registry/core/replicationcontroller/storage"
	resourcequotastore "k8s.io/kubernetes/pkg/registry/core/resourcequota/storage"
	secretstore "k8s.io/kubernetes/pkg/registry/core/secret/storage"
	"k8s.io/kubernetes/pkg/registry/core/service/allocator"
	serviceallocator "k8s.io/kubernetes/pkg/registry/core/service/allocator/storage"
	"k8s.io/kubernetes/pkg/registry/core/service/ipallocator"
	serviceipallocatorcontroller "k8s.io/kubernetes/pkg/registry/core/service/ipallocator/controller"
	"k8s.io/kubernetes/pkg/registry/core/service/portallocator"
	portallocatorcontroller "k8s.io/kubernetes/pkg/registry/core/service/portallocator/controller"
	servicestore "k8s.io/kubernetes/pkg/registry/core/service/storage"
	serviceaccountstore "k8s.io/kubernetes/pkg/registry/core/serviceaccount/storage"
	kubeschedulerconfig "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/serviceaccount"
	"k8s.io/kubernetes/pkg/util/async"
)

// GenericConfig provides information needed to build RESTStorage
// for generic resources in core. It implements the "normal" RESTStorageProvider interface.
type GenericConfig struct {
	StorageFactory serverstorage.StorageFactory
	// Used for custom proxy dialing, and proxy TLS options
	ProxyTransport http.RoundTripper
	EventTTL       time.Duration

	ServiceAccountIssuer        serviceaccount.TokenGenerator
	ServiceAccountMaxExpiration time.Duration
	ExtendExpiration            bool

	APIAudiences authenticator.Audiences

	LoopbackClientConfig *restclient.Config
	Informers            informers.SharedInformerFactory
}

// Config provides information needed to build RESTStorage for core.
type Config struct {
	GenericConfig

	KubeletClientConfig kubeletclient.KubeletClientConfig

	// Service IP ranges
	ServiceIPRange          net.IPNet
	ServiceSecondaryIPRange net.IPNet
	ServiceNodePortRange    utilnet.PortRange

	ServiceIPRepairInterval time.Duration
}

type serviceIPRangeRegistries struct {
	clusterIP          rangeallocation.RangeRegistry
	secondaryClusterIP rangeallocation.RangeRegistry
	nodePort           rangeallocation.RangeRegistry
}

type legacyProvider struct {
	Config

	primaryServiceClusterIPAllocator ipallocator.Interface
	serviceClusterIPAllocators       map[api.IPFamily]ipallocator.Interface
	serviceNodePortAllocator         *portallocator.PortAllocator

	startServiceNodePortsRepair, startServiceClusterIPRepair func(onFirstSuccess func(), stopCh chan struct{})
}

func New(c Config) (*legacyProvider, error) {
	rangeRegistries, serviceClusterIPAllocator, serviceIPAllocators, serviceNodePortAllocator, err := c.newServiceIPAllocators()
	if err != nil {
		return nil, err
	}

	p := &legacyProvider{
		Config: c,

		primaryServiceClusterIPAllocator: serviceClusterIPAllocator,
		serviceClusterIPAllocators:       serviceIPAllocators,
		serviceNodePortAllocator:         serviceNodePortAllocator,
	}

	// create service node port repair controller
	client, err := kubernetes.NewForConfig(c.LoopbackClientConfig)
	if err != nil {
		return nil, err
	}
	p.startServiceNodePortsRepair = portallocatorcontroller.NewRepair(c.ServiceIPRepairInterval, client.CoreV1(), client.EventsV1(), c.ServiceNodePortRange, rangeRegistries.nodePort).RunUntil

	// create service cluster ip repair controller
	if !utilfeature.DefaultFeatureGate.Enabled(features.MultiCIDRServiceAllocator) {
		p.startServiceClusterIPRepair = serviceipallocatorcontroller.NewRepair(
			c.ServiceIPRepairInterval,
			client.CoreV1(),
			client.EventsV1(),
			&c.ServiceIPRange,
			rangeRegistries.clusterIP,
			&c.ServiceSecondaryIPRange,
			rangeRegistries.secondaryClusterIP,
		).RunUntil
	} else {
		p.startServiceClusterIPRepair = serviceipallocatorcontroller.NewRepairIPAddress(
			c.ServiceIPRepairInterval,
			client,
			&c.ServiceIPRange,
			&c.ServiceSecondaryIPRange,
			c.Informers.Core().V1().Services(),
			c.Informers.Networking().V1alpha1().IPAddresses(),
		).RunUntil
	}

	return p, nil
}

func (c *GenericConfig) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.APIGroupInfo{
		PrioritizedVersions:          legacyscheme.Scheme.PrioritizedVersionsForGroup(""),
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{},
		Scheme:                       legacyscheme.Scheme,
		ParameterCodec:               legacyscheme.ParameterCodec,
		NegotiatedSerializer:         legacyscheme.Codecs,
	}

	eventStorage, err := eventstore.NewREST(restOptionsGetter, uint64(c.EventTTL.Seconds()))
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	resourceQuotaStorage, resourceQuotaStatusStorage, err := resourcequotastore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}
	secretStorage, err := secretstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	configMapStorage, err := configmapstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	namespaceStorage, namespaceStatusStorage, namespaceFinalizeStorage, err := namespacestore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	var serviceAccountStorage *serviceaccountstore.REST
	if c.ServiceAccountIssuer != nil {
		serviceAccountStorage, err = serviceaccountstore.NewREST(restOptionsGetter, c.ServiceAccountIssuer, c.APIAudiences, c.ServiceAccountMaxExpiration, nil, secretStorage.Store, c.ExtendExpiration)
	} else {
		serviceAccountStorage, err = serviceaccountstore.NewREST(restOptionsGetter, nil, nil, 0, nil, nil, false)
	}
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	storage := map[string]rest.Storage{}
	if resource := "events"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = eventStorage
	}

	if resource := "resourcequotas"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = resourceQuotaStorage
		storage[resource+"/status"] = resourceQuotaStatusStorage
	}

	if resource := "namespaces"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = namespaceStorage
		storage[resource+"/status"] = namespaceStatusStorage
		storage[resource+"/finalize"] = namespaceFinalizeStorage
	}

	if resource := "secrets"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = secretStorage
	}

	if resource := "serviceaccounts"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = serviceAccountStorage
		if serviceAccountStorage.Token != nil {
			storage[resource+"/token"] = serviceAccountStorage.Token
		}
	}

	if resource := "configmaps"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = configMapStorage
	}

	if len(storage) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap["v1"] = storage
	}

	return apiGroupInfo, nil
}

func (p *GenericConfig) GroupName() string {
	return api.GroupName
}

func (c *legacyProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo, err := c.GenericConfig.NewRESTStorage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	podDisruptionClient, err := policyclient.NewForConfig(c.LoopbackClientConfig)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	podTemplateStorage, err := podtemplatestore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	limitRangeStorage, err := limitrangestore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	persistentVolumeStorage, persistentVolumeStatusStorage, err := pvstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}
	persistentVolumeClaimStorage, persistentVolumeClaimStatusStorage, err := pvcstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	endpointsStorage, err := endpointsstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	nodeStorage, err := nodestore.NewStorage(restOptionsGetter, c.KubeletClientConfig, c.ProxyTransport)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	podStorage, err := podstore.NewStorage(
		restOptionsGetter,
		nodeStorage.KubeletConnectionInfo,
		c.ProxyTransport,
		podDisruptionClient,
	)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	serviceRESTStorage, serviceStatusStorage, serviceRESTProxy, err := servicestore.NewREST(
		restOptionsGetter,
		c.primaryServiceClusterIPAllocator.IPFamily(),
		c.serviceClusterIPAllocators,
		c.serviceNodePortAllocator,
		endpointsStorage,
		podStorage.Pod,
		c.ProxyTransport)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	storage := apiGroupInfo.VersionedResourcesStorageMap["v1"]
	if storage == nil {
		storage = map[string]rest.Storage{}
	}

	// potentially override the generic serviceaccount storage with one that supports pods
	var serviceAccountStorage *serviceaccountstore.REST
	if c.ServiceAccountIssuer != nil {
		serviceAccountStorage, err = serviceaccountstore.NewREST(restOptionsGetter, c.ServiceAccountIssuer, c.APIAudiences, c.ServiceAccountMaxExpiration, podStorage.Pod.Store, storage["secrets"].(rest.Getter), c.ExtendExpiration)
		if err != nil {
			return genericapiserver.APIGroupInfo{}, err
		}
	}

	if resource := "pods"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = podStorage.Pod
		storage[resource+"/attach"] = podStorage.Attach
		storage[resource+"/status"] = podStorage.Status
		storage[resource+"/log"] = podStorage.Log
		storage[resource+"/exec"] = podStorage.Exec
		storage[resource+"/portforward"] = podStorage.PortForward
		storage[resource+"/proxy"] = podStorage.Proxy
		storage[resource+"/binding"] = podStorage.Binding
		if podStorage.Eviction != nil {
			storage[resource+"/eviction"] = podStorage.Eviction
		}
		storage[resource+"/ephemeralcontainers"] = podStorage.EphemeralContainers
	}
	if resource := "bindings"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = podStorage.LegacyBinding
	}

	if resource := "podtemplates"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = podTemplateStorage
	}

	if resource := "replicationcontrollers"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		controllerStorage, err := controllerstore.NewStorage(restOptionsGetter)
		if err != nil {
			return genericapiserver.APIGroupInfo{}, err
		}

		storage[resource] = controllerStorage.Controller
		storage[resource+"/status"] = controllerStorage.Status
		if legacyscheme.Scheme.IsVersionRegistered(schema.GroupVersion{Group: "autoscaling", Version: "v1"}) {
			storage[resource+"/scale"] = controllerStorage.Scale
		}
	}

	// potentially override generic storage for service account (with pod support)
	if resource := "serviceaccounts"; serviceAccountStorage != nil && apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		// don't leak go routines
		storage[resource].Destroy()
		if storage[resource+"/token"] != nil {
			storage[resource+"/token"].Destroy()
		}

		storage[resource] = serviceAccountStorage
		if serviceAccountStorage.Token != nil {
			storage[resource+"/token"] = serviceAccountStorage.Token
		}
	}

	if resource := "services"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = serviceRESTStorage
		storage[resource+"/proxy"] = serviceRESTProxy
		storage[resource+"/status"] = serviceStatusStorage
	}

	if resource := "endpoints"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = endpointsStorage
	}

	if resource := "nodes"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = nodeStorage.Node
		storage[resource+"/proxy"] = nodeStorage.Proxy
		storage[resource+"/status"] = nodeStorage.Status
	}

	if resource := "limitranges"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = limitRangeStorage
	}

	if resource := "persistentvolumes"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = persistentVolumeStorage
		storage[resource+"/status"] = persistentVolumeStatusStorage
	}

	if resource := "persistentvolumeclaims"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = persistentVolumeClaimStorage
		storage[resource+"/status"] = persistentVolumeClaimStatusStorage
	}

	if resource := "componentstatuses"; apiResourceConfigSource.ResourceEnabled(corev1.SchemeGroupVersion.WithResource(resource)) {
		storage[resource] = componentstatus.NewStorage(componentStatusStorage{c.StorageFactory}.serversToValidate)
	}

	if len(storage) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap["v1"] = storage
	}

	return apiGroupInfo, nil
}

func (c Config) newServiceIPAllocators() (registries serviceIPRangeRegistries, serviceClusterIPAllocator ipallocator.Interface, clusterIPAllocators map[api.IPFamily]ipallocator.Interface, serviceNodePortAllocator *portallocator.PortAllocator, err error) {
	serviceStorageConfig, err := c.StorageFactory.NewConfig(api.Resource("services"))
	if err != nil {
		return serviceIPRangeRegistries{}, nil, nil, nil, err
	}

	serviceClusterIPRange := c.ServiceIPRange
	if serviceClusterIPRange.IP == nil {
		return serviceIPRangeRegistries{}, nil, nil, nil, fmt.Errorf("service clusterIPRange is missing")
	}

	if !utilfeature.DefaultFeatureGate.Enabled(features.MultiCIDRServiceAllocator) {
		serviceClusterIPAllocator, err = ipallocator.New(&serviceClusterIPRange, func(max int, rangeSpec string, offset int) (allocator.Interface, error) {
			var mem allocator.Snapshottable
			mem = allocator.NewAllocationMapWithOffset(max, rangeSpec, offset)
			// TODO etcdallocator package to return a storage interface via the storageFactory
			etcd, err := serviceallocator.NewEtcd(mem, "/ranges/serviceips", serviceStorageConfig.ForResource(api.Resource("serviceipallocations")))
			if err != nil {
				return nil, err
			}
			registries.clusterIP = etcd
			return etcd, nil
		})
		if err != nil {
			return serviceIPRangeRegistries{}, nil, nil, nil, fmt.Errorf("cannot create cluster IP allocator: %v", err)
		}
	} else {
		networkingv1alphaClient, err := networkingv1alpha1client.NewForConfig(c.LoopbackClientConfig)
		if err != nil {
			return serviceIPRangeRegistries{}, nil, nil, nil, err
		}
		serviceClusterIPAllocator, err = ipallocator.NewIPAllocator(&serviceClusterIPRange, networkingv1alphaClient, c.Informers.Networking().V1alpha1().IPAddresses())
		if err != nil {
			return serviceIPRangeRegistries{}, nil, nil, nil, fmt.Errorf("cannot create cluster IP allocator: %v", err)
		}
	}
	serviceClusterIPAllocator.EnableMetrics()

	var secondaryServiceClusterIPAllocator ipallocator.Interface
	if c.ServiceSecondaryIPRange.IP != nil {
		if !utilfeature.DefaultFeatureGate.Enabled(features.MultiCIDRServiceAllocator) {
			var err error
			secondaryServiceClusterIPAllocator, err = ipallocator.New(&c.ServiceSecondaryIPRange, func(max int, rangeSpec string, offset int) (allocator.Interface, error) {
				var mem allocator.Snapshottable
				mem = allocator.NewAllocationMapWithOffset(max, rangeSpec, offset)
				// TODO etcdallocator package to return a storage interface via the storageFactory
				etcd, err := serviceallocator.NewEtcd(mem, "/ranges/secondaryserviceips", serviceStorageConfig.ForResource(api.Resource("serviceipallocations")))
				if err != nil {
					return nil, err
				}
				registries.secondaryClusterIP = etcd
				return etcd, nil
			})
			if err != nil {
				return serviceIPRangeRegistries{}, nil, nil, nil, fmt.Errorf("cannot create cluster secondary IP allocator: %v", err)
			}
		} else {
			networkingv1alphaClient, err := networkingv1alpha1client.NewForConfig(c.LoopbackClientConfig)
			if err != nil {
				return serviceIPRangeRegistries{}, nil, nil, nil, err
			}
			secondaryServiceClusterIPAllocator, err = ipallocator.NewIPAllocator(&c.ServiceSecondaryIPRange, networkingv1alphaClient, c.Informers.Networking().V1alpha1().IPAddresses())
			if err != nil {
				return serviceIPRangeRegistries{}, nil, nil, nil, fmt.Errorf("cannot create cluster secondary IP allocator: %v", err)
			}
		}
		secondaryServiceClusterIPAllocator.EnableMetrics()
	}

	serviceNodePortAllocator, err = portallocator.New(c.ServiceNodePortRange, func(max int, rangeSpec string, offset int) (allocator.Interface, error) {
		mem := allocator.NewAllocationMapWithOffset(max, rangeSpec, offset)
		// TODO etcdallocator package to return a storage interface via the storageFactory
		etcd, err := serviceallocator.NewEtcd(mem, "/ranges/servicenodeports", serviceStorageConfig.ForResource(api.Resource("servicenodeportallocations")))
		if err != nil {
			return nil, err
		}
		registries.nodePort = etcd
		return etcd, nil
	})
	if err != nil {
		return serviceIPRangeRegistries{}, nil, nil, nil, fmt.Errorf("cannot create cluster port allocator: %v", err)
	}
	serviceNodePortAllocator.EnableMetrics()

	clusterIPAllocators = map[api.IPFamily]ipallocator.Interface{
		serviceClusterIPAllocator.IPFamily(): serviceClusterIPAllocator,
	}
	if secondaryServiceClusterIPAllocator != nil {
		clusterIPAllocators[secondaryServiceClusterIPAllocator.IPFamily()] = secondaryServiceClusterIPAllocator
	}

	return
}

var _ genericapiserver.PostStartHookProvider = &legacyProvider{}

func (p *legacyProvider) PostStartHook() (string, genericapiserver.PostStartHookFunc, error) {
	return "start-service-storage-controllers", func(context genericapiserver.PostStartHookContext) error {
		// We start both repairClusterIPs and repairNodePorts to ensure repair
		// loops of ClusterIPs and NodePorts.
		// We run both repair loops using RunUntil public interface.
		// However, we want to fail liveness/readiness until the first
		// successful repair loop, so we basically pass appropriate
		// callbacks to RunUtil methods.
		// Additionally, we ensure that we don't wait for it for longer
		// than 1 minute for backward compatibility of failing the whole
		// apiserver if we can't repair them.
		wg := sync.WaitGroup{}
		wg.Add(2)
		runner := async.NewRunner(
			func(stopCh chan struct{}) { p.startServiceClusterIPRepair(wg.Done, stopCh) },
			func(stopCh chan struct{}) { p.startServiceNodePortsRepair(wg.Done, stopCh) },
		)
		runner.Start()
		go func() {
			defer runner.Stop()
			<-context.StopCh
		}()

		// For backward compatibility, we ensure that if we never are able
		// to repair clusterIPs and/or nodeports, we not only fail the liveness
		// and/or readiness, but also explicitly fail.
		done := make(chan struct{})
		go func() {
			defer close(done)
			wg.Wait()
		}()
		select {
		case <-done:
		case <-time.After(time.Minute):
			return goerrors.New("unable to perform initial IP and Port allocation check")
		}

		return nil
	}, nil
}

func (p *legacyProvider) GroupName() string {
	return api.GroupName
}

type componentStatusStorage struct {
	storageFactory serverstorage.StorageFactory
}

func (s componentStatusStorage) serversToValidate() map[string]componentstatus.Server {
	// this is fragile, which assumes that the default port is being used
	// TODO: switch to secure port until these components remove the ability to serve insecurely.
	serversToValidate := map[string]componentstatus.Server{
		"controller-manager": &componentstatus.HttpServer{EnableHTTPS: true, TLSConfig: &tls.Config{InsecureSkipVerify: true}, Addr: "127.0.0.1", Port: ports.KubeControllerManagerPort, Path: "/healthz"},
		"scheduler":          &componentstatus.HttpServer{EnableHTTPS: true, TLSConfig: &tls.Config{InsecureSkipVerify: true}, Addr: "127.0.0.1", Port: kubeschedulerconfig.DefaultKubeSchedulerPort, Path: "/healthz"},
	}

	for ix, cfg := range s.storageFactory.Configs() {
		serversToValidate[fmt.Sprintf("etcd-%d", ix)] = &componentstatus.EtcdServer{Config: cfg}
	}
	return serversToValidate
}
