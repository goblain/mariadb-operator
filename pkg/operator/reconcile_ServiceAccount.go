package operator

import (
	"reflect"

	"github.com/Sirupsen/logrus"
	componentsv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	"github.com/goblain/mariadb-operator/pkg/util"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (o *Operator) reconcileServiceAccount(mdbc *componentsv1alpha1.MariaDBCluster, name string, transformer func(*v1.ServiceAccount) error) error {
	logger := util.GetClusterLogger(mdbc).WithField("kind", "ServiceAccount").WithField("action", "reconcile").WithField("name", name)
	logger.WithField("event", "started").Debug()
	defer logger.WithField("event", "finished").Debug()
	current, err := o.Client.CoreV1().ServiceAccounts(mdbc.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.WithField("event", "NotFound").Debug("not found in cluster")
			expected := &v1.ServiceAccount{}
			transformer(expected)
			_, err = o.Client.CoreV1().ServiceAccounts(mdbc.Namespace).Create(expected)
			if err != nil {
				logger.Errorf("Creation failed with : %s", err.Error())
				return err
			} else {
				logger.WithField("event", "created").Info()
				return nil
			}
		} else {
			logger.Errorf("Error fetching object : %s", err.Error())
			return err
		}
	} else {
		expected := current.DeepCopy()
		transformer(expected)
		checkAndPatchServiceAccount(current, expected, o.Client.Core(), logger)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	}
}

func (o *Operator) reconcileServerServiceAccount(mdbc *componentsv1alpha1.MariaDBCluster) error {
	return o.reconcileServiceAccount(mdbc, mdbc.GetServerName(), mdbc.ServerServiceAccountTransform)
}

func checkAndPatchServiceAccount(current, expected *v1.ServiceAccount, client clientcorev1.CoreV1Interface, logger *logrus.Entry) (bool, error) {
	// if !reflect.DeepEqual(expected, current) {
	// 	logger.WithField("event", "change").Info("changes detected")
	// 	patchBytes, _ := patchGen(current, expected, appsv1.StatefulSet{})
	// 	logger.Debugf(string(patchBytes))
	// 	// TODO : error handling
	// 	_, err := client.StatefulSets(expected.Namespace).Patch(expected.Name, types.StrategicMergePatchType, patchBytes)
	// 	if err != nil {
	// 		logger.Error(err.Error())
	// 	}
	// 	return true, nil
	// } else {
	// 	logger.WithField("event", "nochange").Info("no changes")
	// }
	// return false, nil

	if !reflect.DeepEqual(expected, current) {
		logger.Info("Spec differs between current and expected, updating")
		// TODO : Switch to Patch as Update fails due to immutable fields
		_, err := client.ServiceAccounts(expected.Namespace).Update(expected)
		return true, err
	}
	return false, nil
}
