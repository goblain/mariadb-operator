package sidecar

import (
	"flag"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	componentsclientset "github.com/goblain/mariadb-operator/pkg/generated/clientset/versioned"
	componentsinformers "github.com/goblain/mariadb-operator/pkg/generated/informers/externalversions"

	"github.com/Sirupsen/logrus"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/staging/src/k8s.io/sample-controller/pkg/signals"
)

const (
	defaultKubeAPIRequestTimeout = 30 * time.Second
	//	TODO: remove these temporary assocs with coded
	name      = "mariadb-operator"
	namespace = "kube-system"
	id        = "123213123"
)

type Operator struct {
	ClientConfig        *rest.Config
	Client              *kubernetes.Clientset
	ComponentsClient    *componentsclientset.Clientset
	ApiExtensionsClient *apiextensionsclientset.Clientset
}

func NewOperator() *Operator {
	op := &Operator{}
	return op
}

func (op *Operator) Run() {
	var err error

	stopCh := signals.SetupSignalHandler()

	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug("Debug logging enabled")
	op.ClientConfig, err = InClusterConfig()
	if err != nil {
		panic(err)
	}
	op.ClientConfig.Timeout = defaultKubeAPIRequestTimeout

	op.Client = kubernetes.NewForConfigOrDie(op.ClientConfig)
	op.ComponentsClient = componentsclientset.NewForConfigOrDie(op.ClientConfig)
	op.ApiExtensionsClient = apiextensionsclientset.NewForConfigOrDie(op.ClientConfig)

	// Take care of termination by signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT)

	kubeInformerFactory := informers.NewSharedInformerFactory(op.Client, time.Second*30)
	componentInformerFactory := componentsinformers.NewSharedInformerFactory(op.ComponentsClient, time.Second*30)

	go kubeInformerFactory.Start(stopCh)
	go componentInformerFactory.Start(stopCh)
	go func() {
		logrus.Infof("received signal: %v, exiting", <-stopCh)
		os.Exit(1)
	}()
	go func() {
		for true {
			time.Sleep(time.Second * 5)
			logrus.Debug("tick")
		}
	}()

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// w.WriteHeader(http.StatusInternalServerError)
		// w.Write([]byte("NotReady"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	go http.ListenAndServe(":8080", nil)

	mysql := exec.Command("/docker-entrypoint.sh")
	mysql.Stdout = os.Stdout
	mysql.Stderr = os.Stderr
	mysql.Run()
}

func InClusterConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
