package main

import (
	"context"
	"github.com/clambin/webmon/version"
	"github.com/clambin/webmon/webmon"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	port     int
	debug    bool
	interval time.Duration
)

func main() {
	a := kingpin.New(filepath.Base(os.Args[0]), "webmon")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("port", "Metrics listener port").Default("8080").IntVar(&port)
	a.Flag("debug", "Log debug messages").BoolVar(&debug)
	a.Flag("interval", "Measurement interval").Default("1m").DurationVar(&interval)
	hosts := a.Arg("hosts", "hosts to ping").Strings()

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	if len(*hosts) == 0 {
		log.Error("No hosts specified. Aborting")
		os.Exit(2)
	}

	*hosts = webmon.Unique(*hosts)

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	log.WithField("hosts", *hosts).Infof("webmon %s", version.BuildVersion)

	var monitor *webmon.Monitor
	monitor, err = webmon.New(*hosts)
	prometheus.MustRegister(monitor)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err2 := monitor.Run(ctx, interval)
		if err2 != nil {
			log.WithError(err2).Fatal("could not start webmon")
		}
	}()

	log.Info("webmon started")
	listenAddress := ":8080"
	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	err = http.ListenAndServe(listenAddress, r)
	if err != nil {
		log.WithError(err).Fatal("unable to start metrics server")
	}
	log.Info("webmon stopped")
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
