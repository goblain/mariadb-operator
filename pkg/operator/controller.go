package operator

import (
	"time"

	"github.com/Sirupsen/logrus"

	api "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	listers "github.com/goblain/mariadb-operator/pkg/generated/listers/components/v1alpha1"
	appslisters "k8s.io/client-go/listers/apps/v1"
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
	deploymentsLister     appslisters.DeploymentLister
	deploymentsSynced     cache.InformerSynced
	mariadbclustersLister listers.MariaDBClusterLister
	mariadbclustersSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workq workqueue.RateLimitingInterface
}

func NewController(op *Operator) *Controller {
	// func NewController(op *Operator, kubeInformerFactory informers.SharedInformerFactory, componentsInformerFactory mariadbinformers.SharedInformerFactory) *Controller {
	ctrl := &Controller{
		operator: op,
	}
	// deploymentInformer := kubeInformerFactory.Apps().V1().Deployments()
	// mariaInformer := componentsInformerFactory.Components().V1alpha1().MariaDBClusters()
	return ctrl
}

func (c *Controller) Run() {
	for {
		time.Sleep(time.Second * 5)
		logrus.Infof("[" + api.GroupName + "/" + api.Version + "] Tick")
	}
}
