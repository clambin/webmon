// Package v1 data structures and supporting functions for the Target CRD
//
// V1 layout is as follows:
//
//   apiVersion: webmon.clambin.private/v1
//   kind: Target
//   metadata:
//     name: <name>
//     namespace: <namespace>
//   spec:
//     url: https://example.com
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TargetSpec contains the fields within the "spec" entry of the custom resource
type TargetSpec struct {
	URL string `json:"url"`
}

// Target layout for the custom resource
type Target struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TargetSpec `json:"spec"`
}

// TargetList layout for a list of Target custom resources
type TargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Target `json:"items"`
}

// GetObjectKind returns the object's kind.  Needed to cast runtime.Object to Target in watch event handler.
func (in *Target) GetObjectKind() schema.ObjectKind {
	return in.TypeMeta.GetObjectKind()
}
