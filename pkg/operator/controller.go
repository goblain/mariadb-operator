package operator

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	componentsv1alpha1 "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	componentinformers "github.com/goblain/mariadb-operator/pkg/generated/informers/externalversions"
	listers "github.com/goblain/mariadb-operator/pkg/generated/listers/components/v1alpha1"
	"github.com/goblain/mariadb-operator/pkg/util"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type cluster struct {
	name string
}

type Controller struct {
	operator *Operator
	clusters map[string]*cluster

	// Listers and informers for objects on changes to which controller will react
	configmapLister       corelisters.ConfigMapLister
	configmapSynced       cache.InformerSynced
	statefulsetLister     appslisters.StatefulSetLister
	statefulsetSynced     cache.InformerSynced
	mariadbclustersLister listers.MariaDBClusterLister
	mariadbclustersSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	stopChan  chan struct{}
}

func NewController(op *Operator, kubeInformerFactory informers.SharedInformerFactory, componentsInformerFactory componentinformers.SharedInformerFactory) *Controller {
	statefulsetInformer := kubeInformerFactory.Apps().V1().StatefulSets()
	configmapInformer := kubeInformerFactory.Core().V1().ConfigMaps()
	mariaInformer := componentsInformerFactory.Components().V1alpha1().MariaDBClusters()
	c := &Controller{
		operator:              op,
		configmapLister:       configmapInformer.Lister(),
		configmapSynced:       configmapInformer.Informer().HasSynced,
		statefulsetLister:     statefulsetInformer.Lister(),
		statefulsetSynced:     statefulsetInformer.Informer().HasSynced,
		mariadbclustersLister: mariaInformer.Lister(),
		mariadbclustersSynced: mariaInformer.Informer().HasSynced,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "MariaDBClusters"),
	}

	logrus.Info("Adding event handlers for MariaDBClusters informer")
	mariaInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.MariaDBClusterAddEventHandler,
			UpdateFunc: c.MariaDBClusterUpdateEventHandler,
			DeleteFunc: c.MariaDBClusterDeleteEventHandler,
		})

	logrus.Info("Adding event handlers for StatefulSet informer")
	statefulsetInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.StatefulSetAddEventHandler,
			UpdateFunc: c.StatefulSetUpdateEventHandler,
			DeleteFunc: c.StatefulSetDeleteEventHandler,
		})

	return c
}

func (c *Controller) WaitForCacheSync() {
	if ok := cache.WaitForCacheSync(c.stopChan, c.statefulsetSynced, c.configmapSynced, c.mariadbclustersSynced); !ok {
		panic("Failed to sync cache")
	}
}

func (c *Controller) MariaDBClusterEnqueue(obj interface{}) error {
	mdb := obj.(*componentsv1alpha1.MariaDBCluster)
	logrus.WithFields(logrus.Fields{"cluster": mdb.Namespace + "/" + mdb.Name}).Debugf("Adding MariaDBCluster to workqueue")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	c.workqueue.AddRateLimited(key)
	return err
}

func (c *Controller) processNextFromQueue() error {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return nil
	}
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		return nil
	}(obj)
	if err != nil {
		return fmt.Errorf("Ooops somthing failed processing a work item")
	}
	return nil
}

func (c *Controller) syncHandler(key string) error {
	logrus.Debugf("Controller.syncHandler called with %s", key)
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Cluster resource with this namespace/name
	cluster, err := c.mariadbclustersLister.MariaDBClusters(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("Cluster '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	c.reconcileCluster(cluster)
	return nil
}

func (c *Controller) noConflictingResources(cluster *componentsv1alpha1.MariaDBCluster) bool {
	var resources string
	var err error
	_, err = c.statefulsetLister.StatefulSets(cluster.Namespace).Get(cluster.Name)
	if !errors.IsNotFound(err) {
		resources = resources + " StatefulSet"
	}
	_, err = c.configmapLister.ConfigMaps(cluster.Namespace).Get(cluster.Name)
	if !errors.IsNotFound(err) {
		resources = resources + " ConfigMap"
	}

	if resources == "" {
		return true
	} else {
		logrus.Debugf("Found conflicting resources : %s", resources)
		return false
	}
}

func (c *Controller) reconcileCluster(cluster *componentsv1alpha1.MariaDBCluster) {
	c.reconcileMariaDBCluster(cluster)
	pvc := cluster.GetSnapshotPVC()
	reconcile(c.operator.Client.CoreV1(), cluster, pvc)
	c.operator.reconcileServerServiceAccount(cluster)
	c.operator.reconcileServerRole(cluster)
	c.operator.reconcileServerRoleBinding(cluster)
	c.operator.reconcileServerConfigMap(cluster)
	c.operator.reconcileStatefulSet(cluster)
	c.operator.reconcileServerService(cluster)
	c.operator.reconcileProxyService(cluster)
}

type Patch []PatchSpec

type PatchSpec struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (c *Controller) syncWorker() {
	for {
		c.processNextFromQueue()
	}
}

func (c *Controller) Run() {
	c.WaitForCacheSync()
	go c.syncWorker()
}

// check if any criteria for state transition are met
func (c *Controller) MariaDBClusterTransform(mdbc *componentsv1alpha1.MariaDBCluster) error {
	logger := logrus.WithField("kind", "MariaDBCluster")
	// Start cluster bootstrap if phase is empty
	switch mdbc.Status.Phase {

	case "":
		logger.Debug("Detected no Phase, transitioning to BootstrapFirst")
		mdbc.Status.Phase = componentsv1alpha1.PhaseBootstrapFirst

	// First phase of bootstrap, starting the cluster with --wsrep-cluster-new
	case componentsv1alpha1.PhaseBootstrapFirst:
		logger.Debug("Detected BootstrapFirst Phase, checking transitions")
		sset, err := c.statefulsetLister.StatefulSets(mdbc.Namespace).Get(mdbc.GetServerName())
		if err == nil {
			if mdbc.Spec.Replicas > 1 &&
				isStatefulSetReady(sset) {
				logger.WithField("event", "phaseTransition").Info("Transitioning to BootstrapFirstRestart phase")
				mdbc.Status.Phase = componentsv1alpha1.PhaseBootstrapFirstRestart
				mdbc.Status.StatefulSetObservedGeneration = sset.Status.ObservedGeneration
			}
		}

	// Restart loosing --wsrep-cluste-new so we do not wipe cluster IP
	case componentsv1alpha1.PhaseBootstrapFirstRestart:
		logger.Debug("Detected BootstrapFirstRestart Phase, checking transitions")
		sset, _ := c.statefulsetLister.StatefulSets(mdbc.Namespace).Get(mdbc.GetServerName())
		if mdbc.Spec.Replicas > 1 &&
			isStatefulSetUpdated(mdbc, sset) &&
			isStatefulSetReady(sset) {
			logger.WithField("event", "phaseTransition").Info("Transitioning to BootstrapSecond phase")
			mdbc.Status.Phase = componentsv1alpha1.PhaseBootstrapSecond
			mdbc.Status.StatefulSetObservedGeneration = sset.Status.ObservedGeneration
		}

	// Bootstrap second node of galera cluster
	case componentsv1alpha1.PhaseBootstrapSecond:
		logger.Debug("Detected BootstrapSecond Phase, checking transitions")
		sset, _ := c.statefulsetLister.StatefulSets(mdbc.Namespace).Get(mdbc.GetServerName())
		if mdbc.Spec.Replicas > 2 &&
			isStatefulSetUpdated(mdbc, sset) &&
			isStatefulSetReady(sset) {
			logger.WithField("event", "phaseTransition").Info("Transitioning to BootstrapSecond phase")
			mdbc.Status.Phase = componentsv1alpha1.PhaseBootstrapThird
			mdbc.Status.StatefulSetObservedGeneration = sset.Status.ObservedGeneration
		}

	// Bootstrap third node of galera cluster
	case componentsv1alpha1.PhaseBootstrapThird:
		logger.Debug("Detected BootstrapThird Phase, checking transitions")
		sset, _ := c.statefulsetLister.StatefulSets(mdbc.Namespace).Get(mdbc.GetServerName())
		if mdbc.Spec.Replicas > 2 &&
			isStatefulSetUpdated(mdbc, sset) &&
			isStatefulSetReady(sset) {
			logger.WithField("event", "phaseTransition").Info("Transitioning to BootstrapSecond phase")
			mdbc.Status.Phase = componentsv1alpha1.PhaseOperational
			mdbc.Status.StatefulSetObservedGeneration = sset.Status.ObservedGeneration
		}
		// Detect unhealthy state
	case componentsv1alpha1.PhaseOperational:
		logger.Debug("Detected BootstrapThird Phase, checking transitions")
		sset, _ := c.statefulsetLister.StatefulSets(mdbc.Namespace).Get(mdbc.GetServerName())
		if sset.Status.ReadyReplicas == 0 {
			mdbc.Status.Phase = componentsv1alpha1.PhaseRecovery
		}

	case componentsv1alpha1.PhaseRecovery:
		logger.Debug("Detected Recovery Phase, checking transitions")
		if mdbc.Spec.Replicas > 1 {
			// if all the pods reported their state
			reported := int32(len(mdbc.Status.StatefulSetPodConditions))
			if reported == mdbc.Spec.Replicas {
				allAreEqual := true
				for k, v := range mdbc.Status.StatefulSetPodConditions[1:] {
					if v.GRAState.SeqNo != mdbc.Status.StatefulSetPodConditions[k].GRAState.SeqNo {
						allAreEqual = false
					}
				}
				if allAreEqual {
					mdbc.Status.Phase = componentsv1alpha1.PhaseRecoveryReleaseAll
				}
			}
		} else {
			mdbc.Status.Phase = componentsv1alpha1.PhaseRecoveryReleaseAll
		}

	case componentsv1alpha1.PhaseRecoveryReleaseAll:
		logger.Debug("Detected RecoveryReleaseAll Phase, checking transitions")
		sset, _ := c.statefulsetLister.StatefulSets(mdbc.Namespace).Get(mdbc.GetServerName())
		if isStatefulSetReady(sset) {
			mdbc.Status.Phase = componentsv1alpha1.PhaseOperational
			mdbc.Status.StatefulSetPodConditions = nil
		}

	}
	return nil
}

func isStatefulSetUpdated(mdbc *componentsv1alpha1.MariaDBCluster, sset *apps.StatefulSet) bool {
	return sset.Status.ObservedGeneration > mdbc.Status.StatefulSetObservedGeneration
}

func isStatefulSetReady(sset *apps.StatefulSet) bool {
	return *sset.Spec.Replicas == sset.Status.CurrentReplicas &&
		*sset.Spec.Replicas == sset.Status.Replicas &&
		*sset.Spec.Replicas == sset.Status.ReadyReplicas &&
		sset.Status.CurrentRevision == sset.Status.UpdateRevision
}

func (c *Controller) reconcileMariaDBCluster(mdbc *componentsv1alpha1.MariaDBCluster) error {
	logger := util.GetClusterLogger(mdbc).WithField("kind", "MariaDBCluster").WithField("action", "reconcile")
	logger.WithField("event", "started").Debug()
	defer logger.WithField("event", "finished").Debug()
	original := mdbc.DeepCopy()
	c.MariaDBClusterTransform(mdbc)
	checkAndPatchMariaDBCluster(original, mdbc, c.operator.ComponentsClient.Components(), logger)
	return nil
}
