package mock

import (
	"context"
	typesV1 "github.com/clambin/webmon/crds/targets/api/types/v1"
	v1 "github.com/clambin/webmon/crds/targets/clientset/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// New creates a new mocked API client
func New() (client *Client) {
	return &Client{
		Handler{
			watch: watch.NewFake(),
		},
	}
}

// Client structure for a mocked API client
type Client struct {
	// Handler implements the required API endpoints
	Handler Handler
}

// Targets emulates the API's Targets function. Currently, it passes through to the Handler,
// i.e. it does not support any namespace awareness.
func (client *Client) Targets(_ string) v1.TargetInterface {
	return &client.Handler
}

// Add test function.  Creates a watcher event indicating a customer resource has been added.
func (client *Client) Add(namespace, name, url string) {
	client.Handler.watch.Add(&typesV1.Target{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: typesV1.TargetSpec{
			URL: url,
		},
	})
}

// Delete test function.  Creates a watcher event indicating a customer resource has been deleted.
func (client *Client) Delete(namespace, name string) {
	client.Handler.watch.Delete(&typesV1.Target{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: typesV1.TargetSpec{
			URL: "",
		},
	})
}

// Modify test function.  Creates a watcher event indicating a customer resource has been modified.
func (client *Client) Modify(namespace, name, url string) {
	client.Handler.watch.Modify(&typesV1.Target{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: typesV1.TargetSpec{
			URL: url,
		},
	})
}

// Handler implements the supported endpoints
type Handler struct {
	// targets typesV1.TargetList
	// lock    sync.RWMutex
	watch *watch.FakeWatcher
}

// Watch creates a watcher that will notify when a target is added/updated/deleted
func (handler *Handler) Watch(_ context.Context, _ metav1.ListOptions) (watch.Interface, error) {
	return handler.watch, nil
}

/*
// List retrieves the list of Targets
func (handler *Handler) List(_ context.Context, _ metav1.ListOptions) (targets *typesV1.TargetList, err error) {
	return nil, nil
}
*/
