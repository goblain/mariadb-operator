package operator

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	mariadbv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (op *Operator) EnsureSupportedCRDs() error {
	crds := mariadbv1alpha1.GetCRDs()
	for _, crd := range crds {
		_, err := op.ApiExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
		if apierrors.IsAlreadyExists(err) {
			logrus.Info("CRD already exists, not creating but ok to pass")
		} else if err != nil {
			// if err != nil {
			panic(err)
		}
		op.WaitCRDReady(mariadbv1alpha1.CRDName)
		return nil
	}
	return nil
}

func (op *Operator) WaitCRDReady(name string) error {
	// err := retryutil.Retry(5*time.Second, 20, func() (bool, error) {
	crd, err := op.ApiExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
	if err != nil {
		panic(err)
		// return false, err
	}
	for _, cond := range crd.Status.Conditions {
		switch cond.Type {
		case apiextensionsv1beta1.Established:
			if cond.Status == apiextensionsv1beta1.ConditionTrue {
				// return true, nil
				return nil
			}
		case apiextensionsv1beta1.NamesAccepted:
			if cond.Status == apiextensionsv1beta1.ConditionFalse {
				// return false, fmt.Errorf("Name conflict: %v", cond.Reason)
				return fmt.Errorf("Name conflict: %v", cond.Reason)
			}
		}
	}
	// return false, nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("wait CRD created failed: %v", err)
	// }
	return nil
}
