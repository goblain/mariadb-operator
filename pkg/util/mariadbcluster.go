package util

import (
	"reflect"

	"github.com/Sirupsen/logrus"
	componentsv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	componentsclient "github.com/goblain/mariadb-operator/pkg/generated/clientset/versioned/typed/components/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

func CheckAndPatchMariaDBCluster(current, expected *componentsv1alpha1.MariaDBCluster, client componentsclient.ComponentsV1alpha1Interface, logger *logrus.Entry) (bool, error) {
	if !reflect.DeepEqual(expected, current) {
		patchBytes, _ := PatchGen(current, expected, componentsv1alpha1.MariaDBCluster{})
		logger.Debugf(string(patchBytes))
		// TODO : error handling
		_, err := client.MariaDBClusters(expected.Namespace).Patch(expected.Name, types.MergePatchType, patchBytes)
		if err != nil {
			logger.Error(err.Error())
		}
		return true, nil
	}
	return false, nil
}

func GetClusterLogger(mdbc *componentsv1alpha1.MariaDBCluster) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{"cluster": mdbc.Namespace + "/" + mdbc.Name})
}
