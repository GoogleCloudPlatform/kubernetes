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

package admission

import (
	"fmt"
	"io"
	"strings"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	admission "k8s.io/apiserver/pkg/admission"
	api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/storage"
	storageutil "k8s.io/kubernetes/pkg/apis/storage/util"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"
	storagelisters "k8s.io/kubernetes/pkg/client/listers/storage/internalversion"
	kubeapiserveradmission "k8s.io/kubernetes/pkg/kubeapiserver/admission"
)

const (
	PluginName = "DefaultStorageClass"
)

func init() {
	admission.RegisterPlugin(PluginName, func(config io.Reader) (admission.Interface, error) {
		plugin := newPlugin()
		return plugin, nil
	})
}

// claimDefaulterPlugin holds state for and implements the admission plugin.
type claimDefaulterPlugin struct {
	*admission.Handler

	lister storagelisters.StorageClassLister
}

var _ admission.Interface = &claimDefaulterPlugin{}
var _ = kubeapiserveradmission.WantsInformerFactory(&claimDefaulterPlugin{})

// newPlugin creates a new admission plugin.
func newPlugin() *claimDefaulterPlugin {
	return &claimDefaulterPlugin{
		Handler: admission.NewHandler(admission.Create),
	}
}

func (a *claimDefaulterPlugin) SetInformerFactory(f informers.SharedInformerFactory) {
	informer := f.Storage().InternalVersion().StorageClasses()
	a.lister = informer.Lister()
	a.SetReadyFunc(informer.Informer().HasSynced)
}

// Validate ensures an authorizer is set.
func (a *claimDefaulterPlugin) Validate() error {
	if a.lister == nil {
		return fmt.Errorf("missing lister")
	}
	return nil
}

// Admit sets the default value of a PersistentVolumeClaim's storage class, in case the user did
// not provide a value.
//
// 1.  Find available StorageClasses.
// 2.  Figure which is the default
// 3.  Write to the PVClaim
func (c *claimDefaulterPlugin) Admit(a admission.Attributes) error {
	if a.GetResource().GroupResource() != api.Resource("persistentvolumeclaims") {
		return nil
	}

	if len(a.GetSubresource()) != 0 {
		return nil
	}

	pvc, ok := a.GetObject().(*api.PersistentVolumeClaim)
	// if we can't convert then we don't handle this object so just return
	if !ok {
		return nil
	}

	if storageutil.HasStorageClassAnnotation(pvc.ObjectMeta) {
		// The user asked for a class.
		// check and make sure the user has access to this class
		// if they do not, then return a NewForbidden error to prevent creation
		err := isValidStorageClassNamespace(c.store, pvc)
		if err != nil {
			return admission.NewForbidden(a, err)
		}

		return nil
	}

	glog.V(4).Infof("no storage class for claim %s (generate: %s)", pvc.Name, pvc.GenerateName)

	def, err := getDefaultClass(c.lister)
	if err != nil {
		return admission.NewForbidden(a, err)
	}
	if def == nil {
		// No default class selected, do nothing about the PVC.
		return nil
	}

	glog.V(4).Infof("defaulting storage class for claim %s (generate: %s) to %s", pvc.Name, pvc.GenerateName, def.Name)
	if pvc.ObjectMeta.Annotations == nil {
		pvc.ObjectMeta.Annotations = map[string]string{}
	}
	pvc.Annotations[storageutil.StorageClassAnnotation] = def.Name
	return nil
}

// getDefaultClass returns the default StorageClass from the store, or nil.
func getDefaultClass(lister storagelisters.StorageClassLister) (*storage.StorageClass, error) {
	list, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	defaultClasses := []*storage.StorageClass{}
	for _, class := range list {
		if storageutil.IsDefaultAnnotation(class.ObjectMeta) {
			defaultClasses = append(defaultClasses, class)
			glog.V(4).Infof("getDefaultClass added: %s", class.Name)
		}
	}

	if len(defaultClasses) == 0 {
		return nil, nil
	}
	if len(defaultClasses) > 1 {
		glog.V(4).Infof("getDefaultClass %s defaults found", len(defaultClasses))
		return nil, errors.NewInternalError(fmt.Errorf("%d default StorageClasses were found", len(defaultClasses)))
	}
	return defaultClasses[0], nil
}

// isValidStorageClassNamespace checks to see if a StorageClass has been restricted to what
// namespaces can use it via an annotation, this will return an error if the requested StorageClass on
// the claim does not match the StorageClass namespaces annotation and Name.
func isValidStorageClassNamespace(store cache.Store, claim *api.PersistentVolumeClaim) error {
	sc := []*storage.StorageClass{}

	// if no storageclasses exist yet, then we can return nil and wait
	// for the storageclass as normal
	if len(store.List()) == 0 {
		return nil
	}

	for _, c := range store.List() {
		class, ok := c.(*storage.StorageClass)
		if !ok {
			return errors.NewInternalError(fmt.Errorf("error converting stored object to StorageClass: %v", c))
		}
		// Check the storageclass Namespace Annotation,
		// if exists, split the strings and see if we match
		scProjects := storageutil.GetStorageClassNamespaces(class.ObjectMeta)
		if len(scProjects) > 0 {
			result := strings.Split(scProjects, ",")
			for i := range result {
				// annotation must equal Namespace AND class Name
				if strings.TrimSpace(result[i]) == claim.Namespace && class.Name == storageutil.GetClaimStorageClass(claim){
					return nil
				}
			}
		} else {
			// StorageClass has no annotation for namespace and therefor no
			// restrictions
			sc = append(sc, class)
		}
	}

	if len(sc) > 0 {
		// we have at least one storageclass that exists
		// and no restrictions
		return nil
	}

	return errors.NewInternalError(fmt.Errorf("claim %s in namespace %s is not allowed to use StorageClass %s - contact the cluster-admin or storage-admin for access", claim.Name, claim.Namespace, storageutil.GetStorageClassAnnotation(claim.ObjectMeta)))
}