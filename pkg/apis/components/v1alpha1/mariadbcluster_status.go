package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PhaseBootstrapFirst        = "BootstrapFirst"
	PhaseBootstrapFirstRestart = "BootstrapFirstRestart"
	PhaseBootstrapSecond       = "BootstrapSecond"
	PhaseBootstrapThird        = "BootstrapThird"
	PhaseOperational           = "Operational"
	PhaseRecovery              = "Recovery"
	PhaseRecoverSeqNo          = "RecoverSeqNo"
	PhaseRecoveryReleaseAll    = "RecoveryReleaseAll"
	ConditionScaling           = "Scaling"
)

type MariaDBClusterCondition struct {
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time `json:"lastUpdateTime,omitempty"`
	Message            string      `json:"message,omitempty"`
	Reason             string      `json:"reason,omitempty"`
	Status             bool        `json:"status,omitempty"`
	Type               string      `json:"type,omitempty"`
}

type MariaDBClusterStatus struct {
	Phase                         string                    `json:"phase"`
	Stage                         string                    `json:"stage"`
	Conditions                    []MariaDBClusterCondition `json:"conditions"`
	CurrentVersion                string                    `json:"currentVersion"`
	TargetVersion                 string                    `json:"targetVersion"`
	StatefulSetObservedGeneration int64                     `json:"statefulSetObservedGeneration"`
	StatefulSetPodConditions      []PodCondition            `json:"statefulSetPodConditions"`
	BootstrapFrom                 string                    `json:"bootstrapFrom,omitempty"`
}

// PodCondition publishes grstate.dat values with some additional meta
type PodCondition struct {
	Hostname string
	Reported metav1.Time
	GRAState PodConditionGRAState
}

type PodConditionGRAState struct {
	Version         string `json:"version" yaml:"version"`
	UUID            string `json:"uuid" yaml:"uuid"`
	SeqNo           int64  `json:"seqno" yaml:"seqno"`
	SafeToBootstrap int    `json:"safe_to_bootstrap" yaml:"safe_to_bootstrap"`
}
