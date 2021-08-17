package watcher_test

import (
	"context"
	v1 "github.com/clambin/webmon/crds/targets/api/types/v1"
	"github.com/clambin/webmon/crds/targets/clientset/v1/mock"
	"github.com/clambin/webmon/monitor"
	"github.com/clambin/webmon/watcher"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestWatcher_Run(t *testing.T) {
	client := mock.New()

	register := make(chan monitor.SiteSpec)
	unregister := make(chan monitor.SiteSpec)

	w := watcher.NewWithClient(register, unregister, "", client)

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		w.Run(ctx)
		wg.Done()
	}()

	client.Add("foo", "bar", v1.TargetSpec{URL: "https://example.com"})
	site := <-register
	assert.Equal(t, "https://example.com", site.URL)

	client.Modify("foo", "bar", v1.TargetSpec{URL: "https://example.com:443"})
	site = <-unregister
	assert.Equal(t, "https://example.com", site.URL)
	site = <-register
	assert.Equal(t, "https://example.com:443", site.URL)

	client.Delete("foo", "bar")
	site = <-unregister
	assert.Equal(t, "https://example.com:443", site.URL)

	cancel()

	wg.Wait()
}
