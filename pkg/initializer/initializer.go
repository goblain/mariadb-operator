package initializer

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	components "github.com/goblain/mariadb-operator/pkg/apis/components/v1alpha1"
	componentsclientset "github.com/goblain/mariadb-operator/pkg/generated/clientset/versioned"
	"github.com/goblain/mariadb-operator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	defaultKubeAPIRequestTimeout = 30 * time.Second
	//	TODO: remove these temporary assocs with coded
	name      = "mariadb-operator"
	namespace = "kube-system"
	id        = "123213123"
)

type Initializer struct {
	Hostname     string
	clientConfig *rest.Config
	// client           *kubernetes.Clientset
	componentsClient *componentsclientset.Clientset
	// apiExtensionsClient *apiextensionsclientset.Clientset
	logger    *logrus.Entry
	name      string
	namespace string
}

func (i *Initializer) Run() {

	// Take care of termination by signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		logrus.Infof("received signal: %v, exiting", <-c)
		os.Exit(1)
	}()

	var err error

	i.name = "rocket"
	i.namespace = "default"
	if i.Hostname, err = os.Hostname(); err != nil {
		panic(err.Error())
	}

	logrus.SetLevel(logrus.DebugLevel)
	i.logger = logrus.WithField("namespace", namespace).WithField("name", name)
	i.logger.Debug("Debug logging enabled")

	i.clientConfig, err = InClusterConfig()
	if err != nil {
		panic(err)
	}
	i.clientConfig.Timeout = defaultKubeAPIRequestTimeout
	i.componentsClient = componentsclientset.NewForConfigOrDie(i.clientConfig)

	if i.getMariaDBCluster().Status.Phase == components.PhaseRecovery {
		for true {
			i.logger.Debug("Recovery phase detected, reporting my condition to MariaDBCluster object")
			i.reportToMariaDBCluster()
			time.Sleep(time.Second * 5)
			if i.getMariaDBCluster().Status.Phase == components.PhaseRecoveryReleaseAll {
				os.Exit(0)
			}
		}
	}
	i.logger.Info("Finishing init in 5 sec")
	// Be gracefull, don't rush ahead
	time.Sleep(time.Second * 5)
}

func (i *Initializer) getMariaDBCluster() *components.MariaDBCluster {
	mdbc, err := i.componentsClient.Components().MariaDBClusters(i.namespace).Get(i.name, metav1.GetOptions{})
	if err != nil {
		i.logger.Error("where is my cluster?!? : " + err.Error())
		panic("where is my cluster?!? : " + err.Error())
	}
	return mdbc
}

func (i *Initializer) reportToMariaDBCluster() {
	current := i.getMariaDBCluster()
	expected := current.DeepCopy()
	// /var/lib/mysql/grastate.dat
	// # GALERA saved state
	// version: 2.1
	// uuid:    a0044fb3-3cb6-11e8-9641-321932c64bd8
	// seqno:   6
	// safe_to_bootstrap: 0
	stateString, err := ioutil.ReadFile("/var/lib/mysql/grastate.dat")
	if err != nil {
		panic("missing grastate.dat : " + err.Error())
	}

	var safeToBootstrap int
	var seqno int64
	var uuid, version string
	var re *regexp.Regexp
	var result []string

	re = regexp.MustCompile(`version:\s*([0-9\.]*)`)
	result = re.FindStringSubmatch(string(stateString))
	if len(result) > 1 {
		version = result[1]
	} else {
		panic("Version missing")
	}

	re = regexp.MustCompile(`uuid:\s*([A-Za-z0-9-]*)`)
	result = re.FindStringSubmatch(string(stateString))
	if len(result) > 1 {
		uuid = result[1]
	} else {
		panic("UUID missing")
	}

	re = regexp.MustCompile(`seqno:\s*(\d+)`)
	result = re.FindStringSubmatch(string(stateString))
	if len(result) > 1 {
		seqno, err = strconv.ParseInt(result[1], 10, 64)
	} else {
		panic("SeqNo missing : " + err.Error())
	}

	re = regexp.MustCompile(`safe_to_bootstrap:\s*(\d)`)
	result = re.FindStringSubmatch(string(stateString))
	if len(result) > 1 {
		safeToBootstrap, err = strconv.Atoi(result[1])
	} else {
		logrus.Warn("safe_to_bootstrap missing : " + err.Error())
	}

	podCondition := components.PodCondition{
		Hostname: i.Hostname,
		Reported: metav1.Now(),
		GRAState: components.PodConditionGRAState{Version: version, UUID: uuid, SeqNo: seqno, SafeToBootstrap: safeToBootstrap},
	}
	match := false
	for k, v := range expected.Status.StatefulSetPodConditions {
		if v.Hostname == podCondition.Hostname {
			match = true
			if !reflect.DeepEqual(v.GRAState, podCondition.GRAState) {
				logrus.Debug("Entry found and status needs update")
				expected.Status.StatefulSetPodConditions[k].Hostname = podCondition.Hostname
				expected.Status.StatefulSetPodConditions[k].Reported = podCondition.Reported
				expected.Status.StatefulSetPodConditions[k].GRAState = podCondition.GRAState
			} else {
				logrus.Debug("Entry found and no status update required")
			}
			break
		}
	}
	if !match {
		expected.Status.StatefulSetPodConditions = append(expected.Status.StatefulSetPodConditions, podCondition)
	}
	js, _ := json.Marshal(current)
	i.logger.Debug(string(js))
	js, _ = json.Marshal(expected)
	i.logger.Debug(string(js))
	util.CheckAndPatchMariaDBCluster(current, expected, i.componentsClient.Components(), i.logger)
}

func InClusterConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
