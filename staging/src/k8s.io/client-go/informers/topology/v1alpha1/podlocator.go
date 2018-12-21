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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	topologyv1alpha1 "k8s.io/api/topology/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	kubernetes "k8s.io/client-go/kubernetes"
	v1alpha1 "k8s.io/client-go/listers/topology/v1alpha1"
	cache "k8s.io/client-go/tools/cache"
)

// PodLocatorInformer provides access to a shared informer and lister for
// PodLocators.
type PodLocatorInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.PodLocatorLister
}

type podLocatorInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewPodLocatorInformer constructs a new informer for PodLocator type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewPodLocatorInformer(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredPodLocatorInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredPodLocatorInformer constructs a new informer for PodLocator type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredPodLocatorInformer(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TopologyV1alpha1().PodLocators(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TopologyV1alpha1().PodLocators(namespace).Watch(options)
			},
		},
		&topologyv1alpha1.PodLocator{},
		resyncPeriod,
		indexers,
	)
}

func (f *podLocatorInformer) defaultInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPodLocatorInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *podLocatorInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&topologyv1alpha1.PodLocator{}, f.defaultInformer)
}

func (f *podLocatorInformer) Lister() v1alpha1.PodLocatorLister {
	return v1alpha1.NewPodLocatorLister(f.Informer().GetIndexer())
}
