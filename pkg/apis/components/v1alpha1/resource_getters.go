package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PhaseVars struct {
	Replicas          int32 `json:"replicas"`
	UseReadinessProbe bool  `json:"useReadinessProbe"`
	UseLivenessProbe  bool  `json:"useLivenessProbe"`
}

func GetPhaseVars(cluster *MariaDBCluster) *PhaseVars {
	var replicas int32
	useReadinessProbe := true
	useLivenessProbe := true
	replicas = cluster.Spec.Replicas
	if cluster.Status.Phase == PhaseBootstrapFirst || cluster.Status.Phase == PhaseBootstrapFirstRestart {
		replicas = int32(1)
	} else if cluster.Status.Phase == PhaseBootstrapSecond {
		replicas = int32(2)
	} else if cluster.Status.Phase == PhaseBootstrapThird {
		replicas = int32(3)
	} else if cluster.Status.Phase == PhaseRecovery {
		useReadinessProbe = false
		useLivenessProbe = false
	}
	vars := &PhaseVars{
		Replicas:          replicas,
		UseReadinessProbe: useReadinessProbe,
		UseLivenessProbe:  useLivenessProbe,
	}
	// jvars, _ := json.Marshal(vars)
	// cluster.Logger.WithField("action", "phaseVars").Debugf("%s", jvars)
	return vars
}

func GetSnapshotPersistentVolumeClaim(cluster *MariaDBCluster) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-snapshot",
			Namespace: cluster.Namespace,
		},
		Spec: cluster.Spec.Storages.Snapshot.GetPersistentVolumeClaimSpecWithMode(v1.ReadWriteMany),
	}
	return pvc
}
