package operator

import (
	"os"
	"os/signal"
	"time"

	"github.com/Sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	defaultKubeAPIRequestTimeout = 30 * time.Second
	//	TODO: remove these temporary assocs with coded
	name      = "mariadb-operator"
	namespace = "kube-system"
	id        = "123213123"
)

type Operator struct {
	Name                string
	ClientConfig        *rest.Config
	Client              *kubernetes.Clientset
	ApiExtensionsClient *apiextensionsclientset.Clientset
}

func NewOperator() *Operator {
	op := &Operator{
		Name: "mariadb-operator",
	}
	return op
}

func (op *Operator) Start() {
	var err error
	logrus.Info(op.Name)
	op.ClientConfig, err = InClusterConfig()
	if err != nil {
		op.ClientConfig, err = OutOfClusterConfig()
		if err != nil {
			panic(err)
		}
	}
	op.ClientConfig.Timeout = defaultKubeAPIRequestTimeout

	op.Client = kubernetes.NewForConfigOrDie(op.ClientConfig)
	op.ApiExtensionsClient = apiextensionsclientset.NewForConfigOrDie(op.ClientConfig)

	// Take care of termination by signal
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		logrus.Infof("received signal: %v", <-c)
		os.Exit(1)
	}()

	lock, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		namespace,
		op.Name,
		op.Client.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: createRecorder(op.Client, name, namespace),
		})
	if err != nil {
		panic("no configuration for kubernetes client")
	}

	leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: op.run,
			OnStoppedLeading: func() {
				logrus.Fatalf("leader election lost")
			},
		},
	})

	panic("wtf")
}

// Register all supported CRDs and launch all supported controller versions
func (op *Operator) run(stop <-chan struct{}) {
	// v1alpha1api :=
	// Register all supported CRDs
	op.EnsureSupportedCRDs()
	// Get informerFactories
	// kubeInformerFactory := informers.NewSharedInformerFactory(op.Client, time.Second*30)
	// componentInformerFactory := componentinformers.NewSharedInformerFactory(op.Client, time.Second*30)
	// Launch all supported controller versions
	// v1alpha1ctrl := NewController(op, kubeInformerFactory)
	v1alpha1ctrl := NewController(op)
	go v1alpha1ctrl.Run()

}

func InClusterConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func OutOfClusterConfig() (*rest.Config, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", "/home/goblin/.kube/kubectl-config")
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func createRecorder(kcs *kubernetes.Clientset, name, namespace string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logrus.Infof)
	eventBroadcaster.StartRecordingToSink(&v1.EventSinkImpl{Interface: kcs.CoreV1().Events(namespace)})
	return eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: name})
}

// func getMyPodServiceAccount(kubecli kubernetes.Interface) (string, error) {
// 	var sa string
// 	err := retryutil.Retry(5*time.Second, 100, func() (bool, error) {
// 		pod, err := kubecli.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
// 		if err != nil {
// 			logrus.Errorf("fail to get operator pod (%s): %v", name, err)
// 			return false, nil
// 		}
// 		sa = pod.Spec.ServiceAccountName
// 		return true, nil
// 	})
// 	return sa, err
// }
