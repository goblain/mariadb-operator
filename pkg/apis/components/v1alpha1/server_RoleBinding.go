package v1alpha1

import (
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (mdbc *MariaDBCluster) ServerRoleBindingTransform(rb *rbac.RoleBinding) error {
	labels := mdbc.GetServerLabels()
	labels[MariaDBClusterNameLabel] = mdbc.Name

	rb.SetName(mdbc.GetServerName())
	rb.SetNamespace(mdbc.Namespace)
	rb.SetLabels(labels)
	rb.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(mdbc, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	rb.Subjects = append(rb.Subjects, rbac.Subject{Kind: rbac.ServiceAccountKind, Name: mdbc.GetServerName(), Namespace: mdbc.Namespace})
	rb.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	rb.RoleRef.Kind = "Role"
	rb.RoleRef.Name = mdbc.GetServerName()
	return nil
}
