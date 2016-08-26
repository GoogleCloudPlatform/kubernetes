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

package master

import (
	"fmt"
	"sync"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apis/rbac"
	rbacapiv1alpha1 "k8s.io/kubernetes/pkg/apis/rbac/v1alpha1"
	rbacvalidation "k8s.io/kubernetes/pkg/apis/rbac/validation"
	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/registry/clusterrole"
	clusterroleetcd "k8s.io/kubernetes/pkg/registry/clusterrole/etcd"
	clusterrolepolicybased "k8s.io/kubernetes/pkg/registry/clusterrole/policybased"
	"k8s.io/kubernetes/pkg/registry/clusterrolebinding"
	clusterrolebindingetcd "k8s.io/kubernetes/pkg/registry/clusterrolebinding/etcd"
	clusterrolebindingpolicybased "k8s.io/kubernetes/pkg/registry/clusterrolebinding/policybased"
	"k8s.io/kubernetes/pkg/registry/role"
	roleetcd "k8s.io/kubernetes/pkg/registry/role/etcd"
	rolepolicybased "k8s.io/kubernetes/pkg/registry/role/policybased"
	"k8s.io/kubernetes/pkg/registry/rolebinding"
	rolebindingetcd "k8s.io/kubernetes/pkg/registry/rolebinding/etcd"
	rolebindingpolicybased "k8s.io/kubernetes/pkg/registry/rolebinding/policybased"
	utilruntime "k8s.io/kubernetes/pkg/util/runtime"
	"k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac/bootstrappolicy"
)

type RBACRESTStorageProvider struct {
	AuthorizerRBACSuperUser string
}

var _ RESTStorageProvider = &RBACRESTStorageProvider{}
var _ RESTStoragePostStartHookProvider = &RBACRESTStorageProvider{}

func (p RBACRESTStorageProvider) NewRESTStorage(apiResourceConfigSource genericapiserver.APIResourceConfigSource, restOptionsGetter RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(rbac.GroupName)

	if apiResourceConfigSource.AnyResourcesForVersionEnabled(rbacapiv1alpha1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[rbacapiv1alpha1.SchemeGroupVersion.Version] = p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
		apiGroupInfo.GroupMeta.GroupVersion = rbacapiv1alpha1.SchemeGroupVersion
	}

	return apiGroupInfo, true
}

func (p RBACRESTStorageProvider) v1alpha1Storage(apiResourceConfigSource genericapiserver.APIResourceConfigSource, restOptionsGetter RESTOptionsGetter) map[string]rest.Storage {
	version := rbacapiv1alpha1.SchemeGroupVersion

	once := new(sync.Once)
	var authorizationRuleResolver rbacvalidation.AuthorizationRuleResolver
	newRuleValidator := func() rbacvalidation.AuthorizationRuleResolver {
		once.Do(func() {
			authorizationRuleResolver = rbacvalidation.NewDefaultRuleResolver(
				role.NewRegistry(roleetcd.NewREST(restOptionsGetter(rbac.Resource("roles")))),
				rolebinding.NewRegistry(rolebindingetcd.NewREST(restOptionsGetter(rbac.Resource("rolebindings")))),
				clusterrole.NewRegistry(clusterroleetcd.NewREST(restOptionsGetter(rbac.Resource("clusterroles")))),
				clusterrolebinding.NewRegistry(clusterrolebindingetcd.NewREST(restOptionsGetter(rbac.Resource("clusterrolebindings")))),
			)
		})
		return authorizationRuleResolver
	}

	storage := map[string]rest.Storage{}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("roles")) {
		rolesStorage := roleetcd.NewREST(restOptionsGetter(rbac.Resource("roles")))
		storage["roles"] = rolepolicybased.NewStorage(rolesStorage, newRuleValidator(), p.AuthorizerRBACSuperUser)
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("rolebindings")) {
		roleBindingsStorage := rolebindingetcd.NewREST(restOptionsGetter(rbac.Resource("rolebindings")))
		storage["rolebindings"] = rolebindingpolicybased.NewStorage(roleBindingsStorage, newRuleValidator(), p.AuthorizerRBACSuperUser)
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("clusterroles")) {
		clusterRolesStorage := clusterroleetcd.NewREST(restOptionsGetter(rbac.Resource("clusterroles")))
		storage["clusterroles"] = clusterrolepolicybased.NewStorage(clusterRolesStorage, newRuleValidator(), p.AuthorizerRBACSuperUser)
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("clusterrolebindings")) {
		clusterRoleBindingsStorage := clusterrolebindingetcd.NewREST(restOptionsGetter(rbac.Resource("clusterrolebindings")))
		storage["clusterrolebindings"] = clusterrolebindingpolicybased.NewStorage(clusterRoleBindingsStorage, newRuleValidator(), p.AuthorizerRBACSuperUser)
	}
	return storage
}

func (p RBACRESTStorageProvider) PostStartHook(apiResourceConfigSource genericapiserver.APIResourceConfigSource, restOptionsGetter RESTOptionsGetter) func() {
	var directClusterRoleAccess *clusterroleetcd.REST

	if apiResourceConfigSource.AnyResourcesForVersionEnabled(rbacapiv1alpha1.SchemeGroupVersion) {
		if apiResourceConfigSource.ResourceEnabled(rbacapiv1alpha1.SchemeGroupVersion.WithResource("clusterroles")) {
			directClusterRoleAccess = clusterroleetcd.NewREST(restOptionsGetter(rbac.Resource("clusterroles")))
		}
	}

	if directClusterRoleAccess == nil {
		return func() {}
	}

	return func() {
		defer utilruntime.HandleCrash()

		ctx := api.NewContext()

		existingClusterRoles, err := directClusterRoleAccess.List(ctx, &api.ListOptions{})
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("unable to initialize clusterroles: %v", err))
			return
		}
		// if clusterroles already exist, then assume we don't have work to do
		if len(existingClusterRoles.(*rbac.ClusterRoleList).Items) > 0 {
			return
		}

		for _, clusterRole := range bootstrappolicy.ClusterRoles() {
			if _, err := directClusterRoleAccess.Create(ctx, &clusterRole); err != nil {
				// don't fail on failures, try to create as many as you can
				utilruntime.HandleError(fmt.Errorf("unable to initialize clusterroles: %v", err))
				continue
			}
			glog.Infof("Created clusterrole.%s/%s", rbac.GroupName, clusterRole.Name)
		}
	}
}
