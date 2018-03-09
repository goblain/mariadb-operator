package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	ResourceShorts = [...]string{"mdb", "mdbs"}
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion, &MariaDBCluster{}, &MariaDBClusterList{})
	metav1.AddToGroupVersion(s, SchemeGroupVersion)
	return nil
}
