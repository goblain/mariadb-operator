package operator

import (
	"reflect"

	"github.com/Sirupsen/logrus"
	componentsv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	apps "k8s.io/api/apps/v1"
)

/*
 *  MariaDBCluster Event Handlers
 */

func (c *Controller) MariaDBClusterAddEventHandler(obj interface{}) {
	mdb := obj.(*componentsv1alpha1.MariaDBCluster)
	logrus.Infof("MariaDBCluster Add Event logged for %s/%s", mdb.Namespace, mdb.Name)
	c.MariaDBClusterEnqueue(obj)
}

func (c *Controller) MariaDBClusterUpdateEventHandler(oldobj, newobj interface{}) {
	oldmdb := oldobj.(*componentsv1alpha1.MariaDBCluster)
	newmdb := newobj.(*componentsv1alpha1.MariaDBCluster)
	logger := logrus.WithFields(logrus.Fields{"cluster": oldmdb.Namespace + "/" + oldmdb.Name})
	logger.Debug("MariaDBCluster Update Event recieved")

	if !reflect.DeepEqual(newmdb.Spec, oldmdb.Spec) || !reflect.DeepEqual(newmdb.Status, oldmdb.Status) {
		logger.Debug("MariaDBCluster change detected, queue for reconcile")
		c.MariaDBClusterEnqueue(newobj)
	} else {
		logger.Debug("MariaDBCluster has not changed")
	}
}

func (c *Controller) MariaDBClusterDeleteEventHandler(obj interface{}) {
	mdb := obj.(*componentsv1alpha1.MariaDBCluster)
	logger := logrus.WithFields(logrus.Fields{"cluster": mdb.Namespace + "/" + mdb.Name})
	logger.Infof("MariaDBCluster Delete Event recieved")
	c.MariaDBClusterEnqueue(obj)
}

/*
 *  StatefulSet Handlers
 */

func (c *Controller) StatefulSetAddEventHandler(obj interface{}) {
	sset := obj.(*apps.StatefulSet)
	logrus.Infof("StatefulSet Add Event logged for %s/%s", sset.Namespace, sset.Name)
	if len(sset.Labels[componentsv1alpha1.MariaDBClusterNameLabel]) > 0 {
		c.workqueue.AddRateLimited(sset.Namespace + "/" + sset.Labels[componentsv1alpha1.MariaDBClusterNameLabel])
	}
}

func (c *Controller) StatefulSetUpdateEventHandler(oldobj, newobj interface{}) {
	oldsset := oldobj.(*apps.StatefulSet)
	newsset := newobj.(*apps.StatefulSet)
	logrus.Infof("StatefulSet Update Event logged for %s/%s", oldsset.Namespace, oldsset.Name)
	if len(newsset.Labels[componentsv1alpha1.MariaDBClusterNameLabel]) > 0 && !reflect.DeepEqual(oldsset, newsset) {
		c.workqueue.AddRateLimited(newsset.Namespace + "/" + newsset.Labels[componentsv1alpha1.MariaDBClusterNameLabel])
	}
}

func (c *Controller) StatefulSetDeleteEventHandler(obj interface{}) {
	sset := obj.(*apps.StatefulSet)
	logrus.Infof("StatefulSet Delete Event logged for %s/%s", sset.Namespace, sset.Name)
	if len(sset.Labels[componentsv1alpha1.MariaDBClusterNameLabel]) > 0 {
		c.workqueue.AddRateLimited(sset.Namespace + "/" + sset.Labels[componentsv1alpha1.MariaDBClusterNameLabel])
	}
}
