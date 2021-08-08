package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupName for the custom resource API
const GroupName = "webmon.clambin.private"

// GroupVersion for the custom resource API
const GroupVersion = "v1"

var schemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: GroupVersion}

var (
	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme adds the know types to the scheme
	AddToScheme = schemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schemeGroupVersion,
		&Target{},
		&TargetList{},
	)

	metav1.AddToGroupVersion(scheme, schemeGroupVersion)
	return nil
}
