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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "k8s.io/sample-apiserver/pkg/apis/wardle/v1alpha1"
)

// FischerLister helps list Fischers.
// All objects returned here must be treated as read-only.
type FischerLister interface {
	// List lists all Fischers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Fischer, err error)
	// Get retrieves the Fischer from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Fischer, error)
	FischerListerExpansion
}

// fischerLister implements the FischerLister interface.
type fischerLister struct {
	indexer cache.Indexer
}

// NewFischerLister returns a new FischerLister.
func NewFischerLister(indexer cache.Indexer) FischerLister {
	return &fischerLister{indexer: indexer}
}

// List lists all Fischers in the indexer.
func (s *fischerLister) List(selector labels.Selector) (ret []*v1alpha1.Fischer, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Fischer))
	})
	return ret, err
}

// Get retrieves the Fischer from the index for a given name.
func (s *fischerLister) Get(name string) (*v1alpha1.Fischer, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("fischer"), name)
	}
	return obj.(*v1alpha1.Fischer), nil
}
