package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	MariaDBClusterResourceKind   = "MariaDBCluster"
	MariaDBClusterResourcePlural = "mariadbclusters"
	groupName                    = "components.dsg.dk"
	version                      = "v1alpha1"
)

var (
	SchemeBuilder = runtime.SchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme

	SchemeGroupVersion    = schema.GroupVersion{Group: groupName, Version: version}
	MariaDBClusterCRDName = MariaDBClusterResourcePlural + "." + groupName
)

func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion, &MariaDBCluster{}, &MariaDBClusterList{})
	metav1.AddToGroupVersion(s, SchemeGroupVersion)
	return nil
}
