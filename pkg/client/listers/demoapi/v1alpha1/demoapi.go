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
	v1alpha1 "github.com/ngsin/demo-controller/pkg/apis/demoapi/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DemoAPILister helps list DemoAPIs.
// All objects returned here must be treated as read-only.
type DemoAPILister interface {
	// List lists all DemoAPIs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DemoAPI, err error)
	// DemoAPIs returns an object that can list and get DemoAPIs.
	DemoAPIs(namespace string) DemoAPINamespaceLister
	DemoAPIListerExpansion
}

// demoAPILister implements the DemoAPILister interface.
type demoAPILister struct {
	indexer cache.Indexer
}

// NewDemoAPILister returns a new DemoAPILister.
func NewDemoAPILister(indexer cache.Indexer) DemoAPILister {
	return &demoAPILister{indexer: indexer}
}

// List lists all DemoAPIs in the indexer.
func (s *demoAPILister) List(selector labels.Selector) (ret []*v1alpha1.DemoAPI, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DemoAPI))
	})
	return ret, err
}

// DemoAPIs returns an object that can list and get DemoAPIs.
func (s *demoAPILister) DemoAPIs(namespace string) DemoAPINamespaceLister {
	return demoAPINamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DemoAPINamespaceLister helps list and get DemoAPIs.
// All objects returned here must be treated as read-only.
type DemoAPINamespaceLister interface {
	// List lists all DemoAPIs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.DemoAPI, err error)
	// Get retrieves the DemoAPI from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.DemoAPI, error)
	DemoAPINamespaceListerExpansion
}

// demoAPINamespaceLister implements the DemoAPINamespaceLister
// interface.
type demoAPINamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all DemoAPIs in the indexer for a given namespace.
func (s demoAPINamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.DemoAPI, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DemoAPI))
	})
	return ret, err
}

// Get retrieves the DemoAPI from the indexer for a given namespace and name.
func (s demoAPINamespaceLister) Get(name string) (*v1alpha1.DemoAPI, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("demoapi"), name)
	}
	return obj.(*v1alpha1.DemoAPI), nil
}