package v1alpha1

import (
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (cluster *MariaDBCluster) ProxyDeploymentTransform(obj *apps.Deployment) error {
	// pvars := GetPhaseVars(cluster)
	name := cluster.GetProxyName()
	labels := cluster.GetProxyLabels()

	obj.SetName(name)
	obj.SetNamespace(cluster.Namespace)
	obj.SetLabels(labels)
	obj.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(cluster, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	replicas := int32(2)
	obj.Spec.Replicas = &replicas
	obj.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	obj.Spec.Template.ObjectMeta.Labels = labels
	if len(obj.Spec.Template.Spec.Containers) < 1 {
		obj.Spec.Template.Spec.Containers = append(obj.Spec.Template.Spec.Containers, v1.Container{})
	}
	if cluster.Status.Phase == PhaseBootstrapFirst {
		obj.Spec.Template.Spec.Containers[0].Args = []string{"--wsrep-new-cluster"}
	} else {
		obj.Spec.Template.Spec.Containers[0].Args = nil
	}
	obj.Spec.Template.Spec.Containers[0].Name = "proxysql"
	obj.Spec.Template.Spec.Containers[0].Image = "proxysql"
	obj.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullIfNotPresent
	// obj.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{
	// 	v1.EnvVar{Name: "MYSQL_ALLOW_EMPTY_PASSWORD", Value: "yes"},
	// 	v1.EnvVar{Name: "MYSQL_INITDB_SKIP_TZINFO", Value: "yes"},
	// }
	// obj.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
	// 	v1.VolumeMount{Name: "config", MountPath: "/etc/mysql/conf.d/operator.cnf", SubPath: "operator.cnf"},
	// 	v1.VolumeMount{Name: "data", MountPath: "/var/lib/mysql"},
	// }
	// if obj.Spec.Template.Spec.Containers[0].LivenessProbe == nil {
	// 	obj.Spec.Template.Spec.Containers[0].LivenessProbe = &v1.Probe{}
	// }
	// obj.Spec.Template.Spec.Containers[0].LivenessProbe.Handler = v1.Handler{
	// 	Exec: &v1.ExecAction{Command: []string{"mysqladmin", "ping"}},
	// }
	// obj.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds = 30
	// obj.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds = 5
	// obj.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds = 2
	// if obj.Spec.Template.Spec.Containers[0].ReadinessProbe == nil {
	// 	obj.Spec.Template.Spec.Containers[0].ReadinessProbe = &v1.Probe{}
	// }
	// obj.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler = v1.Handler{
	// 	Exec: &v1.ExecAction{Command: []string{"bash", "-c", "mysql --skip-column-names -e \"select variable_value from information_schema.global_status where variable_name='wsrep_local_state_comment'\" -B | grep -q Synced"}},
	// }
	// obj.Spec.Template.Spec.Containers[0].ReadinessProbe.InitialDelaySeconds = 10
	// obj.Spec.Template.Spec.Containers[0].ReadinessProbe.PeriodSeconds = 2
	// obj.Spec.Template.Spec.Containers[0].ReadinessProbe.TimeoutSeconds = 2
	// obj.Spec.Template.Spec.Volumes = cluster.proxySetVolumesTransform(obj.Spec.Template.Spec.Volumes)
	return nil
}

func (mdbc *MariaDBCluster) proxySetVolumesTransform(current []v1.Volume) []v1.Volume {
	if len(current) != 1 {
		current = make([]v1.Volume, 1)
		current[0].VolumeSource = v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: mdbc.GetServerConfigMapName()}}}
	}
	current[0].Name = "config"
	return current
}
