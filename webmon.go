package main

import (
	"context"
	clientV1 "github.com/clambin/webmon/crds/targets/clientset/v1"
	"github.com/clambin/webmon/monitor"
	"github.com/clambin/webmon/utils"
	"github.com/clambin/webmon/version"
	"github.com/clambin/webmon/watcher"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	go func() {
		listenAddress := ":8080"
		r := mux.NewRouter()
		r.Use(prometheusMiddleware)
		r.Path("/metrics").Handler(promhttp.Handler())
		r.Path("/health").Handler(http.HandlerFunc(myMonitor.Health))
		err = http.ListenAndServe(listenAddress, r)
		if err != nil {
			log.WithError(err).Fatal("unable to start metrics server")
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

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

// Prometheus metrics
var (
	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"path"})
)

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}
