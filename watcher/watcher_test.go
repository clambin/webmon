package watcher_test

import (
	"context"
	"github.com/clambin/webmon/crds/targets/clientset/v1/mock"
	"github.com/clambin/webmon/watcher"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestWatcher_Run(t *testing.T) {
	client := mock.New()

	register := make(chan string)
	unregister := make(chan string)

	w := watcher.NewWithClient(register, unregister, "", client)

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		w.Run(ctx)
		wg.Done()
	}()

	client.Add("foo", "bar", "https://example.com")
	url := <-register
	assert.Equal(t, "https://example.com", url)

	client.Modify("foo", "bar", "https://example.com:443")
	url = <-unregister
	assert.Equal(t, "https://example.com", url)
	url = <-register
	assert.Equal(t, "https://example.com:443", url)

	client.Delete("foo", "bar")
	url = <-unregister
	assert.Equal(t, "https://example.com:443", url)

	cancel()

	wg.Wait()
}
