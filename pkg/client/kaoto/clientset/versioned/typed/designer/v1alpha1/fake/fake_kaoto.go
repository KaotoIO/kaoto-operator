/*
Copyright 2023.

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
	json "encoding/json"
	"fmt"

	v1alpha1 "github.com/kaotoIO/kaoto-operator/apis/designer/v1alpha1"
	designerv1alpha1 "github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/applyconfiguration/designer/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKaotos implements KaotoInterface
type FakeKaotos struct {
	Fake *FakeDesignerV1alpha1
	ns   string
}

var kaotosResource = v1alpha1.SchemeGroupVersion.WithResource("kaotos")

var kaotosKind = v1alpha1.SchemeGroupVersion.WithKind("Kaoto")

// Get takes name of the kaoto, and returns the corresponding kaoto object, and an error if there is any.
func (c *FakeKaotos) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Kaoto, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kaotosResource, c.ns, name), &v1alpha1.Kaoto{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Kaoto), err
}

// List takes label and field selectors, and returns the list of Kaotos that match those selectors.
func (c *FakeKaotos) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.KaotoList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kaotosResource, kaotosKind, c.ns, opts), &v1alpha1.KaotoList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.KaotoList{ListMeta: obj.(*v1alpha1.KaotoList).ListMeta}
	for _, item := range obj.(*v1alpha1.KaotoList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kaotos.
func (c *FakeKaotos) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kaotosResource, c.ns, opts))

}

// Create takes the representation of a kaoto and creates it.  Returns the server's representation of the kaoto, and an error, if there is any.
func (c *FakeKaotos) Create(ctx context.Context, kaoto *v1alpha1.Kaoto, opts v1.CreateOptions) (result *v1alpha1.Kaoto, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kaotosResource, c.ns, kaoto), &v1alpha1.Kaoto{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Kaoto), err
}

// Update takes the representation of a kaoto and updates it. Returns the server's representation of the kaoto, and an error, if there is any.
func (c *FakeKaotos) Update(ctx context.Context, kaoto *v1alpha1.Kaoto, opts v1.UpdateOptions) (result *v1alpha1.Kaoto, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kaotosResource, c.ns, kaoto), &v1alpha1.Kaoto{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Kaoto), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeKaotos) UpdateStatus(ctx context.Context, kaoto *v1alpha1.Kaoto, opts v1.UpdateOptions) (*v1alpha1.Kaoto, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(kaotosResource, "status", c.ns, kaoto), &v1alpha1.Kaoto{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Kaoto), err
}

// Delete takes name of the kaoto and deletes it. Returns an error if one occurs.
func (c *FakeKaotos) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(kaotosResource, c.ns, name, opts), &v1alpha1.Kaoto{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKaotos) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kaotosResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.KaotoList{})
	return err
}

// Patch applies the patch and returns the patched kaoto.
func (c *FakeKaotos) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Kaoto, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kaotosResource, c.ns, name, pt, data, subresources...), &v1alpha1.Kaoto{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Kaoto), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied kaoto.
func (c *FakeKaotos) Apply(ctx context.Context, kaoto *designerv1alpha1.KaotoApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Kaoto, err error) {
	if kaoto == nil {
		return nil, fmt.Errorf("kaoto provided to Apply must not be nil")
	}
	data, err := json.Marshal(kaoto)
	if err != nil {
		return nil, err
	}
	name := kaoto.Name
	if name == nil {
		return nil, fmt.Errorf("kaoto.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kaotosResource, c.ns, *name, types.ApplyPatchType, data), &v1alpha1.Kaoto{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Kaoto), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeKaotos) ApplyStatus(ctx context.Context, kaoto *designerv1alpha1.KaotoApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.Kaoto, err error) {
	if kaoto == nil {
		return nil, fmt.Errorf("kaoto provided to Apply must not be nil")
	}
	data, err := json.Marshal(kaoto)
	if err != nil {
		return nil, err
	}
	name := kaoto.Name
	if name == nil {
		return nil, fmt.Errorf("kaoto.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kaotosResource, c.ns, *name, types.ApplyPatchType, data, "status"), &v1alpha1.Kaoto{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Kaoto), err
}