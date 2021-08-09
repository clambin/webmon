// Package v1 Client for Target CRD
package v1

import (
	"github.com/clambin/webmon/crds/targets/api/types/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// TargetsCRDInterface interface for a TargetsCRDClient
type TargetsCRDInterface interface {
	Targets(namespace string) TargetCRDInterface
}

// TargetsCRDClient structure for a V1 version of the WebMon CRD API client
type TargetsCRDClient struct {
	restClient *rest.RESTClient
}

// NewForConfig creates a new API client for the provided REST configuration
func NewForConfig(c *rest.Config) (client *TargetsCRDClient, err error) {
	err = v1.AddToScheme(scheme.Scheme)

	if err != nil {
		return
	}

	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1.GroupName, Version: v1.GroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion() //serializer.NewCodecFactory(scheme.Scheme)
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	var restClient *rest.RESTClient
	restClient, err = rest.RESTClientFor(&config)

	if err == nil {
		client = &TargetsCRDClient{
			restClient: restClient,
		}
	}

	if err != nil {
		log.WithError(err).Error("unable to create k8s api client")
	}

	return
}

// Targets gives access to the API endpoints for the specified namespace.  If namespace is blank, all namespaces are considered
func (c *TargetsCRDClient) Targets(namespace string) TargetCRDInterface {
	return &targetCRDClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
