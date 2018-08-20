// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by lister-gen. DO NOT EDIT.

// This file was automatically generated by lister-gen

package v1alpha1

import (
	v1alpha1 "github.com/kubeflow/mxnet-operator/pkg/apis/mxnet/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MXJobLister helps list MXJobs.
type MXJobLister interface {
	// List lists all MXJobs in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.MXJob, err error)
	// MXJobs returns an object that can list and get MXJobs.
	MXJobs(namespace string) MXJobNamespaceLister
	MXJobListerExpansion
}

// mXJobLister implements the MXJobLister interface.
type mXJobLister struct {
	indexer cache.Indexer
}

// NewMXJobLister returns a new MXJobLister.
func NewMXJobLister(indexer cache.Indexer) MXJobLister {
	return &mXJobLister{indexer: indexer}
}

// List lists all MXJobs in the indexer.
func (s *mXJobLister) List(selector labels.Selector) (ret []*v1alpha1.MXJob, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MXJob))
	})
	return ret, err
}

// MXJobs returns an object that can list and get MXJobs.
func (s *mXJobLister) MXJobs(namespace string) MXJobNamespaceLister {
	return mXJobNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MXJobNamespaceLister helps list and get MXJobs.
type MXJobNamespaceLister interface {
	// List lists all MXJobs in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.MXJob, err error)
	// Get retrieves the MXJob from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.MXJob, error)
	MXJobNamespaceListerExpansion
}

// mXJobNamespaceLister implements the MXJobNamespaceLister
// interface.
type mXJobNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MXJobs in the indexer for a given namespace.
func (s mXJobNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MXJob, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MXJob))
	})
	return ret, err
}

// Get retrieves the MXJob from the indexer for a given namespace and name.
func (s mXJobNamespaceLister) Get(name string) (*v1alpha1.MXJob, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mxjob"), name)
	}
	return obj.(*v1alpha1.MXJob), nil
}
