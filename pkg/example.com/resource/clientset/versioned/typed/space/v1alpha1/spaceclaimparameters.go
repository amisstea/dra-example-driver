/*
 * Copyright 2024 The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "sigs.k8s.io/dra-example-driver/api/example.com/resource/space/v1alpha1"
	scheme "sigs.k8s.io/dra-example-driver/pkg/example.com/resource/clientset/versioned/scheme"
)

// SpaceClaimParametersGetter has a method to return a SpaceClaimParametersInterface.
// A group's client should implement this interface.
type SpaceClaimParametersGetter interface {
	SpaceClaimParameters(namespace string) SpaceClaimParametersInterface
}

// SpaceClaimParametersInterface has methods to work with SpaceClaimParameters resources.
type SpaceClaimParametersInterface interface {
	Create(ctx context.Context, spaceClaimParameters *v1alpha1.SpaceClaimParameters, opts v1.CreateOptions) (*v1alpha1.SpaceClaimParameters, error)
	Update(ctx context.Context, spaceClaimParameters *v1alpha1.SpaceClaimParameters, opts v1.UpdateOptions) (*v1alpha1.SpaceClaimParameters, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.SpaceClaimParameters, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.SpaceClaimParametersList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SpaceClaimParameters, err error)
	SpaceClaimParametersExpansion
}

// spaceClaimParameters implements SpaceClaimParametersInterface
type spaceClaimParameters struct {
	client rest.Interface
	ns     string
}

// newSpaceClaimParameters returns a SpaceClaimParameters
func newSpaceClaimParameters(c *SpaceV1alpha1Client, namespace string) *spaceClaimParameters {
	return &spaceClaimParameters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the spaceClaimParameters, and returns the corresponding spaceClaimParameters object, and an error if there is any.
func (c *spaceClaimParameters) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SpaceClaimParameters, err error) {
	result = &v1alpha1.SpaceClaimParameters{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SpaceClaimParameters that match those selectors.
func (c *spaceClaimParameters) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SpaceClaimParametersList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SpaceClaimParametersList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested spaceClaimParameters.
func (c *spaceClaimParameters) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a spaceClaimParameters and creates it.  Returns the server's representation of the spaceClaimParameters, and an error, if there is any.
func (c *spaceClaimParameters) Create(ctx context.Context, spaceClaimParameters *v1alpha1.SpaceClaimParameters, opts v1.CreateOptions) (result *v1alpha1.SpaceClaimParameters, err error) {
	result = &v1alpha1.SpaceClaimParameters{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(spaceClaimParameters).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a spaceClaimParameters and updates it. Returns the server's representation of the spaceClaimParameters, and an error, if there is any.
func (c *spaceClaimParameters) Update(ctx context.Context, spaceClaimParameters *v1alpha1.SpaceClaimParameters, opts v1.UpdateOptions) (result *v1alpha1.SpaceClaimParameters, err error) {
	result = &v1alpha1.SpaceClaimParameters{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		Name(spaceClaimParameters.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(spaceClaimParameters).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the spaceClaimParameters and deletes it. Returns an error if one occurs.
func (c *spaceClaimParameters) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *spaceClaimParameters) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched spaceClaimParameters.
func (c *spaceClaimParameters) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SpaceClaimParameters, err error) {
	result = &v1alpha1.SpaceClaimParameters{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("spaceclaimparameters").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
