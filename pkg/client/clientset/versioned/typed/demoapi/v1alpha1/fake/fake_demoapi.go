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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/ngsin/demo-controller/pkg/apis/demoapi/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDemoAPIs implements DemoAPIInterface
type FakeDemoAPIs struct {
	Fake *FakeDemogroupV1alpha1
	ns   string
}

var demoapisResource = schema.GroupVersionResource{Group: "demogroup.k8s.io", Version: "v1alpha1", Resource: "demoapis"}

var demoapisKind = schema.GroupVersionKind{Group: "demogroup.k8s.io", Version: "v1alpha1", Kind: "DemoAPI"}

// Get takes name of the demoAPI, and returns the corresponding demoAPI object, and an error if there is any.
func (c *FakeDemoAPIs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DemoAPI, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(demoapisResource, c.ns, name), &v1alpha1.DemoAPI{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DemoAPI), err
}

// List takes label and field selectors, and returns the list of DemoAPIs that match those selectors.
func (c *FakeDemoAPIs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DemoAPIList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(demoapisResource, demoapisKind, c.ns, opts), &v1alpha1.DemoAPIList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DemoAPIList{ListMeta: obj.(*v1alpha1.DemoAPIList).ListMeta}
	for _, item := range obj.(*v1alpha1.DemoAPIList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested demoAPIs.
func (c *FakeDemoAPIs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(demoapisResource, c.ns, opts))

}

// Create takes the representation of a demoAPI and creates it.  Returns the server's representation of the demoAPI, and an error, if there is any.
func (c *FakeDemoAPIs) Create(ctx context.Context, demoAPI *v1alpha1.DemoAPI, opts v1.CreateOptions) (result *v1alpha1.DemoAPI, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(demoapisResource, c.ns, demoAPI), &v1alpha1.DemoAPI{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DemoAPI), err
}

// Update takes the representation of a demoAPI and updates it. Returns the server's representation of the demoAPI, and an error, if there is any.
func (c *FakeDemoAPIs) Update(ctx context.Context, demoAPI *v1alpha1.DemoAPI, opts v1.UpdateOptions) (result *v1alpha1.DemoAPI, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(demoapisResource, c.ns, demoAPI), &v1alpha1.DemoAPI{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DemoAPI), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDemoAPIs) UpdateStatus(ctx context.Context, demoAPI *v1alpha1.DemoAPI, opts v1.UpdateOptions) (*v1alpha1.DemoAPI, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(demoapisResource, "status", c.ns, demoAPI), &v1alpha1.DemoAPI{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DemoAPI), err
}

// Delete takes name of the demoAPI and deletes it. Returns an error if one occurs.
func (c *FakeDemoAPIs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(demoapisResource, c.ns, name, opts), &v1alpha1.DemoAPI{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDemoAPIs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(demoapisResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DemoAPIList{})
	return err
}

// Patch applies the patch and returns the patched demoAPI.
func (c *FakeDemoAPIs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DemoAPI, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(demoapisResource, c.ns, name, pt, data, subresources...), &v1alpha1.DemoAPI{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DemoAPI), err
}
