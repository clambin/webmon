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
)

//go:generate controller-gen object paths=$GOFILE

// TargetSpec contains the fields within the "spec" entry of the custom resource
type TargetSpec struct {
	// URL of the site to monitor
	URL string `json:"url"`
}

// Target layout for the custom resource
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Target struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TargetSpec `json:"spec"`
}

// TargetList layout for a list of Target custom resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Target `json:"items"`
}
