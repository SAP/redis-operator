/*
SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and redis-operator contributors
SPDX-License-Identifier: Apache-2.0
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/sap/redis-operator/api/v1alpha1"
	scheme "github.com/sap/redis-operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// RedisGetter has a method to return a RedisInterface.
// A group's client should implement this interface.
type RedisGetter interface {
	Redis(namespace string) RedisInterface
}

// RedisInterface has methods to work with Redis resources.
type RedisInterface interface {
	Create(ctx context.Context, redis *v1alpha1.Redis, opts v1.CreateOptions) (*v1alpha1.Redis, error)
	Update(ctx context.Context, redis *v1alpha1.Redis, opts v1.UpdateOptions) (*v1alpha1.Redis, error)
	UpdateStatus(ctx context.Context, redis *v1alpha1.Redis, opts v1.UpdateOptions) (*v1alpha1.Redis, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Redis, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.RedisList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Redis, err error)
	RedisExpansion
}

// redis implements RedisInterface
type redis struct {
	client rest.Interface
	ns     string
}

// newRedis returns a Redis
func newRedis(c *CacheV1alpha1Client, namespace string) *redis {
	return &redis{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the redis, and returns the corresponding redis object, and an error if there is any.
func (c *redis) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Redis, err error) {
	result = &v1alpha1.Redis{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("redis").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Redis that match those selectors.
func (c *redis) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RedisList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.RedisList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("redis").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested redis.
func (c *redis) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("redis").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a redis and creates it.  Returns the server's representation of the redis, and an error, if there is any.
func (c *redis) Create(ctx context.Context, redis *v1alpha1.Redis, opts v1.CreateOptions) (result *v1alpha1.Redis, err error) {
	result = &v1alpha1.Redis{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("redis").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(redis).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a redis and updates it. Returns the server's representation of the redis, and an error, if there is any.
func (c *redis) Update(ctx context.Context, redis *v1alpha1.Redis, opts v1.UpdateOptions) (result *v1alpha1.Redis, err error) {
	result = &v1alpha1.Redis{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("redis").
		Name(redis.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(redis).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *redis) UpdateStatus(ctx context.Context, redis *v1alpha1.Redis, opts v1.UpdateOptions) (result *v1alpha1.Redis, err error) {
	result = &v1alpha1.Redis{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("redis").
		Name(redis.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(redis).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the redis and deletes it. Returns an error if one occurs.
func (c *redis) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("redis").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *redis) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("redis").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched redis.
func (c *redis) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Redis, err error) {
	result = &v1alpha1.Redis{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("redis").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
