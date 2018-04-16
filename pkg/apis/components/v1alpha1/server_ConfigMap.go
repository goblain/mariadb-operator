package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (mdbc *MariaDBCluster) ServerConfigMapTransform(cmap *v1.ConfigMap) error {
	serviceName := mdbc.GetServerServiceName()
	configMapName := mdbc.GetServerConfigMapName()
	statefulSetName := mdbc.GetServerName()
	labels := mdbc.GetServerLabels()

	cmap.SetName(configMapName)
	cmap.SetNamespace(mdbc.Namespace)
	cmap.SetLabels(labels)
	cmap.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(mdbc, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	var wsrep []string
	if mdbc.Status.Phase == PhaseBootstrapFirst || mdbc.Status.Phase == PhaseBootstrapFirstRestart {
		wsrep = []string{}
	} else if mdbc.Status.Phase == PhaseBootstrapSecond {
		wsrep = []string{statefulSetName + "-0." + serviceName}
	} else if mdbc.Status.Phase == PhaseBootstrapThird {
		wsrep = []string{statefulSetName + "-0." + serviceName, statefulSetName + "-1." + serviceName}
	} else {
		wsrep = []string{statefulSetName + "-0." + serviceName, statefulSetName + "-1." + serviceName, statefulSetName + "-2." + serviceName}
	}
	mdbConfig := &MariaDBConfig{WSREPEndpoints: wsrep}

	operatorCnf, err := mdbConfig.Render()
	if err != nil {
		return err
	}
	cmap.Data = map[string]string{
		"operator.cnf": operatorCnf,
		"user.cnf":     mdbc.Spec.ServerConfig}
	return nil
}
