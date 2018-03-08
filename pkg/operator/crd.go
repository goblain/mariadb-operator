package operator

import (
	// 	"fmt"

	// 	api "github.com/goblain/mariadb-operator/pkg/apis/mariadb/v1alpha1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	// 	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// 	"k8s.io/client-go/rest"
	mariadbv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/mariadb/v1alpha1"
)

func (op *Operator) EnsureSupportedCRDs() {
	crds := mariadbv1alpha1.GetCRDs()
	for _, crd := range crds {
		_, err := op.Client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
		if err != nil && !IsKubernetesResourceAlreadyExistError(err) {
			panic(err)
		}
		return nil
	}
}

func (op *Operator) EnsureCRD(crd apiextensionsv1beta1.CustomResourceDefinition) {

}

// func CreateCRD(clientset apiextensionsclientset.Interface, crd apiextensionsv1beta1.CustomResourceDefinition) error {
// 	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
// 	return err
// }

// func WaitCRDReady(clientset *rest.Interface) error {
// 	// err := retryutil.Retry(5*time.Second, 20, func() (bool, error) {
// 	crd, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(api.CRDName, metav1.GetOptions{})
// 	if err != nil {
// 		panic(err)
// 		// return false, err
// 	}
// 	for _, cond := range crd.Status.Conditions {
// 		switch cond.Type {
// 		case apiextensionsv1beta1.Established:
// 			if cond.Status == apiextensionsv1beta1.ConditionTrue {
// 				return true, nil
// 			}
// 		case apiextensionsv1beta1.NamesAccepted:
// 			if cond.Status == apiextensionsv1beta1.ConditionFalse {
// 				return false, fmt.Errorf("Name conflict: %v", cond.Reason)
// 			}
// 		}
// 	}
// 	// return false, nil
// 	// })
// 	// if err != nil {
// 	// 	return fmt.Errorf("wait CRD created failed: %v", err)
// 	// }
// 	return nil
// }
