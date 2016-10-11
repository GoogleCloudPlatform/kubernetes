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

package core

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/v1"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/quota"
	"k8s.io/kubernetes/pkg/quota/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/util/sets"
)

// NewServiceEvaluator returns an evaluator that can evaluate service quotas
func NewServiceEvaluator(kubeClient clientset.Interface) quota.Evaluator {
	return &serviceEvaluator{
		resources: []api.ResourceName{
			api.ResourceServices,
			api.ResourceServicesNodePorts,
			api.ResourceServicesLoadBalancers,
		},
		listFuncByNamespace: func(namespace string, options v1.ListOptions) ([]runtime.Object, error) {
			itemList, err := kubeClient.Core().Services(namespace).List(options)
			if err != nil {
				return nil, err
			}
			results := make([]runtime.Object, 0, len(itemList.Items))
			for i := range itemList.Items {
				results = append(results, &itemList.Items[i])
			}
			return results, nil
		},
		operations: []admission.Operation{admission.Create, admission.Update},
	}
}

// serviceEvaluator knows how to measure usage for services.
type serviceEvaluator struct {
	// list of resources tracked with services
	resources []api.ResourceName
	// knows how to list items by namespace
	listFuncByNamespace generic.ListFuncByNamespace
	// set of operations tracked by this evaluator
	operations []admission.Operation
}

// Constraints verifies that all required resources are present on the item
func (p *serviceEvaluator) Constraints(required []api.ResourceName, item runtime.Object) error {
	service, ok := item.(*api.Service)
	if !ok {
		return fmt.Errorf("unexpected input object %v", item)
	}

	requiredSet := quota.ToSet(required)
	missingSet := sets.NewString()
	serviceUsage := ServiceUsageFunc(service)
	serviceSet := quota.ToSet(quota.ResourceNames(serviceUsage))
	if diff := requiredSet.Difference(serviceSet); len(diff) > 0 {
		missingSet.Insert(diff.List()...)
	}

	if len(missingSet) == 0 {
		return nil
	}
	return fmt.Errorf("must specify %s", strings.Join(missingSet.List(), ","))
}

// GroupKind that this evaluator tracks
func (p *serviceEvaluator) GroupKind() schema.GroupKind {
	return api.Kind("Service")
}

// Handles returns true of the evalutor should handle the specified operation.
func (p *serviceEvaluator) Handles(operation admission.Operation) bool {
	return generic.Contains(p.operations, operation)
}

// Matches returns true if the evaluator matches the specified quota with the provided input item
func (p *serviceEvaluator) Matches(resourceQuota *api.ResourceQuota, item runtime.Object) bool {
	return generic.Matches(resourceQuota, item, p.MatchingResources, generic.MatchesNoScopeFunc)
}

// MatchingResources takes the input specified list of resources and returns the set of resources it matches.
func (p *serviceEvaluator) MatchingResources(input []api.ResourceName) []api.ResourceName {
	return quota.Intersection(input, p.resources)
}

// Usage knows how to measure usage associated with pods
func (p *serviceEvaluator) Usage(item runtime.Object) api.ResourceList {
	return ServiceUsageFunc(item)
}

// UsageStats calculates aggregate usage for the object.
func (p *serviceEvaluator) UsageStats(options quota.UsageStatsOptions) (quota.UsageStats, error) {
	return generic.CalculateUsageStats(options, p.listFuncByNamespace, generic.MatchesNoScopeFunc, p.Usage)
}

var _ quota.Evaluator = &serviceEvaluator{}

// ServiceUsageFunc knows how to measure usage associated with services
func ServiceUsageFunc(object runtime.Object) api.ResourceList {
	result := api.ResourceList{}
	var serviceType api.ServiceType
	var ports int

	switch t := object.(type) {
	case *v1.Service:
		serviceType = api.ServiceType(t.Spec.Type)
		ports = len(t.Spec.Ports)
	case *api.Service:
		serviceType = t.Spec.Type
		ports = len(t.Spec.Ports)
	default:
		panic(fmt.Sprintf("expect *api.Service or *v1.Service, got %v", t))
	}

	// default service usage
	result[api.ResourceServices] = *(resource.NewQuantity(1, resource.DecimalSI))
	result[api.ResourceServicesLoadBalancers] = resource.Quantity{Format: resource.DecimalSI}
	result[api.ResourceServicesNodePorts] = resource.Quantity{Format: resource.DecimalSI}
	switch serviceType {
	case api.ServiceTypeNodePort:
		// node port services need to count node ports
		value := resource.NewQuantity(int64(ports), resource.DecimalSI)
		result[api.ResourceServicesNodePorts] = *value
	case api.ServiceTypeLoadBalancer:
		// load balancer services need to count load balancers
		result[api.ResourceServicesLoadBalancers] = *(resource.NewQuantity(1, resource.DecimalSI))
	}
	return result
}

// QuotaServiceType returns true if the service type is eligible to track against a quota
func QuotaServiceType(service *v1.Service) bool {
	switch service.Spec.Type {
	case v1.ServiceTypeNodePort, v1.ServiceTypeLoadBalancer:
		return true
	}
	return false
}

//GetQuotaServiceType returns ServiceType if the service type is eligible to track against a quota, nor return ""
func GetQuotaServiceType(service *v1.Service) v1.ServiceType {
	switch service.Spec.Type {
	case v1.ServiceTypeNodePort:
		return v1.ServiceTypeNodePort
	case v1.ServiceTypeLoadBalancer:
		return v1.ServiceTypeLoadBalancer
	}
	return v1.ServiceType("")
}
