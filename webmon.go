package main

import (
	"context"
	"github.com/clambin/gotools/metrics"
	clientV1 "github.com/clambin/webmon/crds/targets/clientset/v1"
	"github.com/clambin/webmon/monitor"
	"github.com/clambin/webmon/utils"
	"github.com/clambin/webmon/version"
	"github.com/clambin/webmon/watcher"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var (
	port           int
	debug          bool
	interval       time.Duration
	watch          bool
	watchNamespace string
	kubeconfig     string
)

func main() {
	a := kingpin.New(filepath.Base(os.Args[0]), "webmon")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("port", "Metrics listener port").Default("8080").IntVar(&port)
	a.Flag("debug", "Log debug messages").BoolVar(&debug)
	a.Flag("interval", "Measurement interval").Default("1m").DurationVar(&interval)
	a.Flag("watch", "Watch k8s CRDs for target hosts").BoolVar(&watch)
	a.Flag("watch.namespace", "Namespace to watch for CRDs (default: all namespaces)").Default("").StringVar(&watchNamespace)
	a.Flag("watch.kubeconfig", "~/.kube/config").StringVar(&kubeconfig)
	hosts := a.Arg("hosts", "hosts to ping").Strings()

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	if len(*hosts) == 0 && !watch {
		log.Error("No hosts specified. Aborting")
		os.Exit(2)
	}

	*hosts = utils.Unique(*hosts)

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	log.WithField("hosts", *hosts).Infof("monitor %s", version.BuildVersion)

	myMonitor := monitor.New(nil)
	prometheus.MustRegister(myMonitor)

	ctx, cancel := context.WithCancel(context.Background())

	// run the monitor
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err2 := myMonitor.Run(ctx, interval)
		if err2 != nil {
			log.WithError(err2).Fatal("could not start monitor")
		}
		wg.Done()
	}()

	// Register hosts
	for _, host := range *hosts {
		myMonitor.Register <- monitor.SiteSpec{URL: host}
	}

	if watch {
		var myWatcher *watcher.Watcher
		myWatcher, err = newWatcher(myMonitor, watchNamespace)

		if err != nil {
			log.WithError(err).Fatal("unable to start watcher")
		}

		wg.Add(1)
		go func() {
			myWatcher.Run(ctx)
			wg.Done()
		}()
	}

	promServer := metrics.NewServerWithHandlers(8080, []metrics.Handler{
		{Path: "/health", Handler: http.HandlerFunc(myMonitor.Health)},
	})

	go func() {
		log.Info("prometheus metrics server started")
		err2 := promServer.Run()
		if err2 != http.ErrServerClosed {
			log.WithError(err2).Fatal("unable to start metrics server")
		}
		log.Info("prometheus metrics server stopped")
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	_ = promServer.Shutdown(30 * time.Second)
	cancel()
	wg.Wait()
	log.Info("webmon stopped")
}

func newWatcher(monitor *monitor.Monitor, namespace string) (w *watcher.Watcher, err error) {
	var config *rest.Config
	if kubeconfig == "" {
		log.Info("using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.Infof("using configuration from '%s'", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	var client *clientV1.TargetsCRDClient
	if err == nil {
		client, err = clientV1.NewForConfig(config)
	}

	if err != nil {
		return
	}

	return watcher.NewWithClient(monitor.Register, monitor.Unregister, namespace, client), nil
}
