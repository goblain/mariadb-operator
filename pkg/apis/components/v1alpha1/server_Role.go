package v1alpha1

import (
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (mdbc *MariaDBCluster) ServerRoleTransform(r *rbac.Role) error {
	labels := mdbc.GetServerLabels()
	labels[MariaDBClusterNameLabel] = mdbc.Name

	r.SetName(mdbc.GetServerName())
	r.SetNamespace(mdbc.Namespace)
	r.SetLabels(labels)
	r.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(mdbc, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	r.Rules = nil
	r.Rules = append(r.Rules, rbac.PolicyRule{
		APIGroups: []string{"components.dsg.dk"},
		Resources: []string{"mariadbclusters"},
		Verbs:     []string{"get", "watch", "list", "patch", "update"},
	})
	return nil
}
