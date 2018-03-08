package v1alpha1

import (
	"k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ()

// Expose all matching CRDs programaticaly to other elements of the system but keep them defined in the API it self
// TODO: to be tested for crd collisions!!!

func GetCRDs() []*apiextensionsv1beta1.CustomResourceDefinition {
	mariadbcluster := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "MariaDBCluster"},
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
	Status            MariaDBClusterStatus `json:"status"`
}

//	GetObjectKind() schema.ObjectKind
//	DeepCopyObject() Object

type MariaDBClusterSpec struct {
	// MariaDB container/engine version, no less then 10.2.8
	Version string `json:"version"`
	// Pause any control from operator on this resource
	Paused        bool                    `json:"paused"`
	Size          int                     `json:"size"`
	ConfigMapName string                  `json:"configMapName"`
	Resources     v1.ResourceRequirements `json:"resources"`
	Storages      Storages                `json:"storages"`
	// Notifications
	//   slack
	//   email
}

type Storages struct {
	DataStorage     Storage `json:"data"`
	SnapshotStorage Storage `json:"snapshot"`
}

type Storage struct {
	StorageClassName string `json:"storageClassName"`
	InitialSize      string `json:"initSize"`
	MaximumSize      string `json:"maxSize"`
	GrowBy           string `json:"growBy"`
	GrowThreshold    string `json:"growThreshold"`
	RetentionPolicy  string //keep data after cluster deleted ?
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
