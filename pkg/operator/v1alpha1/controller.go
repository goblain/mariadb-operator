package operator

import (
	"time"

	"github.com/Sirupsen/logrus"
	api "github.com/goblain/mariadb-operator/pkg/apis/mariadb/v1alpha1"
	"k8s.io/client-go/rest"
)

// import (
// 	"time"

// 	"github.com/Sirupsen/logrus"
// 	api "github.com/goblain/mariadb-operator/pkg/apis/mariadb/v1alpha1"
// 	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/runtime/schema"
// 	"k8s.io/client-go/rest"
// )

// var (
// 	ShemeBuilder       = runtime.NewSchemeBuilder(addKnownTypes)
// 	SchemeGroupVersion = schema.GroupVersion{Group: api.GroupName, Version: api.Version}
// )

type Controller struct {
	ClientConfig *rest.Config
}

func (c *Controller) Run() {
	// 	c.init()
	// 	// kcs := kubernetes.NewForConfigOrDie(c.ClientConfig)
	// 	// mdbcs := mdbclientset.NewForConfigOrDie(c.ClientConfig)
	for {
		time.Sleep(time.Second * 5)
		logrus.Infof("[" + api.GroupName + "/" + api.Version + "] Tick")
	}
}

// func (c *Controller) init() {
// 	EnsureCRD(&apiextensionsv1beta1.CustomResourceDefinition{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: api.CRDName,
// 		},
// 		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
// 			Group:   SchemeGroupVersion.Group,
// 			Version: SchemeGroupVersion.Version,
// 			Scope:   apiextensionsv1beta1.NamespaceScoped,
// 			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
// 				Plural:     api.ResourcePlural,
// 				Kind:       api.ResourceKind,
// 				ShortNames: []string{api.ResourceShorts},
// 			},
// 		},
// 	})
// }

// // func addKnownTypes(s *runtime.Scheme) error {
// // 	s.AddKnownTypes(SchemeGroupVersion, &api.MariaDBCluster{})
// // 	metav1.AddToGroupVersion(s, SchemeGroupVersion)
// // 	return nil
// // }
