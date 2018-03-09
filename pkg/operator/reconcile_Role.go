package operator

import (
	"reflect"

	"github.com/Sirupsen/logrus"
	componentsv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	"github.com/goblain/mariadb-operator/pkg/util"
	rbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientrbac "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

func (o *Operator) reconcileRole(mdbc *componentsv1alpha1.MariaDBCluster, name string, transformer func(*rbac.Role) error) error {
	logger := util.GetClusterLogger(mdbc).WithField("kind", "Role").WithField("action", "reconcile").WithField("name", name)
	logger.WithField("event", "started").Debug()
	defer logger.WithField("event", "finished").Debug()
	current, err := o.Client.RbacV1().Roles(mdbc.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.WithField("event", "NotFound").Debug("not found in cluster")
			expected := &rbac.Role{}
			transformer(expected)
			_, err = o.Client.RbacV1().Roles(mdbc.Namespace).Create(expected)
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
		checkAndPatchRole(current, expected, o.Client.RbacV1(), logger)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	}
}

func (o *Operator) reconcileServerRole(mdbc *componentsv1alpha1.MariaDBCluster) error {
	return o.reconcileRole(mdbc, mdbc.GetServerName(), mdbc.ServerRoleTransform)
}

func checkAndPatchRole(current, expected *rbac.Role, client clientrbac.RbacV1Interface, logger *logrus.Entry) (bool, error) {
	if !reflect.DeepEqual(expected, current) {
		logger.WithField("event", "change").Info("changes detected")
		patchBytes, _ := patchGen(current, expected, rbac.Role{})
		logger.Debugf(string(patchBytes))
		_, err := client.Roles(expected.Namespace).Patch(expected.Name, types.StrategicMergePatchType, patchBytes)
		if err != nil {
			logger.Error(err.Error())
		}
		return true, nil
	} else {
		logger.WithField("event", "nochange").Info("no changes")
	}
	return false, nil
}
