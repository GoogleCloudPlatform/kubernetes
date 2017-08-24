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

package v2alpha1

import (
	autoscaling_v2alpha1 "k8s.io/api/autoscaling/v2alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	kubernetes "k8s.io/client-go/kubernetes"
	v2alpha1 "k8s.io/client-go/listers/autoscaling/v2alpha1"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// HorizontalPodAutoscalerInformer provides access to a shared informer and lister for
// HorizontalPodAutoscalers.
type HorizontalPodAutoscalerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v2alpha1.HorizontalPodAutoscalerLister
}

type horizontalPodAutoscalerInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

// NewHorizontalPodAutoscalerInformer constructs a new informer for HorizontalPodAutoscaler type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewHorizontalPodAutoscalerInformer(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.AutoscalingV2alpha1().HorizontalPodAutoscalers(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.AutoscalingV2alpha1().HorizontalPodAutoscalers(namespace).Watch(options)
			},
		},
		&autoscaling_v2alpha1.HorizontalPodAutoscaler{},
		resyncPeriod,
		indexers,
	)
}

// NewHorizontalPodAutoscalerInformerWithOptions constructs a new informer for HorizontalPodAutoscaler type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewHorizontalPodAutoscalerInformerWithOptions(client kubernetes.Interface, namespace string, options cache.SharedInformerOptions, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformerWithOptions(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.AutoscalingV2alpha1().HorizontalPodAutoscalers(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.AutoscalingV2alpha1().HorizontalPodAutoscalers(namespace).Watch(options)
			},
		},
		options,
		&autoscaling_v2alpha1.HorizontalPodAutoscaler{},
		indexers,
	)
}

func defaultHorizontalPodAutoscalerInformer(client kubernetes.Interface, options cache.SharedInformerOptions) cache.SharedIndexInformer {
	return NewHorizontalPodAutoscalerInformerWithOptions(client, v1.NamespaceAll, options, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (f *horizontalPodAutoscalerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&autoscaling_v2alpha1.HorizontalPodAutoscaler{}, defaultHorizontalPodAutoscalerInformer)
}

func (f *horizontalPodAutoscalerInformer) Lister() v2alpha1.HorizontalPodAutoscalerLister {
	return v2alpha1.NewHorizontalPodAutoscalerLister(f.Informer().GetIndexer())
}
