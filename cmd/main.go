package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	name                         = "mariadb-operator"
	namespace                    = "kube-system"
	id                           = "123213123"
	defaultKubeAPIRequestTimeout = 30 * time.Second
)

func InClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
	// 	addrs, err := net.LookupHost("kubernetes.default.svc")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	// }
	// if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
	// 	os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	// }
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// Set a reasonable default request timeout
	cfg.Timeout = defaultKubeAPIRequestTimeout
	return cfg, nil
}

func OutOfClusterConfig() (*rest.Config, error) {

	// cfg := clientcmd.DirectClientConfig{}.NewNonInteractiveClientConfig()
	// cfg := clientcmd.DirectClientConfig{}.NewNonInteractiveClientConfig()
	// cfg := &clientcmd.ClientConfigGetter{}
	// cfg := clientcmd.GetConfigFromFileOrDie("~/.kube/kube-config")
	// cfg.CurrentContext = "pentex"
	cfg, err := clientcmd.BuildConfigFromFlags("", "/home/goblin/.kube/kubectl-config")
	if err != nil {
		return nil, err
	}
	// Set a reasonable default request timeout
	cfg.Timeout = defaultKubeAPIRequestTimeout
	return cfg, nil
}

func main() {
	cfg, err := InClusterConfig()
	if err != nil {
		cfg, err = OutOfClusterConfig()
		if err != nil {
			panic(err)
		}
	}

	// mdbcs := mdbclientset.NewForConfigOrDie(cfg)
	kcs := kubernetes.NewForConfigOrDie(cfg)

	// Take care of termination by signal
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		logrus.Infof("received signal: %v", <-c)
		os.Exit(1)
	}()

	lock, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		namespace,
		"mariadb-operator",
		kcs.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: createRecorder(kcs, name, namespace),
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
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				logrus.Fatalf("leader election lost")
			},
		},
	})

	panic("wtf")
}

func createRecorder(kcs *kubernetes.Clientset, name, namespace string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logrus.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kcs.CoreV1().Events(namespace)})
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: name})
}

func run(stop <-chan struct{}) {

}
