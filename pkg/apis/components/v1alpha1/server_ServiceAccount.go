package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (mdbc *MariaDBCluster) ServerServiceAccountTransform(sa *v1.ServiceAccount) error {
	labels := mdbc.GetServerLabels()
	labels[MariaDBClusterNameLabel] = mdbc.Name

	sa.SetName(mdbc.GetServerName())
	sa.SetNamespace(mdbc.Namespace)
	sa.SetLabels(labels)
	sa.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(mdbc, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	return nil
}
