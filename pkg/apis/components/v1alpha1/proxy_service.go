package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (mdbc *MariaDBCluster) ProxyServiceTransform(svc *v1.Service) error {
	labels := mdbc.GetProxyLabels()
	svc.SetName(mdbc.GetProxyServiceName())
	svc.SetNamespace(mdbc.Namespace)
	svc.SetLabels(labels)
	svc.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(mdbc, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	svc.Spec.Type = v1.ServiceTypeClusterIP
	if mdbc.isProxyEnabled() {
		svc.Spec.Selector = labels
	} else {
		svc.Spec.Selector = mdbc.GetServerLabels()
	}
	svc.Spec.Ports = []v1.ServicePort{
		v1.ServicePort{
			Name:       "mysql",
			Protocol:   v1.ProtocolTCP,
			Port:       3306,
			TargetPort: intstr.FromInt(3306),
		},
	}
	return nil
}
