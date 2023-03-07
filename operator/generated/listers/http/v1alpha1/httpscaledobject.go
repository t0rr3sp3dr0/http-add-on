/*
Copyright 2023 The KEDA Authors.

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
	v1alpha1 "github.com/kedacore/http-add-on/operator/apis/http/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// HTTPScaledObjectLister helps list HTTPScaledObjects.
// All objects returned here must be treated as read-only.
type HTTPScaledObjectLister interface {
	// List lists all HTTPScaledObjects in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.HTTPScaledObject, err error)
	// HTTPScaledObjects returns an object that can list and get HTTPScaledObjects.
	HTTPScaledObjects(namespace string) HTTPScaledObjectNamespaceLister
	HTTPScaledObjectListerExpansion
}

// hTTPScaledObjectLister implements the HTTPScaledObjectLister interface.
type hTTPScaledObjectLister struct {
	indexer cache.Indexer
}

// NewHTTPScaledObjectLister returns a new HTTPScaledObjectLister.
func NewHTTPScaledObjectLister(indexer cache.Indexer) HTTPScaledObjectLister {
	return &hTTPScaledObjectLister{indexer: indexer}
}

// List lists all HTTPScaledObjects in the indexer.
func (s *hTTPScaledObjectLister) List(selector labels.Selector) (ret []*v1alpha1.HTTPScaledObject, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.HTTPScaledObject))
	})
	return ret, err
}

// HTTPScaledObjects returns an object that can list and get HTTPScaledObjects.
func (s *hTTPScaledObjectLister) HTTPScaledObjects(namespace string) HTTPScaledObjectNamespaceLister {
	return hTTPScaledObjectNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// HTTPScaledObjectNamespaceLister helps list and get HTTPScaledObjects.
// All objects returned here must be treated as read-only.
type HTTPScaledObjectNamespaceLister interface {
	// List lists all HTTPScaledObjects in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.HTTPScaledObject, err error)
	// Get retrieves the HTTPScaledObject from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.HTTPScaledObject, error)
	HTTPScaledObjectNamespaceListerExpansion
}

// hTTPScaledObjectNamespaceLister implements the HTTPScaledObjectNamespaceLister
// interface.
type hTTPScaledObjectNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all HTTPScaledObjects in the indexer for a given namespace.
func (s hTTPScaledObjectNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.HTTPScaledObject, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.HTTPScaledObject))
	})
	return ret, err
}

// Get retrieves the HTTPScaledObject from the indexer for a given namespace and name.
func (s hTTPScaledObjectNamespaceLister) Get(name string) (*v1alpha1.HTTPScaledObject, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("httpscaledobject"), name)
	}
	return obj.(*v1alpha1.HTTPScaledObject), nil
}
