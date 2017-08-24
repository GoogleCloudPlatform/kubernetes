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

package v1beta1

import (
	rbac_v1beta1 "k8s.io/api/rbac/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	kubernetes "k8s.io/client-go/kubernetes"
	v1beta1 "k8s.io/client-go/listers/rbac/v1beta1"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// RoleInformer provides access to a shared informer and lister for
// Roles.
type RoleInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.RoleLister
}

type roleInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

// NewRoleInformer constructs a new informer for Role type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewRoleInformer(client kubernetes.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.RbacV1beta1().Roles(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.RbacV1beta1().Roles(namespace).Watch(options)
			},
		},
		&rbac_v1beta1.Role{},
		resyncPeriod,
		indexers,
	)
}

// NewRoleInformerWithOptions constructs a new informer for Role type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewRoleInformerWithOptions(client kubernetes.Interface, namespace string, options cache.SharedInformerOptions, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformerWithOptions(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.RbacV1beta1().Roles(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.RbacV1beta1().Roles(namespace).Watch(options)
			},
		},
		options,
		&rbac_v1beta1.Role{},
		indexers,
	)
}

func defaultRoleInformer(client kubernetes.Interface, options cache.SharedInformerOptions) cache.SharedIndexInformer {
	return NewRoleInformerWithOptions(client, v1.NamespaceAll, options, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (f *roleInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&rbac_v1beta1.Role{}, defaultRoleInformer)
}

func (f *roleInformer) Lister() v1beta1.RoleLister {
	return v1beta1.NewRoleLister(f.Informer().GetIndexer())
}
