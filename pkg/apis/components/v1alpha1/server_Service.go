package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (mdbc *MariaDBCluster) ServerServiceTransform(svc *v1.Service) error {
	labels := mdbc.GetServerLabels()
	labels[MariaDBClusterNameLabel] = mdbc.Name

	svc.SetName(mdbc.GetServerServiceName())
	svc.SetNamespace(mdbc.Namespace)
	svc.SetLabels(labels)
	svc.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(mdbc, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	annotations := make(map[string]string)
	annotations["service.alpha.kubernetes.io/tolerate-unready-endpoints"] = "true"
	svc.SetAnnotations(annotations)
	svc.Spec.Type = v1.ServiceTypeClusterIP
	svc.Spec.ClusterIP = "None"
	svc.Spec.Selector = labels
	svc.Spec.PublishNotReadyAddresses = true
	svc.Spec.Ports = []v1.ServicePort{
		v1.ServicePort{
			Name:       "mysql",
			Protocol:   v1.ProtocolTCP,
			Port:       3306,
			TargetPort: intstr.FromInt(3306),
		},
		v1.ServicePort{
			Name:       "wsrep",
			Protocol:   v1.ProtocolTCP,
			Port:       4567,
			TargetPort: intstr.FromInt(4567),
		},
	}
	return nil
}
