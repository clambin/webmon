package v1

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// TargetCRDInterface lists the API endpoints. Currently, only watch is supported
type TargetCRDInterface interface {
	// Watch creates a watcher that receives notifications when a target is added/updated/deleted
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	// List(ctx context.Context, opts metav1.ListOptions) (*v1.TargetList, error)
}

type targetCRDClient struct {
	restClient rest.Interface
	ns         string
}

// Watch creates a watcher that will notify when a target is added/updated/deleted
func (c *targetCRDClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("targets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

/*

// List retrieves the list of Targets found in the provided namespace, or all namespaces
func (c *targetCRDClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.TargetList, error) {
	result := v1.TargetList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("targets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}
*/
