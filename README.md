# webmon
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/clambin/webmon?color=green&label=Release&style=plastic)](https://github.com/clambin/webmon/releases)
[![Codecov](https://img.shields.io/codecov/c/gh/clambin/webmon?style=plastic)](https://app.codecov.io/gh/clambin/webmon)
[![Build](https://github.com/clambin/webmon/workflows/Build/badge.svg)](https://github.com/clambin/webmon/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/clambin/webmon)](https://goreportcard.com/report/github.com/clambin/webmon)
[![GitHub](https://img.shields.io/github/license/clambin/webmon?style=plastic)](https://github.com/clambin/webmon)

Monitors website update, latency & certificate expiry.

## Running

### Command line arguments

```
usage: webmon [<flags>] [<hosts>...]

webmon

Flags:
-h, --help            Show context-sensitive help (also try --help-long and --help-man).
-v, --version         Show application version.
--port=8080           Metrics listener port
--debug               Log debug messages
--interval=1m         Measurement interval
--watch               Watch k8s CRDs for target hosts
--watch.namespace=""  Namespace to watch for CRDs (default: all namespaces
--watch.kubeconfig=WATCH.KUBECONFIG  
~/.kube/config

Args:
[<hosts>]  hosts to ping
```

### Kubernetes 

When running in a Kubernetes cluster, sites to monitor can be provisioned through custom resources. 
To install these, apply the [crd.yml](assets/crd/crd.yml) file in this repo.  When RBAC is enabled in your cluster,
you will also need to apply [rbac.yml](assets/crd/rbac.yml).

Once the CRD is installed, add any site to monitor by created the following custom resource:

```
apiVersion: webmon.clambin.private/v1
kind: Target
metadata:
  name: <name>
  namespace: <namespace>
spec:
  url: https://your.url.here
```

## Metrics

Webmon exposes the following metrics to Prometheus:

```
* webmon_site_up: Set to 1 if the site is up
* webmon_site_latency_seconds: Time to check the site, in seconds
* webmon_certificate_expiry: Number of days before the HTTPS certificate expires
```

## Acknowledgements

* Martin Helmich's excellent [article](https://www.martin-helmich.de/en/blog/kubernetes-crd-client.html) on accessing Kubernetes CRDs in Go.

## Author

* **Christophe Lambin**

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
