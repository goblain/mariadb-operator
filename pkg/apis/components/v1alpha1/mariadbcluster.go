package v1alpha1

import (
	"k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	MariaDBClusterLabelPrefix string = "mariadbcluster.components.dsg.dk/"
	MariaDBClusterNameLabel   string = MariaDBClusterLabelPrefix + "cluster-name"
	MariaDBClusterRoleLabel   string = MariaDBClusterLabelPrefix + "role"

	MariaDBClusterServerRole string = "server"
	MariaDBClusterProxyRole  string = "proxy"
)

var ()

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Expose all matching CRDs programaticaly to other elements of the system but keep them defined in the API it self
// TODO: to be tested for crd collisions !!!

func GetCRDs() []*apiextensionsv1beta1.CustomResourceDefinition {
	mariadbcluster := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: CRDName},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   GroupName,
			Version: Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: ResourcePlural,
				Kind:   ResourceKind,
			},
		},
	}
	return []*apiextensionsv1beta1.CustomResourceDefinition{mariadbcluster}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MariaDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBCluster `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MariaDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MariaDBClusterSpec   `json:"spec"`
	Status            MariaDBClusterStatus `json:"status,omitempty"`
}

type MariaDBClusterSpec struct {
	// MariaDB container/engine version, no less then 10.2.8
	Version string `json:"version"`
	// Pause any control from operator on this resource
	Paused        bool                    `json:"paused"`
	Replicas      int32                   `json:"replicas"`
	ConfigMapName string                  `json:"configMapName"`
	Resources     v1.ResourceRequirements `json:"resources"`
	Storages      Storages                `json:"storages"`
	ServerConfig  string                  `json:"serverConfig"`
	Proxy         bool                    `json:"proxy"`
	// Notifications
	//   slack
	//   email
}

type Storages struct {
	Data     Storage `json:"data,omitempty"`
	Snapshot Storage `json:"snapshot,omitempty"`
}

type Storage struct {
	StorageClassName string `json:"storageClassName"`
	InitialSize      string `json:"initSize"`
	MaximumSize      string `json:"maxSize"`
	GrowBy           string `json:"growBy"`
	GrowThreshold    string `json:"growThreshold"`
	RetentionPolicy  string //keep data after cluster deleted ?
}

func (s *Storage) GetResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: map[v1.ResourceName]resource.Quantity{"storage": resource.MustParse(s.InitialSize)},
	}
}

func (s *Storage) GetPersistentVolumeClaimSpecWithMode(mode v1.PersistentVolumeAccessMode) v1.PersistentVolumeClaimSpec {
	return v1.PersistentVolumeClaimSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{mode},
		// Selector:
		Resources: s.GetResourceRequirements(),
		// VolumeName:
		StorageClassName: &s.StorageClassName,
	}
}

func (mdbc *MariaDBCluster) GetWSREPEndpoints() []string {
	var wsrep []string

	statefulSetName := mdbc.GetServerName()
	serviceName := mdbc.GetServerServiceName()

	if mdbc.Status.Phase == PhaseBootstrapFirst || mdbc.Status.Phase == PhaseBootstrapFirstRestart {
		wsrep = []string{}
	} else if mdbc.Status.Phase == PhaseBootstrapSecond {
		wsrep = []string{statefulSetName + "-0." + serviceName}
	} else if mdbc.Status.Phase == PhaseBootstrapThird {
		wsrep = []string{statefulSetName + "-0." + serviceName, statefulSetName + "-1." + serviceName}
	} else {
		wsrep = []string{statefulSetName + "-0." + serviceName, statefulSetName + "-1." + serviceName, statefulSetName + "-2." + serviceName}
	}
	return wsrep
}

func (mdbc *MariaDBCluster) GetSnapshotPVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mdbc.Name,
			Namespace: mdbc.Namespace,
			Labels: map[string]string{
				"app": mdbc.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mdbc, schema.GroupVersionKind{
					Group:   GroupName,
					Version: Version,
					Kind:    "MariaDBCluster",
				}),
			},
		},
		Spec: mdbc.Spec.Storages.Snapshot.GetPersistentVolumeClaimSpecWithMode(v1.ReadWriteMany),
	}
}

func (mdb *MariaDBCluster) SetDefaults() {

}

func (mdb *MariaDBCluster) Validate() error {
	return nil
}

func (mdb *MariaDBCluster) AsOwner() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ResourceKind,
		Name:       mdb.Name,
		UID:        mdb.UID,
		Controller: &trueVar,
	}
}

// Name getters

func (mdbc *MariaDBCluster) GetServerName() string {
	return mdbc.Name + "-" + MariaDBClusterServerRole
}

func (mdbc *MariaDBCluster) GetServerServiceName() string {
	return mdbc.GetServerName()
}

func (mdbc *MariaDBCluster) GetServerConfigMapName() string {
	return mdbc.GetServerName()
}

func (mdbc *MariaDBCluster) GetProxyName() string {
	return mdbc.Name + "-" + MariaDBClusterProxyRole
}

func (mdbc *MariaDBCluster) GetProxyServiceName() string {
	return mdbc.Name
}

func (mdbc *MariaDBCluster) GetProxyConfigMapName() string {
	return mdbc.GetProxyName()
}

func (mdbc *MariaDBCluster) isProxyEnabled() bool {
	return mdbc.Spec.Proxy
}

// Label getters

func (mdbc *MariaDBCluster) GetServerLabels() map[string]string {
	labels := make(map[string]string)
	labels[MariaDBClusterNameLabel] = mdbc.Name
	labels[MariaDBClusterRoleLabel] = MariaDBClusterServerRole
	return labels
}

func (mdbc *MariaDBCluster) GetProxyLabels() map[string]string {
	labels := make(map[string]string)
	labels[MariaDBClusterNameLabel] = mdbc.Name
	labels[MariaDBClusterRoleLabel] = MariaDBClusterProxyRole
	return labels
}
