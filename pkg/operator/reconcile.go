package operator

import (
	"reflect"

	"github.com/Sirupsen/logrus"
	componentsv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	componentsclient "github.com/goblain/mariadb-operator/pkg/generated/clientset/versioned/typed/components/v1alpha1"
	"github.com/goblain/mariadb-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (o *Operator) reconcileStatefulSet(mdbc *componentsv1alpha1.MariaDBCluster) error {
	logger := util.GetClusterLogger(mdbc).WithField("kind", "StatefulSet").WithField("action", "reconcile")
	logger.WithField("event", "started").Debug()
	defer logger.WithField("event", "finished").Debug()
	current, err := o.Client.AppsV1().StatefulSets(mdbc.Namespace).Get(mdbc.GetServerName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.WithField("event", "NotFound").Debug("not found in cluster")
			expected := &appsv1.StatefulSet{}
			mdbc.StatefulSetTransform(expected)
			_, err = o.Client.AppsV1().StatefulSets(mdbc.Namespace).Create(expected)
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
		mdbc.StatefulSetTransform(expected)
		checkAndPatchStatefulSet(current, expected, o.Client.Apps(), logger)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	}
}

func (o *Operator) reconcileServerConfigMap(mdbc *componentsv1alpha1.MariaDBCluster) error {
	logger := util.GetClusterLogger(mdbc).WithField("kind", "ConfigMap").WithField("action", "reconcile")
	logger.WithField("event", "started").Debug()
	defer logger.WithField("event", "finished").Debug()
	current, err := o.Client.CoreV1().ConfigMaps(mdbc.Namespace).Get(mdbc.GetServerConfigMapName(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.WithField("event", "NotFound").Debug("not found in cluster")
			expected := &v1.ConfigMap{}
			mdbc.ServerConfigMapTransform(expected)
			_, err = o.Client.CoreV1().ConfigMaps(mdbc.Namespace).Create(expected)
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
		mdbc.ServerConfigMapTransform(expected)
		checkAndPatchConfigMap(current, expected, o.Client.Core(), logger)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	}
}

func reconcile(clientInterface interface{}, mdbc *componentsv1alpha1.MariaDBCluster, expected interface{}) error {
	var expectedType string
	var err error
	var currentService, expectedService *v1.Service
	var currentStatefulSet, expectedStatefulSet *appsv1.StatefulSet
	var currentPVC, expectedPVC *v1.PersistentVolumeClaim
	expectedType = reflect.TypeOf(expected).Elem().Name()
	logger := util.GetClusterLogger(mdbc).WithField("object", expectedType).WithField("action", "reconcile")
	logger.Debug("Reconcile started")
	defer logger.Debug("Reconcile finished")

	switch expected.(type) {
	case *v1.PersistentVolumeClaim:
		expectedPVC = expected.(*v1.PersistentVolumeClaim)
		currentPVC, err = clientInterface.(clientcorev1.CoreV1Interface).PersistentVolumeClaims(expectedPVC.Namespace).Get(expectedPVC.Name, metav1.GetOptions{})
	case *v1.Service:
		expectedService = expected.(*v1.Service)
		currentService, err = clientInterface.(clientcorev1.CoreV1Interface).Services(expectedService.Namespace).Get(expectedService.Name, metav1.GetOptions{})
	case *appsv1.StatefulSet:
		expectedStatefulSet = expected.(*appsv1.StatefulSet)
		currentStatefulSet, err = clientInterface.(clientappsv1.AppsV1Interface).StatefulSets(expectedStatefulSet.Namespace).Get(expectedStatefulSet.Name, metav1.GetOptions{})
	}
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Debug("Object not found in cluster")
			switch expected.(type) {
			case *v1.PersistentVolumeClaim:
				_, err = clientInterface.(clientcorev1.CoreV1Interface).PersistentVolumeClaims(mdbc.Namespace).Create(expectedPVC)
			case *v1.Service:
				_, err = clientInterface.(clientcorev1.CoreV1Interface).Services(mdbc.Namespace).Create(expectedService)
			case *appsv1.StatefulSet:
				_, err = clientInterface.(clientappsv1.AppsV1Interface).StatefulSets(mdbc.Namespace).Create(expectedStatefulSet)
			}
			if err != nil {
				logger.Errorf("Creation failed with : %s", err.Error())
				return err
			} else {
				logger.Infof("Created successfully")
				return nil
			}
		} else {
			logger.Errorf("Error fetching object : %s", err.Error())
			return err
		}
	} else {
		logger.Debug("Comparing objects")
		switch expected.(type) {
		case *v1.PersistentVolumeClaim:
			checkAndPatch(currentPVC, expectedPVC, clientInterface, logger)
		case *v1.Service:
			checkAndPatch(currentService, expectedService, clientInterface, logger)
		case *appsv1.StatefulSet:
			checkAndPatch(currentStatefulSet, expectedStatefulSet, clientInterface, logger)
		}
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	}
}

func checkAndPatch(current, expected, clientInterface interface{}, logger *logrus.Entry) {
	var updated bool
	var err error
	switch expected.(type) {
	case *v1.ConfigMap:
		updated, err = checkAndPatchConfigMap(current.(*v1.ConfigMap), expected.(*v1.ConfigMap), clientInterface.(clientcorev1.CoreV1Interface), logger)
	case *v1.Service:
		updated, err = checkAndPatchService(current.(*v1.Service), expected.(*v1.Service), clientInterface.(clientcorev1.CoreV1Interface), logger)
	case *v1.PersistentVolumeClaim:
		updated, err = checkAndPatchPVC(current.(*v1.PersistentVolumeClaim), expected.(*v1.PersistentVolumeClaim), clientInterface.(clientcorev1.CoreV1Interface), logger)
	}
	if updated && err != nil {
		logger.Errorf("Object update failed : %s", err.Error())
	} else if updated {
		logger.Info("Object updated successfully")
	} else {
		logger.Debug("Object required no update")
	}
}

func patchGen(current, expected, kind interface{}) ([]byte, error) {
	return util.PatchGen(current, expected, kind)
}

func mergeAnnotations(base, top map[string]string) map[string]string {
	annotations := make(map[string]string)
	for key, val := range base {
		annotations[key] = val
	}
	for key, val := range top {
		annotations[key] = val
	}
	return annotations
}

// https://stackoverflow.com/questions/23555241/golang-reflection-how-to-get-zero-value-of-a-field-type
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).CanSet() {
				z = z && isZero(v.Field(i))
			}
		}
		return z
	case reflect.Ptr:
		return isZero(reflect.Indirect(v))
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	result := v.Interface() == z.Interface()

	return result
}

func mergeObjectMeta(current, expected *metav1.ObjectMeta) error {
	expected.SelfLink = current.SelfLink
	expected.UID = current.UID
	if isZero(reflect.ValueOf(expected.ResourceVersion)) {
		expected.ResourceVersion = current.ResourceVersion
	}
	if isZero(reflect.ValueOf(expected.CreationTimestamp)) {
		expected.CreationTimestamp = current.CreationTimestamp
	}
	expected.Generation = current.Generation
	expected.Annotations = mergeAnnotations(current.Annotations, expected.Annotations)
	return nil
}

func checkAndPatchPVC(current, expected *v1.PersistentVolumeClaim, client clientcorev1.CoreV1Interface, logger *logrus.Entry) (bool, error) {
	// merge current values that should not trigger nor be included in patch
	mergeObjectMeta(&current.ObjectMeta, &expected.ObjectMeta)
	expected.Status = current.Status
	expected.Spec.VolumeName = current.Spec.VolumeName

	if !reflect.DeepEqual(expected, current) {
		patchBytes, _ := patchGen(current, expected, v1.PersistentVolumeClaim{})
		_, err := client.PersistentVolumeClaims(expected.Namespace).Patch(expected.Name, types.StrategicMergePatchType, patchBytes)
		if err != nil {
			logger.Error(err.Error())
		}
		return true, nil
	}
	return false, nil
}

func checkAndPatchConfigMap(current, expected *v1.ConfigMap, client clientcorev1.CoreV1Interface, logger *logrus.Entry) (bool, error) {
	// merge current values that should not trigger nor be included in patch
	mergeObjectMeta(&current.ObjectMeta, &expected.ObjectMeta)

	if !reflect.DeepEqual(expected.Data, current.Data) {
		logger.Debug("Data differs between current and expected, updating")
		patchBytes, _ := patchGen(current, expected, v1.ConfigMap{})
		logger.Debugf(string(patchBytes))
		_, err := client.ConfigMaps(expected.Namespace).Patch(expected.Name, types.MergePatchType, patchBytes)
		return true, err
	}
	return false, nil
}

func checkAndPatchStatefulSet(current, expected *appsv1.StatefulSet, client clientappsv1.AppsV1Interface, logger *logrus.Entry) (bool, error) {
	if !reflect.DeepEqual(expected, current) {
		logger.WithField("event", "change").Info("changes detected")
		patchBytes, _ := patchGen(current, expected, appsv1.StatefulSet{})
		logger.Debugf(string(patchBytes))
		// TODO : error handling
		_, err := client.StatefulSets(expected.Namespace).Patch(expected.Name, types.StrategicMergePatchType, patchBytes)
		if err != nil {
			logger.Error(err.Error())
		}
		return true, nil
	} else {
		logger.WithField("event", "nochange").Info("no changes")
	}
	return false, nil
}

func checkAndPatchMariaDBCluster(current, expected *componentsv1alpha1.MariaDBCluster, client componentsclient.ComponentsV1alpha1Interface, logger *logrus.Entry) (bool, error) {
	return util.CheckAndPatchMariaDBCluster(current, expected, client, logger)
}
