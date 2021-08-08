package watcher

import (
	"context"
	v1 "github.com/clambin/webmon/crds/targets/api/types/v1"
	clientV1 "github.com/clambin/webmon/crds/targets/clientset/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

// A Watcher checks kubernetes custom resources on a periodic basis for newRegistry URLs to monitor
type Watcher struct {
	Client     clientV1.WebMonV1Interface
	register   chan string
	unregister chan string
	namespace  string
	store      *registry
}

// NewWithClient creates a newRegistry Watcher for the specified API client. When Watcher finds a created/removed URL, it sends
// the URL to the register/unregister channel respectively.
// If the namespace is specified, Watcher will only scan that namespace. Otherwise, all namespaces are scanned.
// Note that this needs RBAC setup to ensure the client can access those resources.
func NewWithClient(register, unregister chan string, namespace string, client clientV1.WebMonV1Interface) *Watcher {
	return &Watcher{
		Client:     client,
		register:   register,
		unregister: unregister,
		namespace:  namespace,
		store:      newRegistry(),
	}
}

// Run checks kubernetes for created/removed URLs at the specified interval
func (watcher *Watcher) Run(ctx context.Context) {
	w, resultChan := watcher.watch(ctx)

	// k8s may stop sending events after 30m
	ticker := time.NewTicker(30 * time.Minute)

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case event := <-resultChan:
			watcher.processEvent(event)
		case <-ticker.C:
			// recreate the watch session to prevent timeouts
			log.Debug("renewing custom resource watcher")
			w.Stop()
			w, resultChan = watcher.watch(ctx)
			log.Debug("renewed custom resource watcher")
		}
	}
	ticker.Stop()
	w.Stop()
}

func (watcher *Watcher) watch(ctx context.Context) (w watch.Interface, ch <-chan watch.Event) {
	var err error
	w, err = watcher.Client.Targets(watcher.namespace).Watch(ctx, metav1.ListOptions{})

	if err == nil {
		ch = w.ResultChan()
	} else {
		log.WithError(err).Fatal("failed to set up custom resource watcher")
	}
	return
}

func (watcher *Watcher) processEvent(event watch.Event) {
	target := event.Object.(*v1.Target)
	log.WithFields(log.Fields{
		"name":      target.Name,
		"namespace": target.Namespace,
		"url":       target.Spec.URL,
		"event":     event.Type,
	}).Debug("event received")

	switch event.Type {
	case watch.Added:
		watcher.store.add(target.Namespace, target.Name, target.Spec.URL)
		watcher.register <- target.Spec.URL
	case watch.Deleted:
		url := watcher.store.delete(target.Namespace, target.Name)
		watcher.unregister <- url
	case watch.Modified:
		url, ok := watcher.store.get(target.Namespace, target.Name)
		if ok && url != target.Spec.URL {
			watcher.unregister <- url
			watcher.store.add(target.Namespace, target.Name, target.Spec.URL)
			watcher.register <- target.Spec.URL
		}
	}
}
