/*
Copyright 2014 Google Inc. All rights reserved.

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

package resource

import (
	"github.com/golang/glog"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/meta"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
)

// Selector is a Visitor for resources that match a label selector.
type Selector struct {
	Client    RESTClient
	Mapping   *meta.RESTMapping
	Namespace string
	Selector  labels.Selector
}

// NewSelector creates a resource selector which hides details of getting items by their label selector.
func NewSelector(client RESTClient, mapping *meta.RESTMapping, namespace string, selector labels.Selector) *Selector {
	return &Selector{
		Client:    client,
		Mapping:   mapping,
		Namespace: namespace,
		Selector:  selector,
	}
}

// Visit implements Visitor
func (r *Selector) Visit(fn VisitorFunc) error {
	list, err := NewHelper(r.Client, r.Mapping).List(r.Namespace, r.Selector)
	if err != nil {
		if errors.IsBadRequest(err) || errors.IsNotFound(err) {
			if r.Selector.Empty() {
				glog.V(2).Infof("Unable to list %q: %v", r.Mapping.Resource, err)
			} else {
				glog.V(2).Infof("Unable to find %q that match the selector %q: %v", r.Mapping.Resource, r.Selector, err)
			}
			return nil
		}
		return err
	}
	accessor := r.Mapping.MetadataAccessor
	resourceVersion, _ := accessor.ResourceVersion(list)
	info := &Info{
		Client:    r.Client,
		Mapping:   r.Mapping,
		Namespace: r.Namespace,

		Object:          list,
		ResourceVersion: resourceVersion,
	}
	return fn(info)
}

func (r *Selector) Watch(resourceVersion string) (watch.Interface, error) {
	return NewHelper(r.Client, r.Mapping).Watch(r.Namespace, resourceVersion, r.Selector, labels.Everything())
}

// ResourceMapping returns the mapping for this resource and implements ResourceMapping
func (r *Selector) ResourceMapping() *meta.RESTMapping {
	return r.Mapping
}
