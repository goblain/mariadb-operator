package v1alpha1

import (
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (cluster *MariaDBCluster) StatefulSetTransform(sset *apps.StatefulSet) error {
	pvars := GetPhaseVars(cluster)
	ssetName := cluster.GetServerName()
	serviceName := cluster.GetServerServiceName()
	labels := cluster.GetServerLabels()

	sset.SetName(ssetName)
	sset.SetNamespace(cluster.Namespace)
	sset.SetLabels(labels)
	sset.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(cluster, schema.GroupVersionKind{
			Group:   GroupName,
			Version: Version,
			Kind:    "MariaDBCluster",
		}),
	})
	sset.Spec.ServiceName = serviceName
	sset.Spec.Replicas = &pvars.Replicas
	sset.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	sset.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{Type: "RollingUpdate"}
	sset.Spec.PodManagementPolicy = apps.ParallelPodManagement
	sset.Spec.Template.ObjectMeta.Labels = labels
	sset.Spec.Template.Spec.ServiceAccountName = cluster.GetServerName()
	// InitContainers
	if len(sset.Spec.Template.Spec.InitContainers) < 1 {
		sset.Spec.Template.Spec.InitContainers = append(sset.Spec.Template.Spec.InitContainers, v1.Container{})
	}
	sset.Spec.Template.Spec.InitContainers[0].Name = "init"
	sset.Spec.Template.Spec.InitContainers[0].Image = "goblain/mdbc:dev"
	sset.Spec.Template.Spec.InitContainers[0].ImagePullPolicy = v1.PullAlways
	sset.Spec.Template.Spec.InitContainers[0].Command = []string{"/mdbc"}
	sset.Spec.Template.Spec.InitContainers[0].Args = []string{"init"}
	sset.Spec.Template.Spec.InitContainers[0].VolumeMounts = []v1.VolumeMount{
		v1.VolumeMount{Name: "data", MountPath: "/var/lib/mysql"},
	}

	// Containers
	if len(sset.Spec.Template.Spec.Containers) < 1 {
		sset.Spec.Template.Spec.Containers = append(sset.Spec.Template.Spec.Containers, v1.Container{})
	}
	if cluster.Status.Phase == PhaseBootstrapFirst {
		sset.Spec.Template.Spec.Containers[0].Args = []string{"--wsrep-new-cluster"}
		// } else if cluster.Status.Phase == PhaseRecovery {
		// 	sset.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sleep", "1d"}
		// 	sset.Spec.Template.Spec.Containers[0].Args = nil
	} else {
		sset.Spec.Template.Spec.Containers[0].Args = nil
		sset.Spec.Template.Spec.Containers[0].Command = nil
	}
	sset.Spec.Template.Spec.Containers[0].Name = "mariadb"
	// sset.Spec.Template.Spec.Containers[0].Image = "goblain/mdbc:dev"
	sset.Spec.Template.Spec.Containers[0].Image = "mariadb:10.2"
	// sset.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullIfNotPresent
	sset.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullAlways
	sset.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{
		v1.EnvVar{Name: "MYSQL_ALLOW_EMPTY_PASSWORD", Value: "yes"},
		v1.EnvVar{Name: "MYSQL_INITDB_SKIP_TZINFO", Value: "yes"},
	}
	sset.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		v1.VolumeMount{Name: "config", MountPath: "/etc/mysql/conf.d/operator.cnf", SubPath: "operator.cnf"},
		v1.VolumeMount{Name: "data", MountPath: "/var/lib/mysql"},
	}

	// if pvars.UseLivenessProbe {
	if sset.Spec.Template.Spec.Containers[0].LivenessProbe == nil {
		sset.Spec.Template.Spec.Containers[0].LivenessProbe = &v1.Probe{}
	}
	sset.Spec.Template.Spec.Containers[0].LivenessProbe.Handler = v1.Handler{
		Exec: &v1.ExecAction{Command: []string{"mysqladmin", "ping"}},
		// HTTPGet: &v1.HTTPGetAction{Port: intstr.FromInt(8080), Path: "/alive", Scheme: "HTTP"},
	}
	sset.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds = 30
	sset.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds = 5
	sset.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds = 2
	// } else {
	// 	sset.Spec.Template.Spec.Containers[0].LivenessProbe = nil
	// }
	// if pvars.UseReadinessProbe {
	if sset.Spec.Template.Spec.Containers[0].ReadinessProbe == nil {
		sset.Spec.Template.Spec.Containers[0].ReadinessProbe = &v1.Probe{}
	}
	sset.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler = v1.Handler{
		// bash -c "mysql --skip-column-names -e \"select variable_value from information_schema.global_status where variable_name='wsrep_local_state_comment'\" -B | grep Synced"
		Exec: &v1.ExecAction{Command: []string{"bash", "-c", "mysql --skip-column-names -e \"select variable_value from information_schema.global_status where variable_name='wsrep_local_state_comment'\" -B | grep -q Synced"}},
		// HTTPGet: &v1.HTTPGetAction{Port: intstr.FromInt(8080), Path: "/ready", Scheme: "HTTP"},
	}
	sset.Spec.Template.Spec.Containers[0].ReadinessProbe.InitialDelaySeconds = 10
	sset.Spec.Template.Spec.Containers[0].ReadinessProbe.PeriodSeconds = 2
	sset.Spec.Template.Spec.Containers[0].ReadinessProbe.TimeoutSeconds = 2
	// } else {
	// 	sset.Spec.Template.Spec.Containers[0].ReadinessProbe = nil
	// }
	sset.Spec.Template.Spec.Volumes = cluster.statefulSetVolumesTransform(sset.Spec.Template.Spec.Volumes)
	sset.Spec.VolumeClaimTemplates = cluster.statefulSetVolumeClaimTemplatesTransform(sset.Spec.VolumeClaimTemplates)
	return nil
}

func (mdbc *MariaDBCluster) statefulSetVolumeClaimTemplatesTransform(current []v1.PersistentVolumeClaim) []v1.PersistentVolumeClaim {
	if len(current) != 1 {
		current = make([]v1.PersistentVolumeClaim, 1)
	}
	expectedSpec := mdbc.Spec.Storages.Data.GetPersistentVolumeClaimSpecWithMode(v1.ReadWriteOnce)
	current[0].Name = "data"
	current[0].Spec.AccessModes = expectedSpec.AccessModes
	current[0].Spec.Resources = expectedSpec.Resources
	current[0].Spec.StorageClassName = expectedSpec.StorageClassName
	return current
}

func (mdbc *MariaDBCluster) statefulSetVolumesTransform(current []v1.Volume) []v1.Volume {
	if len(current) != 1 {
		current = make([]v1.Volume, 1)
		current[0].VolumeSource = v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: mdbc.GetServerConfigMapName()}}}
	}
	current[0].Name = "config"
	return current
}
