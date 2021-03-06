package watcher

import (
	"context"
	v1 "github.com/clambin/webmon/crds/targets/api/types/v1"
	clientV1 "github.com/clambin/webmon/crds/targets/clientset/v1"
	"github.com/clambin/webmon/monitor"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

// A Watcher checks kubernetes custom resources ("Target") on a periodic basis for new URLs to monitor
type Watcher struct {
	Client     clientV1.TargetsCRDInterface
	register   chan monitor.SiteSpec
	unregister chan monitor.SiteSpec
	namespace  string
	store      *registry
}

// NewWithClient creates a Watcher for the specified API client. When Watcher finds a created/removed URL, it sends
// the URL to the register/unregister channel respectively.
// If the namespace is specified, Watcher will only scan that namespace. Otherwise, all namespaces are scanned.
// Note that this needs RBAC setup to ensure the client can access those resources.
func NewWithClient(register, unregister chan monitor.SiteSpec, namespace string, client clientV1.TargetsCRDInterface) *Watcher {
	return &Watcher{
		Client:     client,
		register:   register,
		unregister: unregister,
		namespace:  namespace,
		store:      newRegistry(),
	}
}

// Run checks kubernetes for created/removed/modified URLs at the specified interval
func (watcher *Watcher) Run(ctx context.Context) {
	log.Info("watcher started")

	w := watcher.watch(ctx)

	// k8s may stop sending events after 30 minutes, so we renew it periodically
	ticker := time.NewTicker(25 * time.Minute)

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case event := <-w.ResultChan():
			watcher.processEvent(event)
		case <-ticker.C:
			w.Stop()
			w = watcher.watch(ctx)
			log.Debug("renewed custom resource watcher")
		}
	}

	w.Stop()
	ticker.Stop()

	log.Info("watcher stopped")
}

func (watcher *Watcher) watch(ctx context.Context) (w watch.Interface) {
	var err error
	w, err = watcher.Client.Targets(watcher.namespace).Watch(ctx, metav1.ListOptions{})

	if err != nil {
		log.WithError(err).Fatal("failed to set up custom resource watcher")
	}

	return
}

func (watcher *Watcher) processEvent(event watch.Event) {
	if event.Object == nil {
		log.WithField("type", event.Type).Warning("ignoring CRD event without an object")
		return
	}

	target := event.Object.(*v1.Target)
	log.WithFields(log.Fields{
		"name":      target.Name,
		"namespace": target.Namespace,
		"spec":      target.Spec,
		"event":     event.Type,
	}).Debug("event received")

	switch event.Type {
	case watch.Added:
		watcher.store.add(target.Namespace, target.Name, target.Spec)
		watcher.register <- monitor.SiteSpec{
			URL:  target.Spec.URL,
			Name: target.Spec.Name,
		}
	case watch.Deleted:
		spec := watcher.store.delete(target.Namespace, target.Name)
		watcher.unregister <- monitor.SiteSpec{URL: spec.URL}
	case watch.Modified:
		spec, ok := watcher.store.get(target.Namespace, target.Name)
		if ok && spec != target.Spec {
			watcher.store.add(target.Namespace, target.Name, target.Spec)
			watcher.unregister <- monitor.SiteSpec{URL: spec.URL}
			watcher.register <- monitor.SiteSpec{
				URL:  target.Spec.URL,
				Name: target.Spec.Name,
			}
		}
	}
}
