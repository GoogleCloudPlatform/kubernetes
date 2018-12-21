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

// Code generated by client-gen. DO NOT EDIT.

package internalclientset

import (
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
	admissionregistrationinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/admissionregistration/internalversion"
	appsinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/apps/internalversion"
	auditregistrationinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/auditregistration/internalversion"
	authenticationinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/authentication/internalversion"
	authorizationinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/authorization/internalversion"
	autoscalinginternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/autoscaling/internalversion"
	batchinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/batch/internalversion"
	certificatesinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/certificates/internalversion"
	coordinationinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/coordination/internalversion"
	coreinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	eventsinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/events/internalversion"
	extensionsinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/extensions/internalversion"
	networkinginternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/networking/internalversion"
	policyinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/policy/internalversion"
	rbacinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/rbac/internalversion"
	schedulinginternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/scheduling/internalversion"
	settingsinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/settings/internalversion"
	storageinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/storage/internalversion"
	topologyinternalversion "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/topology/internalversion"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	Admissionregistration() admissionregistrationinternalversion.AdmissionregistrationInterface
	Core() coreinternalversion.CoreInterface
	Apps() appsinternalversion.AppsInterface
	Auditregistration() auditregistrationinternalversion.AuditregistrationInterface
	Authentication() authenticationinternalversion.AuthenticationInterface
	Authorization() authorizationinternalversion.AuthorizationInterface
	Autoscaling() autoscalinginternalversion.AutoscalingInterface
	Batch() batchinternalversion.BatchInterface
	Certificates() certificatesinternalversion.CertificatesInterface
	Coordination() coordinationinternalversion.CoordinationInterface
	Events() eventsinternalversion.EventsInterface
	Extensions() extensionsinternalversion.ExtensionsInterface
	Networking() networkinginternalversion.NetworkingInterface
	Policy() policyinternalversion.PolicyInterface
	Rbac() rbacinternalversion.RbacInterface
	Scheduling() schedulinginternalversion.SchedulingInterface
	Settings() settingsinternalversion.SettingsInterface
	Storage() storageinternalversion.StorageInterface
	Topology() topologyinternalversion.TopologyInterface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	admissionregistration *admissionregistrationinternalversion.AdmissionregistrationClient
	core                  *coreinternalversion.CoreClient
	apps                  *appsinternalversion.AppsClient
	auditregistration     *auditregistrationinternalversion.AuditregistrationClient
	authentication        *authenticationinternalversion.AuthenticationClient
	authorization         *authorizationinternalversion.AuthorizationClient
	autoscaling           *autoscalinginternalversion.AutoscalingClient
	batch                 *batchinternalversion.BatchClient
	certificates          *certificatesinternalversion.CertificatesClient
	coordination          *coordinationinternalversion.CoordinationClient
	events                *eventsinternalversion.EventsClient
	extensions            *extensionsinternalversion.ExtensionsClient
	networking            *networkinginternalversion.NetworkingClient
	policy                *policyinternalversion.PolicyClient
	rbac                  *rbacinternalversion.RbacClient
	scheduling            *schedulinginternalversion.SchedulingClient
	settings              *settingsinternalversion.SettingsClient
	storage               *storageinternalversion.StorageClient
	topology              *topologyinternalversion.TopologyClient
}

// Admissionregistration retrieves the AdmissionregistrationClient
func (c *Clientset) Admissionregistration() admissionregistrationinternalversion.AdmissionregistrationInterface {
	return c.admissionregistration
}

// Core retrieves the CoreClient
func (c *Clientset) Core() coreinternalversion.CoreInterface {
	return c.core
}

// Apps retrieves the AppsClient
func (c *Clientset) Apps() appsinternalversion.AppsInterface {
	return c.apps
}

// Auditregistration retrieves the AuditregistrationClient
func (c *Clientset) Auditregistration() auditregistrationinternalversion.AuditregistrationInterface {
	return c.auditregistration
}

// Authentication retrieves the AuthenticationClient
func (c *Clientset) Authentication() authenticationinternalversion.AuthenticationInterface {
	return c.authentication
}

// Authorization retrieves the AuthorizationClient
func (c *Clientset) Authorization() authorizationinternalversion.AuthorizationInterface {
	return c.authorization
}

// Autoscaling retrieves the AutoscalingClient
func (c *Clientset) Autoscaling() autoscalinginternalversion.AutoscalingInterface {
	return c.autoscaling
}

// Batch retrieves the BatchClient
func (c *Clientset) Batch() batchinternalversion.BatchInterface {
	return c.batch
}

// Certificates retrieves the CertificatesClient
func (c *Clientset) Certificates() certificatesinternalversion.CertificatesInterface {
	return c.certificates
}

// Coordination retrieves the CoordinationClient
func (c *Clientset) Coordination() coordinationinternalversion.CoordinationInterface {
	return c.coordination
}

// Events retrieves the EventsClient
func (c *Clientset) Events() eventsinternalversion.EventsInterface {
	return c.events
}

// Extensions retrieves the ExtensionsClient
func (c *Clientset) Extensions() extensionsinternalversion.ExtensionsInterface {
	return c.extensions
}

// Networking retrieves the NetworkingClient
func (c *Clientset) Networking() networkinginternalversion.NetworkingInterface {
	return c.networking
}

// Policy retrieves the PolicyClient
func (c *Clientset) Policy() policyinternalversion.PolicyInterface {
	return c.policy
}

// Rbac retrieves the RbacClient
func (c *Clientset) Rbac() rbacinternalversion.RbacInterface {
	return c.rbac
}

// Scheduling retrieves the SchedulingClient
func (c *Clientset) Scheduling() schedulinginternalversion.SchedulingInterface {
	return c.scheduling
}

// Settings retrieves the SettingsClient
func (c *Clientset) Settings() settingsinternalversion.SettingsInterface {
	return c.settings
}

// Storage retrieves the StorageClient
func (c *Clientset) Storage() storageinternalversion.StorageInterface {
	return c.storage
}

// Topology retrieves the TopologyClient
func (c *Clientset) Topology() topologyinternalversion.TopologyInterface {
	return c.topology
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.admissionregistration, err = admissionregistrationinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.core, err = coreinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.apps, err = appsinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.auditregistration, err = auditregistrationinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.authentication, err = authenticationinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.authorization, err = authorizationinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.autoscaling, err = autoscalinginternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.batch, err = batchinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.certificates, err = certificatesinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.coordination, err = coordinationinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.events, err = eventsinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.extensions, err = extensionsinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.networking, err = networkinginternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.policy, err = policyinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.rbac, err = rbacinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.scheduling, err = schedulinginternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.settings, err = settingsinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.storage, err = storageinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.topology, err = topologyinternalversion.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.admissionregistration = admissionregistrationinternalversion.NewForConfigOrDie(c)
	cs.core = coreinternalversion.NewForConfigOrDie(c)
	cs.apps = appsinternalversion.NewForConfigOrDie(c)
	cs.auditregistration = auditregistrationinternalversion.NewForConfigOrDie(c)
	cs.authentication = authenticationinternalversion.NewForConfigOrDie(c)
	cs.authorization = authorizationinternalversion.NewForConfigOrDie(c)
	cs.autoscaling = autoscalinginternalversion.NewForConfigOrDie(c)
	cs.batch = batchinternalversion.NewForConfigOrDie(c)
	cs.certificates = certificatesinternalversion.NewForConfigOrDie(c)
	cs.coordination = coordinationinternalversion.NewForConfigOrDie(c)
	cs.events = eventsinternalversion.NewForConfigOrDie(c)
	cs.extensions = extensionsinternalversion.NewForConfigOrDie(c)
	cs.networking = networkinginternalversion.NewForConfigOrDie(c)
	cs.policy = policyinternalversion.NewForConfigOrDie(c)
	cs.rbac = rbacinternalversion.NewForConfigOrDie(c)
	cs.scheduling = schedulinginternalversion.NewForConfigOrDie(c)
	cs.settings = settingsinternalversion.NewForConfigOrDie(c)
	cs.storage = storageinternalversion.NewForConfigOrDie(c)
	cs.topology = topologyinternalversion.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.admissionregistration = admissionregistrationinternalversion.New(c)
	cs.core = coreinternalversion.New(c)
	cs.apps = appsinternalversion.New(c)
	cs.auditregistration = auditregistrationinternalversion.New(c)
	cs.authentication = authenticationinternalversion.New(c)
	cs.authorization = authorizationinternalversion.New(c)
	cs.autoscaling = autoscalinginternalversion.New(c)
	cs.batch = batchinternalversion.New(c)
	cs.certificates = certificatesinternalversion.New(c)
	cs.coordination = coordinationinternalversion.New(c)
	cs.events = eventsinternalversion.New(c)
	cs.extensions = extensionsinternalversion.New(c)
	cs.networking = networkinginternalversion.New(c)
	cs.policy = policyinternalversion.New(c)
	cs.rbac = rbacinternalversion.New(c)
	cs.scheduling = schedulinginternalversion.New(c)
	cs.settings = settingsinternalversion.New(c)
	cs.storage = storageinternalversion.New(c)
	cs.topology = topologyinternalversion.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
