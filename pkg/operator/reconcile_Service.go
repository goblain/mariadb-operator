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

func (o *Operator) reconcileService(mdbc *componentsv1alpha1.MariaDBCluster, serviceName string, transformer func(*v1.Service) error) error {
	logger := util.GetClusterLogger(mdbc).WithField("kind", "Service").WithField("action", "reconcile").WithField("name", serviceName)
	logger.WithField("event", "started").Debug()
	defer logger.WithField("event", "finished").Debug()
	current, err := o.Client.CoreV1().Services(mdbc.Namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.WithField("event", "NotFound").Debug("not found in cluster")
			expected := &v1.Service{}
			transformer(expected)
			_, err = o.Client.CoreV1().Services(mdbc.Namespace).Create(expected)
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
		checkAndPatchService(current, expected, o.Client.Core(), logger)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	}
}

func (o *Operator) reconcileProxyService(mdbc *componentsv1alpha1.MariaDBCluster) error {
	return o.reconcileService(mdbc, mdbc.GetProxyServiceName(), mdbc.ProxyServiceTransform)
}

func (o *Operator) reconcileServerService(mdbc *componentsv1alpha1.MariaDBCluster) error {
	return o.reconcileService(mdbc, mdbc.GetServerServiceName(), mdbc.ServerServiceTransform)
}

func checkAndPatchService(current, expected *v1.Service, client clientcorev1.CoreV1Interface, logger *logrus.Entry) (bool, error) {
	// merge current values that should not trigger nor be included in patch
	mergeObjectMeta(&current.ObjectMeta, &expected.ObjectMeta)

	if !reflect.DeepEqual(expected.Spec.Ports, current.Spec.Ports) ||
		!reflect.DeepEqual(expected.Spec.Type, current.Spec.Type) ||
		!reflect.DeepEqual(expected.Annotations, current.Annotations) ||
		!reflect.DeepEqual(expected.Spec.Selector, current.Spec.Selector) {
		logger.Info("Spec differs between current and expected, updating")
		// TODO : Switch to Patch as Update fails due to immutable fields
		_, err := client.Services(expected.Namespace).Update(expected)
		return true, err
	}
	return false, nil
}
