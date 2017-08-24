/*
Copyright 2017 The Kubernetes Authors.

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

// This file was automatically generated by informer-gen

package v1

import (
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	kubernetes "k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// ComponentStatusInformer provides access to a shared informer and lister for
// ComponentStatuses.
type ComponentStatusInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.ComponentStatusLister
}

type componentStatusInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

// NewComponentStatusInformer constructs a new informer for ComponentStatus type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewComponentStatusInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().ComponentStatuses().List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().ComponentStatuses().Watch(options)
			},
		},
		&core_v1.ComponentStatus{},
		resyncPeriod,
		indexers,
	)
}

// NewComponentStatusInformerWithOptions constructs a new informer for ComponentStatus type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewComponentStatusInformerWithOptions(client kubernetes.Interface, options cache.SharedInformerOptions, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformerWithOptions(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().ComponentStatuses().List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().ComponentStatuses().Watch(options)
			},
		},
		options,
		&core_v1.ComponentStatus{},
		indexers,
	)
}

func defaultComponentStatusInformer(client kubernetes.Interface, options cache.SharedInformerOptions) cache.SharedIndexInformer {
	return NewComponentStatusInformerWithOptions(client, options, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (f *componentStatusInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&core_v1.ComponentStatus{}, defaultComponentStatusInformer)
}

func (f *componentStatusInformer) Lister() v1.ComponentStatusLister {
	return v1.NewComponentStatusLister(f.Informer().GetIndexer())
}
